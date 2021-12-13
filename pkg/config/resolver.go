package config

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/sqlite3"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
	"github.com/kyverno/policy-reporter/pkg/target/ui"

	"github.com/patrickmn/go-cache"
	"k8s.io/client-go/dynamic"

	_ "github.com/mattn/go-sqlite3"
	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config              *Config
	k8sConfig           *rest.Config
	mapper              kubernetes.Mapper
	publisher           report.EventPublisher
	policyStore         sqlite3.PolicyReportStore
	policyReportClient  report.PolicyReportClient
	lokiClient          target.Client
	elasticsearchClient target.Client
	slackClient         target.Client
	discordClient       target.Client
	teamsClient         target.Client
	uiClient            target.Client
	s3Client            target.Client
	resultCache         *cache.Cache
}

// APIServer resolver method
func (r *Resolver) APIServer(foundResources map[string]string) api.Server {
	return api.NewServer(
		r.TargetClients(),
		r.config.API.Port,
		foundResources,
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

		r.EventPublisher().RegisterListener(newResultListener.Listen)
	}
}

// RegisterSendResultListener resolver method
func (r *Resolver) RegisterStoreListener(store report.PolicyReportStore) {
	r.EventPublisher().RegisterListener(listener.NewStoreListener(store))
}

// RegisterMetricsListener resolver method
func (r *Resolver) RegisterMetricsListener() {
	r.EventPublisher().RegisterListener(listener.NewMetricsListener())
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

// LokiClient resolver method
func (r *Resolver) LokiClient() target.Client {
	if r.lokiClient != nil {
		return r.lokiClient
	}

	if r.config.Loki.Host == "" {
		return nil
	}

	r.lokiClient = loki.NewClient(
		r.config.Loki.Host,
		r.config.Loki.MinimumPriority,
		r.config.Loki.Sources,
		r.config.Loki.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] Loki configured")

	return r.lokiClient
}

// ElasticsearchClient resolver method
func (r *Resolver) ElasticsearchClient() target.Client {
	if r.elasticsearchClient != nil {
		return r.elasticsearchClient
	}

	if r.config.Elasticsearch.Host == "" {
		return nil
	}
	if r.config.Elasticsearch.Index == "" {
		r.config.Elasticsearch.Index = "policy-reporter"
	}
	if r.config.Elasticsearch.Rotation == "" {
		r.config.Elasticsearch.Rotation = elasticsearch.Dayli
	}

	r.elasticsearchClient = elasticsearch.NewClient(
		r.config.Elasticsearch.Host,
		r.config.Elasticsearch.Index,
		r.config.Elasticsearch.Rotation,
		r.config.Elasticsearch.MinimumPriority,
		r.config.Elasticsearch.Sources,
		r.config.Elasticsearch.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] Elasticsearch configured")

	return r.elasticsearchClient
}

// SlackClient resolver method
func (r *Resolver) SlackClient() target.Client {
	if r.slackClient != nil {
		return r.slackClient
	}

	if r.config.Slack.Webhook == "" {
		return nil
	}

	r.slackClient = slack.NewClient(
		r.config.Slack.Webhook,
		r.config.Slack.MinimumPriority,
		r.config.Slack.Sources,
		r.config.Slack.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] Slack configured")

	return r.slackClient
}

// DiscordClient resolver method
func (r *Resolver) DiscordClient() target.Client {
	if r.discordClient != nil {
		return r.discordClient
	}

	if r.config.Discord.Webhook == "" {
		return nil
	}

	r.discordClient = discord.NewClient(
		r.config.Discord.Webhook,
		r.config.Discord.MinimumPriority,
		r.config.Discord.Sources,
		r.config.Discord.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] Discord configured")

	return r.discordClient
}

// TeamsClient resolver method
func (r *Resolver) TeamsClient() target.Client {
	if r.teamsClient != nil {
		return r.teamsClient
	}

	if r.config.Teams.Webhook == "" {
		return nil
	}

	r.teamsClient = teams.NewClient(
		r.config.Teams.Webhook,
		r.config.Teams.MinimumPriority,
		r.config.Teams.Sources,
		r.config.Teams.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] Teams configured")

	return r.teamsClient
}

// UIClient resolver method
func (r *Resolver) UIClient() target.Client {
	if r.uiClient != nil {
		return r.uiClient
	}

	if r.config.UI.Host == "" {
		return nil
	}

	r.uiClient = ui.NewClient(
		r.config.UI.Host,
		r.config.UI.MinimumPriority,
		r.config.UI.Sources,
		r.config.UI.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] UI configured")

	return r.uiClient
}

func (r *Resolver) S3Client() target.Client {
	if r.s3Client != nil {
		return r.s3Client
	}
	if r.config.S3.Endpoint == "" {
		return nil
	}
	if r.config.S3.AccessKeyID == "" {
		log.Printf("[ERROR] S3.AccessKeyID has not been declared")
		return nil
	}
	if r.config.S3.SecretAccessKey == "" {
		log.Printf("[ERROR] S3.SecretAccessKey has not been declared")
		return nil
	}
	if r.config.S3.Region == "" {
		log.Printf("[ERROR] S3.Region has not been declared")
		return nil
	}
	if r.config.S3.Bucket == "" {
		log.Printf("[ERROR] S3.Bucket has to be declared")
		return nil
	}
	if r.config.S3.Prefix == "" {
		r.config.S3.Prefix = "policy-reporter/"
	}

	s3Client := helper.NewClient(
		r.config.S3.AccessKeyID,
		r.config.S3.SecretAccessKey,
		r.config.S3.Region,
		r.config.S3.Endpoint,
		r.config.S3.Bucket,
	)

	r.s3Client = s3.NewClient(
		s3Client,
		r.config.S3.Prefix,
		r.config.S3.MinimumPriority,
		r.config.S3.Sources,
		r.config.S3.SkipExisting,
	)

	log.Println("[INFO] S3 configured")

	return r.s3Client
}

// TargetClients resolver method
func (r *Resolver) TargetClients() []target.Client {
	clients := make([]target.Client, 0)

	if loki := r.LokiClient(); loki != nil {
		clients = append(clients, loki)
	}

	if elasticsearch := r.ElasticsearchClient(); elasticsearch != nil {
		clients = append(clients, elasticsearch)
	}

	if slack := r.SlackClient(); slack != nil {
		clients = append(clients, slack)
	}

	if discord := r.DiscordClient(); discord != nil {
		clients = append(clients, discord)
	}

	if teams := r.TeamsClient(); teams != nil {
		clients = append(clients, teams)
	}

	if ui := r.UIClient(); ui != nil {
		clients = append(clients, ui)
	}

	if s3 := r.S3Client(); s3 != nil {
		clients = append(clients, s3)
	}

	return clients
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

func (r *Resolver) PolicyReportClient() (report.PolicyReportClient, error) {
	if r.policyReportClient != nil {
		return r.policyReportClient, nil
	}

	client, err := dynamic.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	r.policyReportClient = kubernetes.NewPolicyReportClient(client, r.Mapper(), 5*time.Second)

	return r.policyReportClient, nil
}

// ResultCache resolver method
func (r *Resolver) ResultCache() *cache.Cache {
	if r.resultCache != nil {
		return r.resultCache
	}
	r.resultCache = cache.New(time.Minute*150, time.Minute*15)

	return r.resultCache
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}
