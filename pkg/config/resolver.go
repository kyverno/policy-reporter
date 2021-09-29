package config

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
	"github.com/kyverno/policy-reporter/pkg/target/ui"
	"github.com/kyverno/policy-reporter/pkg/target/yandex"

	"github.com/patrickmn/go-cache"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config              *Config
	k8sConfig           *rest.Config
	mapper              kubernetes.Mapper
	policyAdapter       kubernetes.PolicyReportAdapter
	policyStore         *report.PolicyReportStore
	policyClient        report.PolicyResultClient
	lokiClient          target.Client
	elasticsearchClient target.Client
	slackClient         target.Client
	discordClient       target.Client
	teamsClient         target.Client
	uiClient            target.Client
	yandexClient        target.Client
	resultCache         *cache.Cache
}

// APIServer resolver method
func (r *Resolver) APIServer() api.Server {
	foundResources := make(map[string]string, 0)

	client := r.policyClient
	if client != nil {
		foundResources = client.GetFoundResources()
	}

	return api.NewServer(
		r.PolicyReportStore(),
		r.TargetClients(),
		r.config.API.Port,
		foundResources,
	)
}

// PolicyReportStore resolver method
func (r *Resolver) PolicyReportStore() *report.PolicyReportStore {
	if r.policyStore != nil {
		return r.policyStore
	}

	r.policyStore = report.NewPolicyReportStore()

	return r.policyStore
}

// PolicyReportClient resolver method
func (r *Resolver) PolicyReportClient(ctx context.Context) (report.PolicyResultClient, error) {
	if r.policyClient != nil {
		return r.policyClient, nil
	}

	policyAPI, err := r.policyReportAPI(ctx)
	if err != nil {
		return nil, err
	}

	client := kubernetes.NewPolicyReportClient(
		policyAPI,
		r.PolicyReportStore(),
		time.Now(),
		r.ResultCache(),
	)

	r.policyClient = client

	return client, nil
}

// Mapper resolver method
func (r *Resolver) Mapper(ctx context.Context) (kubernetes.Mapper, error) {
	if r.mapper != nil {
		return r.mapper, nil
	}

	cmAPI, err := r.configMapAPI()
	if err != nil {
		return nil, err
	}

	mapper := kubernetes.NewMapper(make(map[string]string), cmAPI)
	mapper.FetchPriorities(ctx)
	go mapper.SyncPriorities(ctx)

	r.mapper = mapper

	return mapper, err
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
		r.config.UI.SkipExisting,
		&http.Client{},
	)

	log.Println("[INFO] UI configured")

	return r.uiClient
}

func (r *Resolver) YandexClient() target.Client {
	if r.yandexClient != nil {
		return r.yandexClient
	}
	if r.config.Yandex.AccessKeyID == "" && r.config.Yandex.SecretAccessKey == "" {
		return nil
	}

	if r.config.Yandex.Region == "" {
		log.Printf("[INFO] Yandex.Region has not been declared using ru-central1")
		r.config.Yandex.Region = "ru-central1"
	}
	if r.config.Yandex.Endpoint == "" {
		log.Printf("[INFO] Yandex.Endpoint has not been declared using ru-central1")
		r.config.Yandex.Endpoint = "https://storage.yandexcloud.net"
	}
	if r.config.Yandex.Prefix == "" {
		log.Printf("[INFO] Yandex.Prefix has not been declared using policy-reporter prefix")
		r.config.Yandex.Prefix = "policy-reporter/"
	}

	log.Printf("[INFO] Yandex Session has been configured successfully")
	if r.config.Yandex.Bucket == "" || r.config.Yandex.AccessKeyID == "" || r.config.Yandex.SecretAccessKey == "" {
		log.Printf("[ERROR] One of Yandex.Bucket,Yandex.AccessKeyID or Yandex.SecretAccessKey  has not been declared")
		return nil
	}

	r.yandexClient = yandex.NewClient(
		r.config.Yandex.AccessKeyID,
		r.config.Yandex.SecretAccessKey,
		r.config.Yandex.Region,
		r.config.Yandex.Endpoint,
		r.config.Yandex.Prefix,
		r.config.Yandex.Bucket,
		r.config.Yandex.MinimumPriority,
		r.config.Yandex.SkipExisting,
	)

	log.Println("[INFO] Yandex Session configured")

	return r.yandexClient
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

	if yandex := r.YandexClient(); yandex != nil {
		clients = append(clients, yandex)
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

// ConfigMapClient resolver method
func (r *Resolver) ConfigMapClient() (v1.ConfigMapInterface, error) {
	var err error

	client, err := v1.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	return client.ConfigMaps(r.config.Namespace), nil
}

func (r *Resolver) configMapAPI() (kubernetes.ConfigMapAdapter, error) {
	client, err := r.ConfigMapClient()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewConfigMapAdapter(client), nil
}

func (r *Resolver) policyReportAPI(ctx context.Context) (kubernetes.PolicyReportAdapter, error) {
	if r.policyAdapter != nil {
		return r.policyAdapter, nil
	}

	client, err := dynamic.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}
	mapper, err := r.Mapper(ctx)
	if err != nil {
		return nil, err
	}

	r.policyAdapter = kubernetes.NewPolicyReportAdapter(client, mapper)

	return r.policyAdapter, nil
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
