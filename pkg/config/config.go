package config

import "github.com/fjogeleit/policy-reporter/pkg/report"

type Config struct {
	Loki struct {
		Host string `mapstructure:"host"`
	} `mapstructure:"loki"`
	Kubeconfig       string                     `mapstructure:"kubeconfig"`
	PolicyPriorities map[string]report.Priority `mapstructure:"policy_priorities"`
}
