package config

import "github.com/kyverno/policy-reporter/pkg/target"

type ValueFilter struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

type EmailReportFilter struct {
	DisableClusterReports bool        `mapstructure:"disableClusterReports"`
	Namespaces            ValueFilter `mapstructure:"namespaces"`
	Sources               ValueFilter `mapstructure:"sources"`
}

type TargetFilter struct {
	Namespaces   ValueFilter `mapstructure:"namespaces"`
	Priorities   ValueFilter `mapstructure:"priorities"`
	Policies     ValueFilter `mapstructure:"policies"`
	ReportLabels ValueFilter `mapstructure:"reportLabels"`
}

type MetricsFilter struct {
	Namespaces ValueFilter `mapstructure:"namespaces"`
	Policies   ValueFilter `mapstructure:"policies"`
	Severities ValueFilter `mapstructure:"severities"`
	Status     ValueFilter `mapstructure:"status"`
	Sources    ValueFilter `mapstructure:"sources"`
}

type TargetBaseOptions struct {
	Name            string            `mapstructure:"name"`
	MinimumPriority string            `mapstructure:"minimumPriority"`
	Filter          TargetFilter      `mapstructure:"filter"`
	SecretRef       string            `mapstructure:"secretRef"`
	MountedSecret   string            `mapstructure:"mountedSecret"`
	Sources         []string          `mapstructure:"sources"`
	CustomFields    map[string]string `mapstructure:"customFields"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
}

func (config *TargetBaseOptions) MapBaseParent(parent TargetBaseOptions) {
	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}
}

func (config *TargetBaseOptions) ClientOptions() target.ClientOptions {
	return target.ClientOptions{
		Name:                  config.Name,
		SkipExistingOnStartup: config.SkipExisting,
		ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
		ReportFilter:          createReportFilter(config.Filter),
	}
}

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"accessKeyID"`
	SecretAccessKey string `mapstructure:"secretAccessKey"`
	Region          string `mapstructure:"region"`
	Endpoint        string `mapstructure:"endpoint"`
}

func (config *AWSConfig) MapAWSParent(parent AWSConfig) {
	if config.Endpoint == "" {
		config.Endpoint = parent.Endpoint
	}

	if config.AccessKeyID == "" {
		config.AccessKeyID = parent.AccessKeyID
	}

	if config.SecretAccessKey == "" {
		config.SecretAccessKey = parent.SecretAccessKey
	}

	if config.Region == "" {
		config.Region = parent.Region
	}
}

type TargetOption interface {
	BaseOptions() *TargetBaseOptions
}

// Loki configuration
type Loki struct {
	TargetBaseOptions `mapstructure:",squash"`
	CustomLabels      map[string]string `mapstructure:"customLabels"`
	Host              string            `mapstructure:"host"`
	SkipTLS           bool              `mapstructure:"skipTLS"`
	Certificate       string            `mapstructure:"certificate"`
	Path              string            `mapstructure:"path"`
	Channels          []*Loki           `mapstructure:"channels"`
}

// Elasticsearch configuration
type Elasticsearch struct {
	TargetBaseOptions `mapstructure:",squash"`
	Host              string           `mapstructure:"host"`
	SkipTLS           bool             `mapstructure:"skipTLS"`
	Certificate       string           `mapstructure:"certificate"`
	Index             string           `mapstructure:"index"`
	Rotation          string           `mapstructure:"rotation"`
	Username          string           `mapstructure:"username"`
	Password          string           `mapstructure:"password"`
	Channels          []*Elasticsearch `mapstructure:"channels"`
}

// Slack configuration
type Slack struct {
	TargetBaseOptions `mapstructure:",squash"`
	Webhook           string   `mapstructure:"webhook"`
	Channel           string   `mapstructure:"channel"`
	Channels          []*Slack `mapstructure:"channels"`
}

