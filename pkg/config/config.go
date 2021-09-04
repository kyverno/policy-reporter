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

// Teams configuration
type Teams struct {
	Webhook         string `mapstructure:"webhook"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string `mapstructure:"minimumPriority"`
}

// UI configuration
type UI struct {
	Host            string `mapstructure:"host"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string `mapstructure:"minimumPriority"`
}

type YandexS3 struct {
	AccessKeyID     string `mapstructure:"accessKeyID"`
	SecretAccessKey string `mapstructure:"SecretAccessKey"`
	SkipExisting    bool   `mapstructure:"skipExistingOnStartup"`
	Region          string `mapstructure:"Region"`
	Endpoint        string `mapstructure:"Endpoint"`
	Prefix          string `mapstructure:"Prefix"`
	Bucket          string `mapstructure:"Bucket"`
	MinimumPriority string `mapstructure:"MinimumPriority"`
}

// API configuration
type API struct {
	Port int `mapstructure:"port"`
}

// Config of the PolicyReporter
type Config struct {
	Loki          Loki          `mapstructure:"loki"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Slack         Slack         `mapstructure:"slack"`
	Discord       Discord       `mapstructure:"discord"`
	Teams         Teams         `mapstructure:"teams"`
	YandexS3      YandexS3      `mapstructure:"yandexS3"`
	UI            UI            `mapstructure:"ui"`
	API           API           `mapstructure:"api"`
	Kubeconfig    string        `mapstructure:"kubeconfig"`
	Namespace     string        `mapstructure:"namespace"`
}
