package config

type ValueFilter struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

type EmailReportFilter struct {
	Namespaces ValueFilter `mapstructure:"namespaces"`
	Sources    []string    `mapstructure:"sources"`
}

type TargetFilter struct {
	Namespaces ValueFilter `mapstructure:"namespaces"`
	Priorities ValueFilter `mapstructure:"priorities"`
	Policies   ValueFilter `mapstructure:"policies"`
}

type MetricsFilter struct {
	Namespaces ValueFilter `mapstructure:"namespaces"`
	Policies   ValueFilter `mapstructure:"policies"`
	Severities ValueFilter `mapstructure:"severities"`
	Status     ValueFilter `mapstructure:"status"`
	Sources    ValueFilter `mapstructure:"sources"`
}

// Loki configuration
type Loki struct {
	Name            string            `mapstructure:"name"`
	Host            string            `mapstructure:"host"`
	CustomLabels    map[string]string `mapstructure:"customLabels"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string            `mapstructure:"minimumPriority"`
	Filter          TargetFilter      `mapstructure:"filter"`
	Sources         []string          `mapstructure:"sources"`
	Channels        []Loki            `mapstructure:"channels"`
}

// Elasticsearch configuration
type Elasticsearch struct {
	Name            string          `mapstructure:"name"`
	Host            string          `mapstructure:"host"`
	Index           string          `mapstructure:"index"`
	Rotation        string          `mapstructure:"rotation"`
	SkipExisting    bool            `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string          `mapstructure:"minimumPriority"`
	Filter          TargetFilter    `mapstructure:"filter"`
	Sources         []string        `mapstructure:"sources"`
	Channels        []Elasticsearch `mapstructure:"channels"`
}

// Slack configuration
type Slack struct {
	Name            string       `mapstructure:"name"`
	Webhook         string       `mapstructure:"webhook"`
	SkipExisting    bool         `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string       `mapstructure:"minimumPriority"`
	Filter          TargetFilter `mapstructure:"filter"`
	Sources         []string     `mapstructure:"sources"`
	Channels        []Slack      `mapstructure:"channels"`
}

// Discord configuration
type Discord struct {
	Name            string       `mapstructure:"name"`
	Webhook         string       `mapstructure:"webhook"`
	SkipExisting    bool         `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string       `mapstructure:"minimumPriority"`
	Filter          TargetFilter `mapstructure:"filter"`
	Sources         []string     `mapstructure:"sources"`
	Channels        []Discord    `mapstructure:"channels"`
}

// Teams configuration
type Teams struct {
	Name            string       `mapstructure:"name"`
	Webhook         string       `mapstructure:"webhook"`
	SkipExisting    bool         `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string       `mapstructure:"minimumPriority"`
	Filter          TargetFilter `mapstructure:"filter"`
	Sources         []string     `mapstructure:"sources"`
	Channels        []Teams      `mapstructure:"channels"`
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
	Name            string            `mapstructure:"name"`
	Host            string            `mapstructure:"host"`
	Headers         map[string]string `mapstructure:"headers"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string            `mapstructure:"minimumPriority"`
	Filter          TargetFilter      `mapstructure:"filter"`
	Sources         []string          `mapstructure:"sources"`
	Channels        []Webhook         `mapstructure:"channels"`
}

// S3 configuration
type S3 struct {
	Name            string       `mapstructure:"name"`
	AccessKeyID     string       `mapstructure:"accessKeyID"`
	SecretAccessKey string       `mapstructure:"secretAccessKey"`
	Region          string       `mapstructure:"region"`
	Endpoint        string       `mapstructure:"endpoint"`
	Prefix          string       `mapstructure:"prefix"`
	Bucket          string       `mapstructure:"bucket"`
	SkipExisting    bool         `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string       `mapstructure:"minimumPriority"`
	Filter          TargetFilter `mapstructure:"filter"`
	Sources         []string     `mapstructure:"sources"`
	Channels        []S3         `mapstructure:"channels"`
}

// Kinesis configuration
type Kinesis struct {
	Name            string       `mapstructure:"name"`
	AccessKeyID     string       `mapstructure:"accessKeyID"`
	SecretAccessKey string       `mapstructure:"secretAccessKey"`
	Region          string       `mapstructure:"region"`
	Endpoint        string       `mapstructure:"endpoint"`
	StreamName      string       `mapstructure:"streamName"`
	SkipExisting    bool         `mapstructure:"skipExistingOnStartup"`
	MinimumPriority string       `mapstructure:"minimumPriority"`
	Filter          TargetFilter `mapstructure:"filter"`
	Sources         []string     `mapstructure:"sources"`
	Channels        []Kinesis    `mapstructure:"channels"`
}

// SMTP configuration
type SMTP struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	From       string `mapstructure:"from"`
	Encryption string `mapstructure:"encryption"`
}

// EmailReport configuration
type EmailReport struct {
	To                    []string          `mapstructure:"to"`
	Format                string            `mapstructure:"format"`
	Filter                EmailReportFilter `mapstructure:"filter"`
	Channels              []EmailReport     `mapstructure:"channels"`
	DisableClusterReports bool              `mapstructure:"disableClusterReports"`
}

// EmailReport configuration
type EmailTemplates struct {
	Dir string `mapstructure:"dir"`
}

// EmailReports configuration
type EmailReports struct {
	SMTP        SMTP           `mapstructure:"smtp"`
	Templates   EmailTemplates `mapstructure:"templates"`
	Summary     EmailReport    `mapstructure:"summary"`
	Violations  EmailReport    `mapstructure:"violations"`
	ClusterName string         `mapstructure:"clusterName"`
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
	Enabled bool          `mapstructure:"enabled"`
	Filter  MetricsFilter `mapstructure:"filter"`
}

// Profiling configuration
type Profiling struct {
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
	Namespaces     ValueFilter         `mapstructure:"namespaces"`
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
	Kinesis       Kinesis       `mapstructure:"kinesis"`
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
	Profiling     Profiling     `mapstructure:"profiling"`
	EmailReports  EmailReports  `mapstructure:"emailReports"`
}