// Discord configuration
type Discord struct {
	TargetBaseOptions `mapstructure:",squash"`
	Webhook           string     `mapstructure:"webhook"`
	Channels          []*Discord `mapstructure:"channels"`
}

// Teams configuration
type Teams struct {
	TargetBaseOptions `mapstructure:",squash"`
	Webhook           string   `mapstructure:"webhook"`
	SkipTLS           bool     `mapstructure:"skipTLS"`
	Certificate       string   `mapstructure:"certificate"`
	Channels          []*Teams `mapstructure:"channels"`
}

// UI configuration
type UI struct {
	TargetBaseOptions `mapstructure:",squash"`
	Host              string `mapstructure:"host"`
	SkipTLS           bool   `mapstructure:"skipTLS"`
	Certificate       string `mapstructure:"certificate"`
}

// Webhook configuration
type Webhook struct {
	TargetBaseOptions `mapstructure:",squash"`
	Host              string            `mapstructure:"host"`
	SkipTLS           bool              `mapstructure:"skipTLS"`
	Certificate       string            `mapstructure:"certificate"`
	Headers           map[string]string `mapstructure:"headers"`
	Channels          []*Webhook        `mapstructure:"channels"`
}

// Telegram configuration
type Telegram struct {
	TargetBaseOptions `mapstructure:",squash"`
	Host              string            `mapstructure:"host"`
	Token             string            `mapstructure:"token"`
	ChatID            string            `mapstructure:"chatID"`
	SkipTLS           bool              `mapstructure:"skipTLS"`
	Certificate       string            `mapstructure:"certificate"`
	Headers           map[string]string `mapstructure:"headers"`
	Channels          []*Telegram       `mapstructure:"channels"`
}

// GoogleChat configuration
type GoogleChat struct {
	TargetBaseOptions `mapstructure:",squash"`
	Webhook           string            `mapstructure:"webhook"`
	SkipTLS           bool              `mapstructure:"skipTLS"`
	Certificate       string            `mapstructure:"certificate"`
	Headers           map[string]string `mapstructure:"headers"`
	Channels          []*GoogleChat     `mapstructure:"channels"`
}

// S3 configuration
type S3 struct {
	TargetBaseOptions    `mapstructure:",squash"`
	AWSConfig            `mapstructure:",squash"`
	Prefix               string `mapstructure:"prefix"`
	Bucket               string `mapstructure:"bucket"`
	BucketKeyEnabled     bool   `mapstructure:"bucketKeyEnabled"`
	KmsKeyID             string `mapstructure:"kmsKeyId"`
	ServerSideEncryption string `mapstructure:"serverSideEncryption"`
	PathStyle            bool   `mapstructure:"pathStyle"`
	Channels             []*S3  `mapstructure:"channels"`
}

// Kinesis configuration
type Kinesis struct {
	TargetBaseOptions `mapstructure:",squash"`
	AWSConfig         `mapstructure:",squash"`
	StreamName        string     `mapstructure:"streamName"`
	Channels          []*Kinesis `mapstructure:"channels"`
}

// SecurityHub configuration
type SecurityHub struct {
	TargetBaseOptions `mapstructure:",squash"`
	AWSConfig         `mapstructure:",squash"`
	AccountID         string         `mapstructure:"accountId"`
	Channels          []*SecurityHub `mapstructure:"channels"`
}

// GCS configuration
type GCS struct {
	TargetBaseOptions `mapstructure:",squash"`
	Credentials       string   `mapstructure:"credentials"`
	Prefix            string   `mapstructure:"prefix"`
	Bucket            string   `mapstructure:"bucket"`
	Sources           []string `mapstructure:"sources"`
	Channels          []*GCS   `mapstructure:"channels"`
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
	To       []string          `mapstructure:"to"`
	Format   string            `mapstructure:"format"`
	Filter   EmailReportFilter `mapstructure:"filter"`
	Channels []EmailReport     `mapstructure:"channels"`
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
	TitlePrefix string         `mapstructure:"titlePrefix"`
}

