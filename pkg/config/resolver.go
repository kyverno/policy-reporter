package config

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/discord"
	"github.com/fjogeleit/policy-reporter/pkg/target/elasticsearch"
	"github.com/fjogeleit/policy-reporter/pkg/target/loki"
	"github.com/fjogeleit/policy-reporter/pkg/target/slack"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config              *Config
	k8sConfig           *rest.Config
	kubeClient          report.Client
	lokiClient          target.Client
	elasticsearchClient target.Client
	slackClient         target.Client
	discordClient       target.Client
}

// PolicyReportClient resolver method
func (r *Resolver) PolicyReportClient() (report.Client, error) {
	if r.kubeClient != nil {
		return r.kubeClient, nil
	}

	cmAPI, err := r.configMapAPI()
	if err != nil {
		return nil, err
	}

	policyAPI, err := r.policyReportAPI()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		policyAPI,
		cmAPI,
		time.Now(),
	)

	r.kubeClient = client

	return client, err
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

	return clients
}

func (r *Resolver) SkipExistingOnStartup() bool {
	for _, client := range r.TargetClients() {
		if !client.SkipExistingOnStartup() {
			return false
		}
	}

	return true
}

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

func (r *Resolver) policyReportAPI() (kubernetes.PolicyReportAdapter, error) {
	client, err := dynamic.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewPolicyReportAdapter(client), nil
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}
