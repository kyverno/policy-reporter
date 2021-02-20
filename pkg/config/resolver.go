package config

import (
	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/metrics"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/loki"
)

var (
	kubeClient                 kubernetes.Client
	lokiClient                 target.Client
	policyReportMetrics        metrics.Metrics
	clusterPolicyReportMetrics metrics.Metrics
)

type Resolver struct {
	config *Config
}

func (r *Resolver) KubernetesClient() (kubernetes.Client, error) {
	if kubeClient != nil {
		return kubeClient, nil
	}

	return kubernetes.NewDynamicClient(r.config.Kubeconfig, r.config.PolicyPriorities)
}

func (r *Resolver) LokiClient() target.Client {
	if kubeClient != nil {
		return lokiClient
	}

	if r.config.Loki.Host == "" {
		return nil
	}

	return loki.NewClient(r.config.Loki.Host)
}

func (r *Resolver) PolicyReportMetrics() (metrics.Metrics, error) {
	if policyReportMetrics != nil {
		return policyReportMetrics, nil
	}

	client, err := r.KubernetesClient()
	if err != nil {
		return nil, err
	}

	return metrics.NewPolicyReportMetrics(client), nil
}

func (r *Resolver) ClusterPolicyReportMetrics() (metrics.Metrics, error) {
	if policyReportMetrics != nil {
		return clusterPolicyReportMetrics, nil
	}

	client, err := r.KubernetesClient()
	if err != nil {
		return nil, err
	}

	return metrics.NewClusterPolicyMetrics(client), nil
}

func NewResolver(config *Config) Resolver {
	return Resolver{config}
}
