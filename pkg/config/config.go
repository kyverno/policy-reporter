package config

type Config struct {
	Loki struct {
		Host            string `mapstructure:"host"`
		SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
		MinimumPriority string `mapstructure:"minimumPriority"`
	} `mapstructure:"loki"`
	Kubeconfig       string            `mapstructure:"kubeconfig"`
	PolicyPriorities map[string]string `mapstructure:"policy_priorities"`
}
