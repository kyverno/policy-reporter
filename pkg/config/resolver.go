package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"time"

	goredis "github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	mail "github.com/xhit/go-simple-mail/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned"
	wgpolicyk8sv1alpha2 "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/leaderelection"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/report/result"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

// Resolver manages dependencies
type Resolver struct {
	config             *Config
	k8sConfig          *rest.Config
	publisher          report.EventPublisher
	policyStore        *database.Store
	database           *bun.DB
	policyReportClient report.PolicyReportClient
	leaderElector      *leaderelection.Client
	resultCache        cache.Cache
	targetClients      []target.Client
	targetsCreated     bool
	logger             *zap.Logger
	resultListener     *listener.ResultListener
}

// APIServer resolver method
func (r *Resolver) APIServer(ctx context.Context, synced func() bool) api.Server {
	var logger *zap.Logger
	if r.config.API.Logging {
		logger, _ = r.Logger()
	}

	if r.config.API.BasicAuth.SecretRef != "" {
		values, err := r.SecretClient().Get(ctx, r.config.API.BasicAuth.SecretRef)
		if err != nil {
			zap.L().Error("failed to load basic auth secret", zap.Error(err))
		}

		if values.Username != "" {
			r.config.API.BasicAuth.Username = values.Username
		}
		if values.Password != "" {
			r.config.API.BasicAuth.Password = values.Password
		}
	}

	var auth *api.BasicAuth
	if r.config.API.BasicAuth.Username != "" && r.config.API.BasicAuth.Password != "" {
		auth = &api.BasicAuth{
			Username: r.config.API.BasicAuth.Username,
			Password: r.config.API.BasicAuth.Password,
		}

		zap.L().Info("API BasicAuth enabled")
	}

	return api.NewServer(
		r.TargetClients(),
		r.config.API.Port,
		logger,
		auth,
		synced,
	)
}

// Database resolver method
func (r *Resolver) Database() *bun.DB {
	if r.database != nil {
		return r.database
	}

	factory := r.DatabaseFactory()

	switch r.config.Database.Type {
	case database.MySQL:
		if r.database = factory.NewMySQL(r.config.Database); r.database != nil {
			zap.L().Info("mysql connection created")
			return r.database
		}
	case database.MariaDB:
		if r.database = factory.NewMySQL(r.config.Database); r.database != nil {
			zap.L().Info("mariadb connection created")
			return r.database
		}
	case database.PostgreSQL:
		if r.database = factory.NewPostgres(r.config.Database); r.database != nil {
			zap.L().Info("postgres connection created")
			return r.database
		}
	}

	zap.L().Info("sqlite connection created")
	r.database = factory.NewSQLite(r.config.DBFile)
	return r.database
}

// PolicyReportStore resolver method
func (r *Resolver) PolicyReportStore(db *bun.DB) (*database.Store, error) {
	if r.policyStore != nil {
		return r.policyStore, nil
	}

	s, err := database.NewStore(db, r.config.Version)
	r.policyStore = s

	return r.policyStore, err
}

// LeaderElectionClient resolver method
func (r *Resolver) LeaderElectionClient() (*leaderelection.Client, error) {
	if r.leaderElector != nil {
		return r.leaderElector, nil
	}

	clientset, err := k8s.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	r.leaderElector = leaderelection.New(
		clientset.CoordinationV1(),
		r.config.LeaderElection.LockName,
		r.config.LeaderElection.Namespace,
		r.config.LeaderElection.PodName,
		time.Duration(r.config.LeaderElection.LeaseDuration)*time.Second,
		time.Duration(r.config.LeaderElection.RenewDeadline)*time.Second,
		time.Duration(r.config.LeaderElection.RetryPeriod)*time.Second,
		r.config.LeaderElection.ReleaseOnCancel,
	)

	return r.leaderElector, nil
}

// EventPublisher resolver method
func (r *Resolver) EventPublisher() report.EventPublisher {
	if r.publisher != nil {
		return r.publisher
	}

	s := report.NewEventPublisher()
	r.publisher = s

	return r.publisher
}

func (r *Resolver) CustomIDGenerators() map[string]result.IDGenerator {
	generators := make(map[string]result.IDGenerator)
	for s, c := range r.config.SourceConfig {
		if !c.Enabled || len(c.Fields) == 0 {
			continue
		}

		generators[strings.ToLower(s)] = result.NewIDGenerator(c.Fields)
	}

	return generators
}

// EventPublisher resolver method
func (r *Resolver) Queue() (*kubernetes.Queue, error) {
	client, err := r.CRDClient()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewQueue(
		kubernetes.NewDebouncer(1*time.Minute, r.EventPublisher()),
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "report-queue"),
		client,
		result.NewReconditioner(r.CustomIDGenerators()),
	), nil
}

