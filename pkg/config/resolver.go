package config

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/leaderelection"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/redis"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/sqlite3"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
	"github.com/kyverno/policy-reporter/pkg/target/kinesis"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
	"github.com/kyverno/policy-reporter/pkg/target/ui"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
	"github.com/kyverno/policy-reporter/pkg/validate"
	mail "github.com/xhit/go-simple-mail/v2"

	goredis "github.com/go-redis/redis/v8"
	"github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	wgpolicyk8sv1alpha2 "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha2"
	_ "github.com/mattn/go-sqlite3"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config             *Config
	k8sConfig          *rest.Config
	mapper             kubernetes.Mapper
	publisher          report.EventPublisher
	policyStore        sqlite3.PolicyReportStore
	policyReportClient report.PolicyReportClient
	leaderElector      *leaderelection.Client
	targetClients      []target.Client
	resultCache        cache.Cache
}

// APIServer resolver method
func (r *Resolver) APIServer(synced func() bool) api.Server {
	return api.NewServer(
		r.TargetClients(),
		r.config.API.Port,
		synced,
	)
}

// Database resolver method
func (r *Resolver) Database() (*sql.DB, error) {
	return sqlite3.NewDatabase(r.config.DBFile)
}

// PolicyReportStore resolver method
func (r *Resolver) PolicyReportStore(db *sql.DB) (sqlite3.PolicyReportStore, error) {
	if r.policyStore != nil {
		return r.policyStore, nil
	}

	s, err := sqlite3.NewPolicyReportStore(db)
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

// RegisterSendResultListener resolver method
func (r *Resolver) RegisterSendResultListener() {
	targets := r.TargetClients()
	if len(targets) > 0 {
		newResultListener := listener.NewResultListener(r.SkipExistingOnStartup(), r.ResultCache(), time.Now())
		newResultListener.RegisterListener(listener.NewSendResultListener(targets))

		r.EventPublisher().RegisterListener(listener.NewResults, newResultListener.Listen)
	}
}

// RegisterSendResultListener resolver method
func (r *Resolver) RegisterStoreListener(store report.PolicyReportStore) {
	r.EventPublisher().RegisterListener(listener.Store, listener.NewStoreListener(store))
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

// Mapper resolver method
func (r *Resolver) Mapper() kubernetes.Mapper {
	if r.mapper != nil {
		return r.mapper
	}

	mapper := kubernetes.NewMapper(r.config.PriorityMap)

	r.mapper = mapper

	return mapper
}

// LokiClients resolver method
func (r *Resolver) LokiClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Loki.Name == "" {
		r.config.Loki.Name = "Loki"
	}
	if r.config.Loki.Path == "" {
		r.config.Loki.Path = "/api/prom/push"
	}

	if loki := createLokiClient(r.config.Loki, Loki{}); loki != nil {
		clients = append(clients, loki)
	}
	for i, channel := range r.config.Loki.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Loki Channel %d", i+1)
		}

		if loki := createLokiClient(channel, r.config.Loki); loki != nil {
			clients = append(clients, loki)
		}
	}

	return clients
}

