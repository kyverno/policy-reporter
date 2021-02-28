package config

import (
	"context"
	"net/http"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/metrics"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/elasticsearch"
	"github.com/fjogeleit/policy-reporter/pkg/target/loki"
	"github.com/fjogeleit/policy-reporter/pkg/target/slack"
)

// Resolver manages dependencies
type Resolver struct {
	config                     *Config
	kubeClient                 report.Client
	lokiClient                 target.Client
	elasticsearchClient        target.Client
	slackClient                target.Client
	policyReportMetrics        metrics.Metrics
	clusterPolicyReportMetrics metrics.Metrics
}

// PolicyReportClient resolver method
func (r *Resolver) PolicyReportClient() (report.Client, error) {
	if r.kubeClient != nil {
		return r.kubeClient, nil
	}

	client, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		r.config.Kubeconfig,
		r.config.Namespace,
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

	return r.elasticsearchClient
}

// ElasticsearchClient resolver method
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

	return r.slackClient
}

// PolicyReportMetrics resolver method
func (r *Resolver) PolicyReportMetrics() (metrics.Metrics, error) {
	if r.policyReportMetrics != nil {
		return r.policyReportMetrics, nil
	}

	client, err := r.PolicyReportClient()
	if err != nil {
		return nil, err
	}

	r.policyReportMetrics = metrics.NewPolicyReportMetrics(client)

	return r.policyReportMetrics, nil
}

// ClusterPolicyReportMetrics resolver method
func (r *Resolver) ClusterPolicyReportMetrics() (metrics.Metrics, error) {
	if r.clusterPolicyReportMetrics != nil {
		return r.clusterPolicyReportMetrics, nil
	}

	client, err := r.PolicyReportClient()
	if err != nil {
		return nil, err
	}

	r.clusterPolicyReportMetrics = metrics.NewClusterPolicyMetrics(client)

	return r.clusterPolicyReportMetrics, nil
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

// Reset all cached dependencies
func (r *Resolver) Reset() {
	r.kubeClient = nil
	r.lokiClient = nil
	r.elasticsearchClient = nil
	r.policyReportMetrics = nil
	r.clusterPolicyReportMetrics = nil
}

// NewResolver constructor function
func NewResolver(config *Config) Resolver {
	return Resolver{
		config: config,
	}
}
