package config

// Config of the PolicyReporter
type Config struct {
	Loki struct {
		Host            string `mapstructure:"host"`
		SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
		MinimumPriority string `mapstructure:"minimumPriority"`
	} `mapstructure:"loki"`
	Kubeconfig string `mapstructure:"kubeconfig"`
	Namespace  string `mapstructure:"namespace"`
}