// RegisterNewResultsListener resolver method
func (r *Resolver) RegisterNewResultsListener() {
	targets := r.TargetClients()
	if len(targets) == 0 {
		return
	}

	newResultListener := listener.NewResultListener(r.SkipExistingOnStartup(), r.ResultCache(), time.Now())
	r.resultListener = newResultListener
	r.EventPublisher().RegisterListener(listener.NewResults, newResultListener.Listen)

	r.EventPublisher().RegisterPostListener(listener.CleanUpListener, listener.NewCleanupListener(context.Background(), targets))
}

// RegisterSendResultListener resolver method
func (r *Resolver) RegisterSendResultListener() {
	targets := r.TargetClients()

	if len(targets) == 0 {
		return
	}

	if r.resultListener == nil {
		r.RegisterNewResultsListener()
	}

	r.resultListener.RegisterListener(listener.NewSendResultListener(targets))
}

// RegisterSendResultListener resolver method
func (r *Resolver) UnregisterSendResultListener() {
	if r.ResultCache().Shared() {
		r.EventPublisher().UnregisterListener(listener.NewResults)
	}

	if r.resultListener == nil {
		return
	}

	r.resultListener.UnregisterListener()
}

// RegisterSendResultListener resolver method
func (r *Resolver) RegisterStoreListener(ctx context.Context, store report.PolicyReportStore) {
	r.EventPublisher().RegisterListener(listener.Store, listener.NewStoreListener(ctx, store))
}

// RegisterMetricsListener resolver method
func (r *Resolver) RegisterMetricsListener() {
	r.EventPublisher().RegisterListener(listener.Metrics, listener.NewMetricsListener(
		metrics.NewResultFilter(
			ToRuleSet(r.config.Metrics.Filter.Namespaces),
			ToRuleSet(r.config.Metrics.Filter.Status),
			ToRuleSet(r.config.Metrics.Filter.Policies),
			ToRuleSet(r.config.Metrics.Filter.Sources),
			ToRuleSet(r.config.Metrics.Filter.Severities),
		),
		metrics.NewReportFilter(
			ToRuleSet(r.config.Metrics.Filter.Namespaces),
			ToRuleSet(r.config.Metrics.Filter.Sources),
		),
		r.config.Metrics.Mode,
		r.config.Metrics.CustomLabels,
	))
}

// SecretClient resolver method
func (r *Resolver) SecretClient() secrets.Client {
	clientset, err := k8s.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil
	}

	return secrets.NewClient(clientset.CoreV1().Secrets(r.config.Namespace))
}

func (r *Resolver) TargetFactory() *TargetFactory {
	return &TargetFactory{
		secretClient: r.SecretClient(),
	}
}

func (r *Resolver) DatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{
		secretClient: r.SecretClient(),
	}
}

// TargetClients resolver method
func (r *Resolver) TargetClients() []target.Client {
	if r.targetsCreated {
		return r.targetClients
	}

	factory := r.TargetFactory()

	clients := make([]target.Client, 0)

	clients = append(clients, factory.LokiClients(r.config.Loki)...)
	clients = append(clients, factory.ElasticsearchClients(r.config.Elasticsearch)...)
	clients = append(clients, factory.SlackClients(r.config.Slack)...)
	clients = append(clients, factory.DiscordClients(r.config.Discord)...)
	clients = append(clients, factory.TeamsClients(r.config.Teams)...)
	clients = append(clients, factory.S3Clients(r.config.S3)...)
	clients = append(clients, factory.KinesisClients(r.config.Kinesis)...)
	clients = append(clients, factory.SecurityHubs(r.config.SecurityHub)...)
	clients = append(clients, factory.WebhookClients(r.config.Webhook)...)
	clients = append(clients, factory.GCSClients(r.config.GCS)...)
	clients = append(clients, factory.TelegramClients(r.config.Telegram)...)
	clients = append(clients, factory.GoogleChatClients(r.config.GoogleChat)...)

	if ui := factory.UIClient(r.config.UI); ui != nil {
		clients = append(clients, ui)
	}

	r.targetClients = clients
	r.targetsCreated = true

	return r.targetClients
}

func (r *Resolver) HasTargets() bool {
	return len(r.TargetClients()) > 0
}

func (r *Resolver) EnableLeaderElection() bool {
	if !r.config.LeaderElection.Enabled {
		return false
	}

	if !r.HasTargets() && r.Database().Dialect().Name() == dialect.SQLite {
		return false
	}

	return true
}

// SkipExistingOnStartup config method
func (r *Resolver) SkipExistingOnStartup() bool {
	for _, client := range r.TargetClients() {
		if !client.SkipExistingOnStartup() {
			return false
		}
	}

	return true
}

func (r *Resolver) CRDClient() (wgpolicyk8sv1alpha2.Wgpolicyk8sV1alpha2Interface, error) {
	client, err := versioned.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	return client.Wgpolicyk8sV1alpha2(), nil
}

