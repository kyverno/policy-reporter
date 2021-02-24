package config

import (
	"context"
	"net/http"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/metrics"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/loki"
)

var (
	kubeClient                 report.Client
	lokiClient                 target.Client
	policyReportMetrics        metrics.Metrics
	clusterPolicyReportMetrics metrics.Metrics
)

// Resolver manages dependencies
type Resolver struct {
	config *Config
}

// PolicyReportClient resolver method
func (r *Resolver) PolicyReportClient() (report.Client, error) {
	if kubeClient != nil {
		return kubeClient, nil
	}

	client, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		r.config.Kubeconfig,
		r.config.Namespace,
		time.Now(),
	)

	kubeClient = client

	return client, err
}

// LokiClient resolver method
func (r *Resolver) LokiClient() target.Client {
	if lokiClient != nil {
		return lokiClient
	}

	if r.config.Loki.Host == "" {
		return nil
	}

	lokiClient = loki.NewClient(
		r.config.Loki.Host,
		r.config.Loki.MinimumPriority,
		&http.Client{},
	)

	return lokiClient
}

// PolicyReportMetrics resolver method
func (r *Resolver) PolicyReportMetrics() (metrics.Metrics, error) {
	if policyReportMetrics != nil {
		return policyReportMetrics, nil
	}

	client, err := r.PolicyReportClient()
	if err != nil {
		return nil, err
	}

	policyReportMetrics = metrics.NewPolicyReportMetrics(client)

	return policyReportMetrics, nil
}

// ClusterPolicyReportMetrics resolver method
func (r *Resolver) ClusterPolicyReportMetrics() (metrics.Metrics, error) {
	if clusterPolicyReportMetrics != nil {
		return clusterPolicyReportMetrics, nil
	}

	client, err := r.PolicyReportClient()
	if err != nil {
		return nil, err
	}

	clusterPolicyReportMetrics = metrics.NewClusterPolicyMetrics(client)

	return clusterPolicyReportMetrics, nil
}

// Reset all cached dependencies
func (r *Resolver) Reset() {
	kubeClient = nil
	lokiClient = nil
	policyReportMetrics = nil
	clusterPolicyReportMetrics = nil
}

// NewResolver constructor function
func NewResolver(config *Config) Resolver {
	return Resolver{config}
}
