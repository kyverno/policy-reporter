package config

// Config of the PolicyReporter
type Config struct {
	Loki struct {
		Host            string `mapstructure:"host"`
		SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
		MinimumPriority string `mapstructure:"minimumPriority"`
	} `mapstructure:"loki"`
	Elasticsearch struct {
		Host            string `mapstructure:"host"`
		Index           string `mapstructure:"index"`
		Rotation        string `mapstructure:"rotation"`
		SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
		MinimumPriority string `mapstructure:"minimumPriority"`
	} `mapstructure:"elasticsearch"`
	Kubeconfig string `mapstructure:"kubeconfig"`
	Namespace  string `mapstructure:"namespace"`
}
