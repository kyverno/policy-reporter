package config

// Loki configuration
type Loki struct {
	Host            string   `mapstructure:"host"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// Elasticsearch configuration
type Elasticsearch struct {
	Host            string   `mapstructure:"host"`
	Index           string   `mapstructure:"index"`
	Rotation        string   `mapstructure:"rotation"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// Slack configuration
type Slack struct {
	Webhook         string   `mapstructure:"webhook"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// Discord configuration
type Discord struct {
	Webhook         string   `mapstructure:"webhook"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// Teams configuration
type Teams struct {
	Webhook         string   `mapstructure:"webhook"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// UI configuration
type UI struct {
	Host            string   `mapstructure:"host"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

type S3 struct {
	AccessKeyID     string   `mapstructure:"accessKeyID"`
	SecretAccessKey string   `mapstructure:"secretAccessKey"`
	Region          string   `mapstructure:"region"`
	Endpoint        string   `mapstructure:"endpoint"`
	Prefix          string   `mapstructure:"prefix"`
	Bucket          string   `mapstructure:"bucket"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// API configuration
type API struct {
	Port int `mapstructure:"port"`
}

// REST configuration
type REST struct {
	Enabled bool `mapstructure:"enabled"`
}

// Metrics configuration
type Metrics struct {
	Enabled bool `mapstructure:"enabled"`
}

type PriorityMap = map[string]string

// Config of the PolicyReporter
type Config struct {
	Loki          Loki          `mapstructure:"loki"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Slack         Slack         `mapstructure:"slack"`
	Discord       Discord       `mapstructure:"discord"`
	Teams         Teams         `mapstructure:"teams"`
	S3            S3            `mapstructure:"s3"`
	UI            UI            `mapstructure:"ui"`
	API           API           `mapstructure:"api"`
	Kubeconfig    string        `mapstructure:"kubeconfig"`
	DBFile        string        `mapstructure:"dbfile"`
	Metrics       Metrics       `mapstructure:"metrics"`
	REST          REST          `mapstructure:"rest"`
	PriorityMap   PriorityMap   `mapstructure:"priorityMap"`
}