// BasicAuth configuration
type BasicAuth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// API configuration
type API struct {
	Port      int       `mapstructure:"port"`
	Logging   bool      `mapstructure:"logging"`
	BasicAuth BasicAuth `mapstructure:"basicAuth"`
}

// REST configuration
type REST struct {
	Enabled bool `mapstructure:"enabled"`
}

// Metrics configuration
type Metrics struct {
	Filter       MetricsFilter `mapstructure:"filter"`
	CustomLabels []string      `mapstructure:"customLabels"`
	Mode         string        `mapstructure:"mode"`
	Enabled      bool          `mapstructure:"enabled"`
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

// LeaderElection configuration
type LeaderElection struct {
	LockName        string `mapstructure:"lockName"`
	PodName         string `mapstructure:"podName"`
	Namespace       string `mapstructure:"namespace"`
	LeaseDuration   int    `mapstructure:"leaseDuration"`
	RenewDeadline   int    `mapstructure:"renewDeadline"`
	RetryPeriod     int    `mapstructure:"retryPeriod"`
	ReleaseOnCancel bool   `mapstructure:"releaseOnCancel"`
	Enabled         bool   `mapstructure:"enabled"`
}

// K8sClient config struct
type K8sClient struct {
	QPS        float32 `mapstructure:"qps"`
	Burst      int     `mapstructure:"burst"`
	Kubeconfig string  `mapstructure:"kubeconfig"`
}

type Logging struct {
	LogLevel    int8   `mapstructure:"logLevel"`
	Encoding    string `mapstructure:"encoding"`
	Development bool   `mapstructure:"development"`
}

type Database struct {
	Type          string `mapstructure:"type"`
	DSN           string `mapstructure:"dsn"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	Database      string `mapstructure:"database"`
	Host          string `mapstructure:"host"`
	EnableSSL     bool   `mapstructure:"enableSSL"`
	SecretRef     string `mapstructure:"secretRef"`
	MountedSecret string `mapstructure:"mountedSecret"`
}

// Config of the PolicyReporter
type Config struct {
	Version        string
	Namespace      string         `mapstructure:"namespace"`
	Loki           *Loki          `mapstructure:"loki"`
	Elasticsearch  *Elasticsearch `mapstructure:"elasticsearch"`
	Slack          *Slack         `mapstructure:"slack"`
	Discord        *Discord       `mapstructure:"discord"`
	Teams          *Teams         `mapstructure:"teams"`
	S3             *S3            `mapstructure:"s3"`
	Kinesis        *Kinesis       `mapstructure:"kinesis"`
	SecurityHub    *SecurityHub   `mapstructure:"securityHub"`
	GCS            *GCS           `mapstructure:"gcs"`
	UI             *UI            `mapstructure:"ui"`
	Webhook        *Webhook       `mapstructure:"webhook"`
	Telegram       *Telegram      `mapstructure:"telegram"`
	GoogleChat     *GoogleChat    `mapstructure:"googleChat"`
	API            API            `mapstructure:"api"`
	WorkerCount    int            `mapstructure:"worker"`
	DBFile         string         `mapstructure:"dbfile"`
	Metrics        Metrics        `mapstructure:"metrics"`
	REST           REST           `mapstructure:"rest"`
	PriorityMap    PriorityMap    `mapstructure:"priorityMap"`
	ReportFilter   ReportFilter   `mapstructure:"reportFilter"`
	Redis          Redis          `mapstructure:"redis"`
	Profiling      Profiling      `mapstructure:"profiling"`
	EmailReports   EmailReports   `mapstructure:"emailReports"`
	LeaderElection LeaderElection `mapstructure:"leaderElection"`
	K8sClient      K8sClient      `mapstructure:"k8sClient"`
	Logging        Logging        `mapstructure:"logging"`
	Database       Database       `mapstructure:"database"`
}