func (r *Resolver) CRDMetadataClient() (metadata.Interface, error) {
	client, err := metadata.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *Resolver) SummaryGenerator() (*summary.Generator, error) {
	client, err := r.CRDClient()
	if err != nil {
		return nil, err
	}

	return summary.NewGenerator(
		client,
		EmailReportFilterFromConfig(r.config.EmailReports.Summary.Filter),
		!r.config.EmailReports.Summary.Filter.DisableClusterReports,
	), nil
}

func (r *Resolver) SummaryReporter() *summary.Reporter {
	return summary.NewReporter(
		r.config.Templates.Dir,
		r.config.EmailReports.ClusterName,
		r.config.EmailReports.TitlePrefix,
	)
}

func (r *Resolver) ViolationsGenerator() (*violations.Generator, error) {
	client, err := r.CRDClient()
	if err != nil {
		return nil, err
	}

	return violations.NewGenerator(
		client,
		EmailReportFilterFromConfig(r.config.EmailReports.Violations.Filter),
		!r.config.EmailReports.Violations.Filter.DisableClusterReports,
	), nil
}

func (r *Resolver) ViolationsReporter() *violations.Reporter {
	return violations.NewReporter(
		r.config.Templates.Dir,
		r.config.EmailReports.ClusterName,
		r.config.EmailReports.TitlePrefix,
	)
}

func (r *Resolver) SMTPServer() *mail.SMTPServer {
	smtp := r.config.EmailReports.SMTP

	server := mail.NewSMTPClient()
	server.Host = smtp.Host
	server.Port = smtp.Port
	server.Username = smtp.Username
	server.Password = smtp.Password
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	server.Encryption = email.EncryptionFromString(smtp.Encryption)
	server.TLSConfig = &tls.Config{InsecureSkipVerify: smtp.SkipTLS}

	if smtp.Certificate != "" {
		caCert, err := os.ReadFile(smtp.Certificate)
		if err != nil {
			zap.L().Error("failed to read certificate for SMTP Client", zap.String("path", smtp.Certificate))
			return server
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		server.TLSConfig.RootCAs = caCertPool
	}

	return server
}

func (r *Resolver) EmailClient() *email.Client {
	return email.NewClient(r.config.EmailReports.SMTP.From, r.SMTPServer())
}

func (r *Resolver) PolicyReportClient() (report.PolicyReportClient, error) {
	if r.policyReportClient != nil {
		return r.policyReportClient, nil
	}

	client, err := r.CRDMetadataClient()
	if err != nil {
		return nil, err
	}

	queue, err := r.Queue()
	if err != nil {
		return nil, err
	}

	r.policyReportClient = kubernetes.NewPolicyReportClient(client, r.ReportFilter(), queue)

	return r.policyReportClient, nil
}

func (r *Resolver) ReportFilter() *report.Filter {
	return report.NewFilter(
		r.config.ReportFilter.ClusterReports.Disabled,
		ToRuleSet(r.config.ReportFilter.Namespaces),
	)
}

// ResultCache resolver method
func (r *Resolver) ResultCache() cache.Cache {
	if r.resultCache != nil {
		return r.resultCache
	}

	if r.config.Redis.Enabled {
		r.resultCache = cache.NewRedisCache(
			r.config.Redis.Prefix,
			goredis.NewClient(&goredis.Options{
				Addr:     r.config.Redis.Address,
				Username: r.config.Redis.Username,
				Password: r.config.Redis.Password,
				DB:       r.config.Redis.Database,
			}),
			6*time.Hour,
		)
	} else {
		r.resultCache = cache.NewInMermoryCache(6*time.Hour, 10*time.Minute)
	}

	return r.resultCache
}

// Logger resolver method
func (r *Resolver) Logger() (*zap.Logger, error) {
	if r.logger != nil {
		return r.logger, nil
	}

	encoder := zap.NewProductionEncoderConfig()
	if r.config.Logging.Development {
		encoder = zap.NewDevelopmentEncoderConfig()
		encoder.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	ouput := "json"
	if r.config.Logging.Encoding != "json" {
		ouput = "console"
		encoder.EncodeCaller = nil
	}

	var sampling *zap.SamplingConfig
	if !r.config.Logging.Development {
		sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.Level(r.config.Logging.LogLevel)),
		Development:       r.config.Logging.Development,
		Sampling:          sampling,
		Encoding:          ouput,
		EncoderConfig:     encoder,
		DisableStacktrace: !r.config.Logging.Development,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	r.logger = logger

	zap.ReplaceGlobals(logger)

	return r.logger, nil
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}

func EmailReportFilterFromConfig(config EmailReportFilter) email.Filter {
	return email.NewFilter(ToRuleSet(config.Namespaces), ToRuleSet(config.Sources))
}

func ToRuleSet(filter ValueFilter) validate.RuleSets {
	return validate.RuleSets{
		Include: filter.Include,
		Exclude: filter.Exclude,
	}
}
