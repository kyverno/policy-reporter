package config

type NamespaceFilter struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

type PriorityFilter struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

type PolicyFilter struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

type Filter struct {
	Namespaces NamespaceFilter `mapstructure:"namespaces"`
	Priorities PriorityFilter  `mapstructure:"priorities"`
	Policies   PolicyFilter    `mapstructure:"policies"`
}

// Loki configuration
type Loki struct {
	Host            string            `mapstructure:"host"`
	CustomLabels    map[string]string `mapstructure:"customLabels"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string            `mapstructure:"minimumPriority"`
	Filter          Filter            `mapstructure:"filter"`
	Sources         []string          `mapstructure:"sources"`
	Channels        []Loki            `mapstructure:"channels"`
}

// Elasticsearch configuration
type Elasticsearch struct {
	Host            string          `mapstructure:"host"`
	Index           string          `mapstructure:"index"`
	Rotation        string          `mapstructure:"rotation"`
	SkipExisting    bool            `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string          `mapstructure:"minimumPriority"`
	Filter          Filter          `mapstructure:"filter"`
	Sources         []string        `mapstructure:"sources"`
	Channels        []Elasticsearch `mapstructure:"channels"`
}

// Slack configuration
type Slack struct {
	Webhook         string   `mapstructure:"webhook"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Filter          Filter   `mapstructure:"filter"`
	Sources         []string `mapstructure:"sources"`
	Channels        []Slack  `mapstructure:"channels"`
}

// Discord configuration
type Discord struct {
	Webhook         string    `mapstructure:"webhook"`
	SkipExisting    bool      `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string    `mapstructure:"minimumPriority"`
	Filter          Filter    `mapstructure:"filter"`
	Sources         []string  `mapstructure:"sources"`
	Channels        []Discord `mapstructure:"channels"`
}

// Teams configuration
type Teams struct {
	Webhook         string   `mapstructure:"webhook"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Filter          Filter   `mapstructure:"filter"`
	Sources         []string `mapstructure:"sources"`
	Channels        []Teams  `mapstructure:"channels"`
}

// UI configuration
type UI struct {
	Host            string   `mapstructure:"host"`
	SkipExisting    bool     `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string   `mapstructure:"minimumPriority"`
	Sources         []string `mapstructure:"sources"`
}

// Webhook configuration
type Webhook struct {
	Host            string            `mapstructure:"host"`
	Headers         map[string]string `mapstructure:"headers"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string            `mapstructure:"minimumPriority"`
	Filter          Filter            `mapstructure:"filter"`
	Sources         []string          `mapstructure:"sources"`
	Channels        []Webhook         `mapstructure:"channels"`
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
	Filter          Filter   `mapstructure:"filter"`
	Sources         []string `mapstructure:"sources"`
	Channels        []S3     `mapstructure:"channels"`
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

// PriorityMap configuration
type PriorityMap = map[string]string

// ClusterReportFilter configuration
type ClusterReportFilter struct {
	Disabled bool `mapstructure:"disabled"`
}

// ReportFilter configuration
type ReportFilter struct {
	Namespaces     NamespaceFilter     `mapstructure:"namespaces"`
	ClusterReports ClusterReportFilter `mapstructure:"clusterReports"`
}

// Redis configuration
type Redis struct {
	Enabled  bool   `mapstructure:"enabled"`
	Address  string `mapstructure:"address"`
	Prefix   string `mapstructure:"prefix"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// Config of the PolicyReporter
type Config struct {
	Loki          Loki          `mapstructure:"loki"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Slack         Slack         `mapstructure:"slack"`
	Discord       Discord       `mapstructure:"discord"`
	Teams         Teams         `mapstructure:"teams"`
	S3            S3            `mapstructure:"s3"`
	UI            UI            `mapstructure:"ui"`
	Webhook       Webhook       `mapstructure:"webhook"`
	API           API           `mapstructure:"api"`
	Kubeconfig    string        `mapstructure:"kubeconfig"`
	DBFile        string        `mapstructure:"dbfile"`
	Metrics       Metrics       `mapstructure:"metrics"`
	REST          REST          `mapstructure:"rest"`
	PriorityMap   PriorityMap   `mapstructure:"priorityMap"`
	ReportFilter  ReportFilter  `mapstructure:"reportFilter"`
	Redis         Redis         `mapstructure:"redis"`
}