// ElasticsearchClients resolver method
func (r *Resolver) ElasticsearchClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Elasticsearch.Name == "" {
		r.config.Elasticsearch.Name = "Elasticsearch"
	}

	if es := createElasticsearchClient(r.config.Elasticsearch, Elasticsearch{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Elasticsearch.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Elasticsearch Channel %d", i+1)
		}

		if es := createElasticsearchClient(channel, r.config.Elasticsearch); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// SlackClients resolver method
func (r *Resolver) SlackClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Slack.Name == "" {
		r.config.Slack.Name = "Slack"
	}

	if es := createSlackClient(r.config.Slack, Slack{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Slack.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Slack Channel %d", i+1)
		}

		if es := createSlackClient(channel, r.config.Slack); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// DiscordClients resolver method
func (r *Resolver) DiscordClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Discord.Name == "" {
		r.config.Discord.Name = "Discord"
	}

	if es := createDiscordClient(r.config.Discord, Discord{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Discord.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Discord Channel %d", i+1)
		}

		if es := createDiscordClient(channel, r.config.Discord); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// TeamsClients resolver method
func (r *Resolver) TeamsClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Teams.Name == "" {
		r.config.Teams.Name = "Teams"
	}

	if es := createTeamsClient(r.config.Teams, Teams{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Teams.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Teams Channel %d", i+1)
		}

		if es := createTeamsClient(channel, r.config.Teams); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// WebhookClients resolver method
func (r *Resolver) WebhookClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Webhook.Name == "" {
		r.config.Webhook.Name = "Webhook"
	}

	if es := createWebhookClient(r.config.Webhook, Webhook{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Webhook.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Webhook Channel %d", i+1)
		}

		if es := createWebhookClient(channel, r.config.Webhook); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// UIClient resolver method
func (r *Resolver) UIClient() target.Client {
	if r.config.UI.Host == "" {
		return nil
	}

	log.Println("[INFO] UI configured")

	return ui.NewClient(
		"UI",
		r.config.UI.Host,
		r.config.UI.SkipExisting,
		createTargetFilter(TargetFilter{}, r.config.UI.MinimumPriority, r.config.UI.Sources),
		&http.Client{},
	)
}

// S3Clients resolver method
func (r *Resolver) S3Clients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.S3.Name == "" {
		r.config.S3.Name = "S3"
	}

	if es := createS3Client(r.config.S3, S3{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.S3.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("S3 Channel %d", i+1)
		}

		if es := createS3Client(channel, r.config.S3); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// KinesisClients resolver method
func (r *Resolver) KinesisClients() []target.Client {
	clients := make([]target.Client, 0)
	if r.config.Kinesis.Name == "" {
		r.config.Kinesis.Name = "Kinesis"
	}

	if es := createKinesisClient(r.config.Kinesis, Kinesis{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Kinesis.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Kinesis Channel %d", i+1)
		}

		if es := createKinesisClient(channel, r.config.Kinesis); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// TargetClients resolver method
func (r *Resolver) TargetClients() []target.Client {
	if len(r.targetClients) > 0 {
		return r.targetClients
	}

	clients := make([]target.Client, 0)

	clients = append(clients, r.LokiClients()...)
	clients = append(clients, r.ElasticsearchClients()...)
	clients = append(clients, r.SlackClients()...)
	clients = append(clients, r.DiscordClients()...)
	clients = append(clients, r.TeamsClients()...)
	clients = append(clients, r.S3Clients()...)
	clients = append(clients, r.KinesisClients()...)
	clients = append(clients, r.WebhookClients()...)

	if ui := r.UIClient(); ui != nil {
		clients = append(clients, ui)
	}

	r.targetClients = clients

	return r.targetClients
}

func (r *Resolver) HasTargets() bool {
	return len(r.TargetClients()) > 0
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
		r.config.EmailReports.Templates.Dir,
		r.config.EmailReports.ClusterName,
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
		r.config.EmailReports.Templates.Dir,
		r.config.EmailReports.ClusterName,
	)
}

func (r *Resolver) SMTPServer() *mail.SMTPServer {
	server := mail.NewSMTPClient()
	server.Host = r.config.EmailReports.SMTP.Host
	server.Port = r.config.EmailReports.SMTP.Port
	server.Username = r.config.EmailReports.SMTP.Username
	server.Password = r.config.EmailReports.SMTP.Password
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	server.Encryption = email.EncryptionFromString(r.config.EmailReports.SMTP.Encryption)

	return server
}

func (r *Resolver) EmailClient() *email.Client {
	return email.NewClient(r.config.EmailReports.SMTP.From, r.SMTPServer())
}

func (r *Resolver) PolicyReportClient() (report.PolicyReportClient, error) {
	if r.policyReportClient != nil {
		return r.policyReportClient, nil
	}

	client, err := versioned.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	r.policyReportClient = kubernetes.NewPolicyReportClient(client, r.Mapper(), r.ReportFilter(), r.EventPublisher())

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
		r.resultCache = redis.New(
			r.config.Redis.Prefix,
			goredis.NewClient(&goredis.Options{
				Addr:     r.config.Redis.Address,
				Username: r.config.Redis.Username,
				Password: r.config.Redis.Password,
				DB:       r.config.Redis.Database,
			}),
			2*time.Hour,
		)
	} else {
		r.resultCache = cache.New(time.Minute*150, time.Minute*15)
	}

	return r.resultCache
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}

func createSlackClient(config Slack, parent Slack) target.Client {
	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", config.Name)

	return slack.NewClient(
		config.Name,
		config.Webhook,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createLokiClient(config Loki, parent Loki) target.Client {
	if config.Host == "" && parent.Host == "" {
		return nil
	} else if config.Host == "" {
		config.Host = parent.Host
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if config.Path == "" {
		config.Path = parent.Path
	}

	log.Printf("[INFO] %s configured", config.Name)

	return loki.NewClient(
		config.Name,
		config.Host+config.Path,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		config.CustomLabels,
		&http.Client{},
	)
}

func createElasticsearchClient(config Elasticsearch, parent Elasticsearch) target.Client {
	if config.Host == "" && parent.Host == "" {
		return nil
	} else if config.Host == "" {
		config.Host = parent.Host
	}

	if config.Index == "" && parent.Index == "" {
		config.Index = "policy-reporter"
	} else if config.Index == "" {
		config.Index = parent.Index
	}

	if config.Rotation == "" && parent.Rotation == "" {
		config.Rotation = elasticsearch.Dayli
	} else if config.Rotation == "" {
		config.Rotation = parent.Rotation
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", config.Name)

	return elasticsearch.NewClient(
		config.Name,
		config.Host,
		config.Index,
		config.Rotation,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createDiscordClient(config Discord, parent Discord) target.Client {
	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", config.Name)

	return discord.NewClient(
		config.Name,
		config.Webhook,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createTeamsClient(config Teams, parent Teams) target.Client {
	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	if !config.SkipTLS {
		config.SkipTLS = parent.SkipTLS
	}

	log.Printf("[INFO] %s configured", config.Name)

	client := &http.Client{}

	if config.SkipTLS {
		client.Transport = http.DefaultTransport.(*http.Transport).Clone()
		client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: config.SkipTLS,
		}
	}

	return teams.NewClient(
		config.Name,
		config.Webhook,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		client,
	)
}

func createWebhookClient(config Webhook, parent Webhook) target.Client {
	if config.Host == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	if len(parent.Headers) > 0 {
		headers := map[string]string{}
		for header, value := range parent.Headers {
			headers[header] = value
		}
		for header, value := range config.Headers {
			headers[header] = value
		}

		config.Headers = headers
	}

	log.Printf("[INFO] %s configured", config.Name)

	return webhook.NewClient(
		config.Name,
		config.Host,
		config.Headers,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createS3Client(config S3, parent S3) target.Client {
	if config.Endpoint == "" && parent.Endpoint == "" {
		return nil
	} else if config.Endpoint == "" {
		config.Endpoint = parent.Endpoint
	}

	if config.AccessKeyID == "" && parent.AccessKeyID == "" {
		log.Printf("[ERROR] %s.AccessKeyID has not been declared", config.Name)
		return nil
	} else if config.AccessKeyID == "" {
		config.AccessKeyID = parent.AccessKeyID
	}

	if config.SecretAccessKey == "" && parent.SecretAccessKey == "" {
		log.Printf("[ERROR] %s.SecretAccessKey has not been declared", config.Name)
		return nil
	} else if config.SecretAccessKey == "" {
		config.SecretAccessKey = parent.SecretAccessKey
	}

	if config.Region == "" && parent.Region == "" {
		log.Printf("[ERROR] %s.Region has not been declared", config.Name)
		return nil
	} else if config.Region == "" {
		config.Region = parent.Region
	}

	if config.Bucket == "" && parent.Bucket == "" {
		log.Printf("[ERROR] %s.Bucket has not been declared", config.Name)
		return nil
	} else if config.Bucket == "" {
		config.Bucket = parent.Bucket
	}

	if config.Prefix == "" && parent.Prefix == "" {
		config.Prefix = "policy-reporter"
	} else if config.Prefix == "" {
		config.Prefix = parent.Prefix
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	s3Client := helper.NewS3Client(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.Bucket,
	)

	log.Printf("[INFO] %s configured", config.Name)

	return s3.NewClient(
		config.Name,
		s3Client,
		config.Prefix,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
	)
}

func createKinesisClient(config Kinesis, parent Kinesis) target.Client {
	if config.Endpoint == "" && parent.Endpoint == "" {
		return nil
	} else if config.Endpoint == "" {
		config.Endpoint = parent.Endpoint
	}

	if config.AccessKeyID == "" && parent.AccessKeyID == "" {
		log.Printf("[ERROR] %s.AccessKeyID has not been declared", config.Name)
		return nil
	} else if config.AccessKeyID == "" {
		config.AccessKeyID = parent.AccessKeyID
	}

	if config.SecretAccessKey == "" && parent.SecretAccessKey == "" {
		log.Printf("[ERROR] %s.SecretAccessKey has not been declared", config.Name)
		return nil
	} else if config.SecretAccessKey == "" {
		config.SecretAccessKey = parent.SecretAccessKey
	}

	if config.Region == "" && parent.Region == "" {
		log.Printf("[ERROR] %s.Region has not been declared", config.Name)
		return nil
	} else if config.Region == "" {
		config.Region = parent.Region
	}

	if config.StreamName == "" && parent.StreamName == "" {
		log.Printf("[ERROR] %s.StreamName has not been declared", config.Name)
		return nil
	} else if config.StreamName == "" {
		config.StreamName = parent.StreamName
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	kinesisClient := helper.NewKinesisClient(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.StreamName,
	)

	log.Printf("[INFO] %s configured", config.Name)

	return kinesis.NewClient(
		config.Name,
		kinesisClient,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
	)
}

func createTargetFilter(filter TargetFilter, minimumPriority string, sources []string) *report.ResultFilter {
	return target.NewClientFilter(
		ToRuleSet(filter.Namespaces),
		ToRuleSet(filter.Priorities),
		ToRuleSet(filter.Policies),
		minimumPriority,
		sources,
	)
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
