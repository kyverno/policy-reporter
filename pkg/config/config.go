package config

// Loki configuration
type Loki struct {
	Host            string `mapstructure:"host"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string `mapstructure:"minimumPriority"`
}

// Elasticsearch configuration
type Elasticsearch struct {
	Host            string `mapstructure:"host"`
	Index           string `mapstructure:"index"`
	Rotation        string `mapstructure:"rotation"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string `mapstructure:"minimumPriority"`
}

// Slack configuration
type Slack struct {
	Webhook         string `mapstructure:"webhook"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string `mapstructure:"minimumPriority"`
}

// Discord configuration
type Discord struct {
	Webhook         string `mapstructure:"webhook"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string `mapstructure:"minimumPriority"`
}

// Config of the PolicyReporter
type Config struct {
	Loki          Loki          `mapstructure:"loki"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Slack         Slack         `mapstructure:"slack"`
	Discord       Discord       `mapstructure:"discord"`
	Kubeconfig    string        `mapstructure:"kubeconfig"`
	Namespace     string        `mapstructure:"namespace"`
}
