package config

import (
	"context"
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

type Resolver struct {
	config *Config
}

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
	)

	return lokiClient
}

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

func NewResolver(config *Config) Resolver {
	return Resolver{config}
}
