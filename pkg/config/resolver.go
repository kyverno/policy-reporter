package config

import (
	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/loki"
)

var (
	kubeClient kubernetes.Client
	lokiClient target.Client
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

func NewResolver(config *Config) Resolver {
	return Resolver{config}
}
