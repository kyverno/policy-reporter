package config

import (
	"database/sql"
	"fmt"
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
	"github.com/kyverno/policy-reporter/pkg/target/webhook"

	"github.com/patrickmn/go-cache"
	"k8s.io/client-go/dynamic"

	_ "github.com/mattn/go-sqlite3"
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
	targetClients      []target.Client
	resultCache        *cache.Cache
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

// LokiClients resolver method
func (r *Resolver) LokiClients() []target.Client {
	clients := make([]target.Client, 0)

	if loki := createLokiClient(r.config.Loki, Loki{}, "Loki"); loki != nil {
		clients = append(clients, loki)
	}
	for i, channel := range r.config.Loki.Channels {
		if loki := createLokiClient(channel, r.config.Loki, fmt.Sprintf("Loki Channel %d", i+1)); loki != nil {
			clients = append(clients, loki)
		}
	}

	return clients
}

// ElasticsearchClients resolver method
func (r *Resolver) ElasticsearchClients() []target.Client {
	clients := make([]target.Client, 0)

	if es := createElasticsearchClient(r.config.Elasticsearch, Elasticsearch{}, "Elasticsearch"); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Elasticsearch.Channels {
		if es := createElasticsearchClient(channel, r.config.Elasticsearch, fmt.Sprintf("Elasticsearch Channel %d", i+1)); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// SlackClients resolver method
func (r *Resolver) SlackClients() []target.Client {
	clients := make([]target.Client, 0)

	if es := createSlackClient(r.config.Slack, Slack{}, "Slack"); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Slack.Channels {
		if es := createSlackClient(channel, r.config.Slack, fmt.Sprintf("Slack Channel %d", i+1)); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// DiscordClients resolver method
func (r *Resolver) DiscordClients() []target.Client {
	clients := make([]target.Client, 0)

	if es := createDiscordClient(r.config.Discord, Discord{}, "Discord"); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Discord.Channels {
		if es := createDiscordClient(channel, r.config.Discord, fmt.Sprintf("Discord Channel %d", i+1)); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// TeamsClients resolver method
func (r *Resolver) TeamsClients() []target.Client {
	clients := make([]target.Client, 0)

	if es := createTeamsClient(r.config.Teams, Teams{}, "Teams"); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Teams.Channels {
		if es := createTeamsClient(channel, r.config.Teams, fmt.Sprintf("Teams Channel %d", i+1)); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// WebhookClients resolver method
func (r *Resolver) WebhookClients() []target.Client {
	clients := make([]target.Client, 0)

	if es := createWebhookClient(r.config.Webhook, Webhook{}, "Webhook"); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.Webhook.Channels {
		if es := createWebhookClient(channel, r.config.Webhook, fmt.Sprintf("Webhook Channel %d", i+1)); es != nil {
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
		createTargetFilter(Filter{}, r.config.UI.MinimumPriority, r.config.UI.Sources),
		&http.Client{},
	)
}

// TeamsClients resolver method
func (r *Resolver) S3Clients() []target.Client {
	clients := make([]target.Client, 0)

	if es := createS3Client(r.config.S3, S3{}, "S3"); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range r.config.S3.Channels {
		if es := createS3Client(channel, r.config.S3, fmt.Sprintf("S3 Channel %d", i+1)); es != nil {
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
	clients = append(clients, r.WebhookClients()...)

	if ui := r.UIClient(); ui != nil {
		clients = append(clients, ui)
	}

	r.targetClients = clients

	return r.targetClients
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

	r.policyReportClient = kubernetes.NewPolicyReportClient(client, r.Mapper(), 5*time.Second, r.ReportFilter())

	return r.policyReportClient, nil
}

func (r *Resolver) ReportFilter() report.Filter {
	return report.NewFilter(
		r.config.ReportFilter.ClusterReports.Disabled,
		r.config.ReportFilter.Namespaces.Include,
		r.config.ReportFilter.Namespaces.Exclude,
	)
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

func createSlackClient(config Slack, parent Slack, name string) target.Client {
	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", name)

	return slack.NewClient(
		name,
		config.Webhook,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createLokiClient(config Loki, parent Loki, name string) target.Client {
	if config.Host == "" && parent.Host == "" {
		return nil
	} else if config.Host == "" {
		config.Host = parent.Host
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", name)

	return loki.NewClient(
		name,
		config.Host,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		config.CustomLabels,
		&http.Client{},
	)
}

func createElasticsearchClient(config Elasticsearch, parent Elasticsearch, name string) target.Client {
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

	log.Printf("[INFO] %s configured", name)

	return elasticsearch.NewClient(
		name,
		config.Host,
		config.Index,
		config.Rotation,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createDiscordClient(config Discord, parent Discord, name string) target.Client {
	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", name)

	return discord.NewClient(
		name,
		config.Webhook,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createTeamsClient(config Teams, parent Teams, name string) target.Client {
	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	log.Printf("[INFO] %s configured", name)

	return teams.NewClient(
		name,
		config.Webhook,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createWebhookClient(config Webhook, parent Webhook, name string) target.Client {
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

	log.Printf("[INFO] %s configured", name)

	return webhook.NewClient(
		name,
		config.Host,
		config.Headers,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
		&http.Client{},
	)
}

func createS3Client(config S3, parent S3, name string) target.Client {
	if config.Endpoint == "" && parent.Endpoint == "" {
		return nil
	} else if config.Endpoint == "" {
		config.Endpoint = parent.Endpoint
	}

	if config.AccessKeyID == "" && parent.AccessKeyID == "" {
		log.Printf("[ERROR] %s.AccessKeyID has not been declared", name)
		return nil
	} else if config.AccessKeyID == "" {
		config.AccessKeyID = parent.AccessKeyID
	}

	if config.SecretAccessKey == "" && parent.SecretAccessKey == "" {
		log.Printf("[ERROR] %s.SecretAccessKey has not been declared", name)
		return nil
	} else if config.SecretAccessKey == "" {
		config.SecretAccessKey = parent.SecretAccessKey
	}

	if config.Region == "" && parent.Region == "" {
		log.Printf("[ERROR] %s.Region has not been declared", name)
		return nil
	} else if config.Region == "" {
		config.Region = parent.Region
	}

	if config.Bucket == "" && parent.Bucket == "" {
		log.Printf("[ERROR] %s.Bucket has not been declared", name)
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

	s3Client := helper.NewClient(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.Bucket,
	)

	log.Printf("[INFO] %s configured", name)

	return s3.NewClient(
		name,
		s3Client,
		config.Prefix,
		config.SkipExisting,
		createTargetFilter(config.Filter, config.MinimumPriority, config.Sources),
	)
}

func createTargetFilter(filter Filter, minimumPriority string, sources []string) *target.Filter {
	return &target.Filter{
		MinimumPriority: minimumPriority,
		Sources:         sources,
		Namespace: target.Rules{
			Include: filter.Namespaces.Include,
			Exclude: filter.Namespaces.Exclude,
		},
		Priority: target.Rules{
			Include: filter.Priorities.Include,
			Exclude: filter.Priorities.Exclude,
		},
		Policy: target.Rules{
			Include: filter.Policies.Include,
			Exclude: filter.Policies.Exclude,
		},
	}
}
