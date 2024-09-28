package config

import "github.com/kyverno/policy-reporter/pkg/target"

type ValueFilter struct {
	Include []string `mapstructure:"include" json:"include,omitempty"`
	Exclude []string `mapstructure:"exclude" json:"exclude,omitempty"`
}

type EmailReportFilter struct {
	DisableClusterReports bool        `mapstructure:"disableClusterReports" json:"disableClusterReports,omitempty"`
	Namespaces            ValueFilter `mapstructure:"namespaces" json:"namespaces,omitempty"`
	Sources               ValueFilter `mapstructure:"sources" json:"sources,omitempty"`
}

type TargetFilter struct {
	Namespaces   ValueFilter `mapstructure:"namespaces" json:"namespaces,omitempty"`
	Priorities   ValueFilter `mapstructure:"priorities" json:"priorities,omitempty"`
	Policies     ValueFilter `mapstructure:"policies" json:"policies,omitempty"`
	ReportLabels ValueFilter `mapstructure:"reportLabels" json:"reportLabels,omitempty"`
}

type MetricsFilter struct {
	Namespaces ValueFilter `mapstructure:"namespaces" json:"namespaces,omitempty"`
	Policies   ValueFilter `mapstructure:"policies" json:"policies,omitempty"`
	Severities ValueFilter `mapstructure:"severities" json:"severities,omitempty"`
	Status     ValueFilter `mapstructure:"status" json:"status,omitempty"`
	Sources    ValueFilter `mapstructure:"sources" json:"sources,omitempty"`
	Kinds      ValueFilter `mapstructure:"kinds" json:"kinds,omitempty"`
}

type TargetBaseOptions struct {
	Name            string            `mapstructure:"name" json:"name,omitempty"`
	MinimumPriority string            `mapstructure:"minimumPriority" json:"minimumPriority,omitempty"`
	Filter          TargetFilter      `mapstructure:"filter" json:"filter,omitempty"`
	SecretRef       string            `mapstructure:"secretRef" json:"secretRef,omitempty"`
	MountedSecret   string            `mapstructure:"mountedSecret" json:"mountedSecret,omitempty"`
	Sources         []string          `mapstructure:"sources" json:"sources,omitempty"`
	CustomFields    map[string]string `mapstructure:"customFields" json:"customFields,omitempty"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup" json:"skipExistingOnStartup,omitempty"`
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
	AccessKeyID     string `mapstructure:"accessKeyID" json:"accessKeyID,omitempty"`
	SecretAccessKey string `mapstructure:"secretAccessKey" json:"secretAccessKey,omitempty"`
	Region          string `mapstructure:"region" json:"region,omitempty"`
	Endpoint        string `mapstructure:"endpoint" json:"endpoint,omitempty"`
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
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	CustomLabels      map[string]string `mapstructure:"customLabels" json:"customLabels,omitempty"`
	Headers           map[string]string `mapstructure:"headers" json:"headers,omitempty"`
	Host              string            `mapstructure:"host" json:"host,omitempty"`
	SkipTLS           bool              `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate       string            `mapstructure:"certificate" json:"certificate,omitempty"`
	Path              string            `mapstructure:"path" json:"path,omitempty"`
	Channels          []*Loki           `mapstructure:"channels" json:"channels,omitempty"`
	Username          string            `mapstructure:"username" json:"username,omitempty"`
	Password          string            `mapstructure:"password" json:"password,omitempty"`
}

// Elasticsearch configuration
type Elasticsearch struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Host              string           `mapstructure:"host" json:"host,omitempty"`
	SkipTLS           bool             `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate       string           `mapstructure:"certificate" json:"certificate,omitempty"`
	Index             string           `mapstructure:"index" json:"index,omitempty"`
	Rotation          string           `mapstructure:"rotation" json:"rotation,omitempty"`
	Username          string           `mapstructure:"username" json:"username,omitempty"`
	Password          string           `mapstructure:"password" json:"password,omitempty"`
	APIKey            string           `mapstructure:"apiKey" json:"apiKey,omitempty"`
	Channels          []*Elasticsearch `mapstructure:"channels" json:"channels,omitempty"`
	TypelessAPI       bool             `mapstructure:"typelessApi" json:"typelessApi,omitempty"`
}

// Slack configuration
type Slack struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Webhook           string   `mapstructure:"webhook" json:"webhook,omitempty"`
	Channel           string   `mapstructure:"channel" json:"channel,omitempty"`
	Channels          []*Slack `mapstructure:"channels" json:"channels,omitempty"`
}

// Discord configuration
type Discord struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Webhook           string     `mapstructure:"webhook" json:"webhook,omitempty"`
	Channels          []*Discord `mapstructure:"channels" json:"channels,omitempty"`
}

// Teams configuration
type Teams struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Webhook           string   `mapstructure:"webhook" json:"webhook,omitempty"`
	SkipTLS           bool     `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate       string   `mapstructure:"certificate" json:"certificate,omitempty"`
	Channels          []*Teams `mapstructure:"channels" json:"channels,omitempty"`
}

// UI configuration
type UI struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Host              string `mapstructure:"host" json:"host,omitempty"`
	SkipTLS           bool   `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate       string `mapstructure:"certificate" json:"certificate,omitempty"`
}

// Webhook configuration
type Webhook struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Host              string            `mapstructure:"host" json:"host,omitempty"`
	SkipTLS           bool              `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate       string            `mapstructure:"certificate" json:"certificate,omitempty"`
	Headers           map[string]string `mapstructure:"headers" json:"headers,omitempty"`
	Channels          []*Webhook        `mapstructure:"channels" json:"channels,omitempty"`
}

// Telegram configuration
type Telegram struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
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
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Webhook           string            `mapstructure:"webhook" json:"webhook,omitempty"`
	SkipTLS           bool              `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate       string            `mapstructure:"certificate" json:"certificate,omitempty"`
	Headers           map[string]string `mapstructure:"headers" json:"headers,omitempty"`
	Channels          []*GoogleChat     `mapstructure:"channels" json:"channels,omitempty"`
}

// S3 configuration
type S3 struct {
	TargetBaseOptions    `mapstructure:",squash" json:",inline"`
	AWSConfig            `mapstructure:",squash" json:",inline"`
	Prefix               string `mapstructure:"prefix" json:"prefix,omitempty"`
	Bucket               string `mapstructure:"bucket" json:"bucket,omitempty"`
	BucketKeyEnabled     bool   `mapstructure:"bucketKeyEnabled" json:"bucketKeyEnabled,omitempty"`
	KmsKeyID             string `mapstructure:"kmsKeyId" json:"kmsKeyId,omitempty"`
	ServerSideEncryption string `mapstructure:"serverSideEncryption" json:"serverSideEncryption,omitempty"`
	PathStyle            bool   `mapstructure:"pathStyle" json:"pathStyle,omitempty"`
	Channels             []*S3  `mapstructure:"channels" json:"channels,omitempty"`
}

// Kinesis configuration
type Kinesis struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	AWSConfig         `mapstructure:",squash" json:",inline"`
	StreamName        string     `mapstructure:"streamName" json:"streamName,omitempty"`
	Channels          []*Kinesis `mapstructure:"channels" json:"channels,omitempty"`
}

// SecurityHub configuration
type SecurityHub struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	AWSConfig         `mapstructure:",squash" json:",inline"`
	AccountID         string         `mapstructure:"accountId" json:"accountId,omitempty"`
	ProductName       string         `mapstructure:"productName" json:"productName,omitempty"`
	CompanyName       string         `mapstructure:"companyName" json:"companyName,omitempty"`
	DelayInSeconds    int            `mapstructure:"delayInSeconds" json:"delayInSeconds,omitempty"`
	Cleanup           bool           `mapstructure:"cleanup" json:"cleanup,omitempty"`
	Channels          []*SecurityHub `mapstructure:"channels" json:"channels,omitempty"`
}

// GCS configuration
type GCS struct {
	TargetBaseOptions `mapstructure:",squash" json:",inline"`
	Credentials       string   `mapstructure:"credentials" json:"customLabels,omitempty"`
	Prefix            string   `mapstructure:"prefix" json:"credentials,omitempty"`
	Bucket            string   `mapstructure:"bucket" json:"bucket,omitempty"`
	Sources           []string `mapstructure:"sources" json:"sources,omitempty"`
	Channels          []*GCS   `mapstructure:"channels" json:"channels,omitempty"`
}

// SMTP configuration
type SMTP struct {
	Host        string `mapstructure:"host" json:"host,omitempty"`
	Port        int    `mapstructure:"port" json:"port,omitempty"`
	Username    string `mapstructure:"username" json:"username,omitempty"`
	Password    string `mapstructure:"password" json:"password,omitempty"`
	From        string `mapstructure:"from" json:"from,omitempty"`
	Encryption  string `mapstructure:"encryption" json:"encryption,omitempty"`
	SkipTLS     bool   `mapstructure:"skipTLS" json:"skipTLS,omitempty"`
	Certificate string `mapstructure:"certificate" json:"certificate,omitempty"`
}

// EmailReport configuration
type EmailReport struct {
	To       []string          `mapstructure:"to" json:"to,omitempty"`
	Format   string            `mapstructure:"format" json:"format,omitempty"`
	Filter   EmailReportFilter `mapstructure:"filter" json:"filter,omitempty"`
	Channels []EmailReport     `mapstructure:"channels" json:"channels,omitempty"`
}

// EmailReport configuration
type Templates struct {
	Dir string `mapstructure:"dir" json:"dir,omitempty"`
}

// EmailReports configuration
type EmailReports struct {
	SMTP        SMTP        `mapstructure:"smtp" json:"smtp,omitempty"`
	Summary     EmailReport `mapstructure:"summary" json:"summary,omitempty"`
	Violations  EmailReport `mapstructure:"violations" json:"violations,omitempty"`
	ClusterName string      `mapstructure:"clusterName" json:"clusterName,omitempty"`
	TitlePrefix string      `mapstructure:"titlePrefix" json:"titlePrefix,omitempty"`
}

// BasicAuth configuration
type BasicAuth struct {
	Username  string `mapstructure:"username" json:"username,omitempty"`
	Password  string `mapstructure:"password" json:"password,omitempty"`
	SecretRef string `mapstructure:"secretRef" json:"secretRef,omitempty"`
}

// API configuration
type API struct {
	Port      int       `mapstructure:"port" json:"port,omitempty"`
	Logging   bool      `mapstructure:"logging" json:"logging,omitempty"`
	BasicAuth BasicAuth `mapstructure:"basicAuth" json:"basicAuth,omitempty"`
}

// REST configuration
type REST struct {
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
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
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
}

// ClusterReportFilter configuration
type ClusterReportFilter struct {
	Disabled bool `mapstructure:"disabled" json:"disabled,omitempty"`
}

// ReportFilter configuration
type ReportFilter struct {
	Namespaces     ValueFilter         `mapstructure:"namespaces" json:"namespaces,omitempty"`
	ClusterReports ClusterReportFilter `mapstructure:"clusterReports" json:"clusterReports,omitempty"`
}

// Redis configuration
type Redis struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled,omitempty"`
	Address  string `mapstructure:"address" json:"address,omitempty"`
	Prefix   string `mapstructure:"prefix" json:"prefix,omitempty"`
	Username string `mapstructure:"username" json:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty"`
	Database int    `mapstructure:"database" json:"database,omitempty"`
}

// LeaderElection configuration
type LeaderElection struct {
	LockName        string `mapstructure:"lockName" json:"lockName,omitempty"`
	PodName         string `mapstructure:"podName" json:"podName,omitempty"`
	Namespace       string `mapstructure:"namespace" json:"namespace,omitempty"`
	LeaseDuration   int    `mapstructure:"leaseDuration" json:"leaseDuration,omitempty"`
	RenewDeadline   int    `mapstructure:"renewDeadline" json:"renewDeadline,omitempty"`
	RetryPeriod     int    `mapstructure:"retryPeriod" json:"retryPeriod,omitempty"`
	ReleaseOnCancel bool   `mapstructure:"releaseOnCancel" json:"releaseOnCancel,omitempty"`
	Enabled         bool   `mapstructure:"enabled" json:"enabled,omitempty"`
}

// K8sClient config struct
type K8sClient struct {
	QPS        float32 `mapstructure:"qps" json:"qps,omitempty"`
	Burst      int     `mapstructure:"burst" json:"burst,omitempty"`
	Kubeconfig string  `mapstructure:"kubeconfig" json:"kubeconfig,omitempty"`
}

type Logging struct {
	LogLevel    int8   `mapstructure:"logLevel" json:"logLevel,omitempty"`
	Encoding    string `mapstructure:"encoding" json:"encoding,omitempty"`
	Development bool   `mapstructure:"development" json:"development,omitempty"`
}

type Database struct {
	Type          string `mapstructure:"type" json:"type,omitempty"`
	DSN           string `mapstructure:"dsn" json:"dsn,omitempty"`
	Username      string `mapstructure:"username" json:"username,omitempty"`
	Password      string `mapstructure:"password" json:"password,omitempty"`
	Database      string `mapstructure:"database" json:"database,omitempty"`
	Host          string `mapstructure:"host" json:"host,omitempty"`
	EnableSSL     bool   `mapstructure:"enableSSL" json:"enableSSL,omitempty"`
	SecretRef     string `mapstructure:"secretRef" json:"secretRef,omitempty"`
	MountedSecret string `mapstructure:"mountedSecret" json:"mountedSecret,omitempty"`
}

type CustomID struct {
	Enabled bool     `mapstructure:"enabled" json:"enabled,omitempty"`
	Fields  []string `mapstructure:"fields" json:"fields,omitempty"`
}

type SourceConfig struct {
	CustomID `mapstructure:"customID" json:"customID,omitempty"`
}

// Config of the PolicyReporter
type Config struct {
	Version        string                  `json:"version,omitempty"`
	Namespace      string                  `mapstructure:"namespace" json:"namespace,omitempty"`
	Loki           *Loki                   `mapstructure:"loki" json:"loki,omitempty"`
	Elasticsearch  *Elasticsearch          `mapstructure:"elasticsearch" json:"elasticsearch,omitempty"`
	Slack          *Slack                  `mapstructure:"slack" json:"slack,omitempty"`
	Discord        *Discord                `mapstructure:"discord" json:"discord,omitempty"`
	Teams          *Teams                  `mapstructure:"teams" json:"teams,omitempty"`
	S3             *S3                     `mapstructure:"s3" json:"s3,omitempty"`
	Kinesis        *Kinesis                `mapstructure:"kinesis" json:"kinesis,omitempty"`
	SecurityHub    *SecurityHub            `mapstructure:"securityHub" json:"securityHub,omitempty"`
	GCS            *GCS                    `mapstructure:"gcs" json:"gcs,omitempty"`
	UI             *UI                     `mapstructure:"ui" json:"ui,omitempty"`
	Webhook        *Webhook                `mapstructure:"webhook" json:"webhook,omitempty"`
	Telegram       *Telegram               `mapstructure:"telegram" json:"telegram,omitempty"`
	GoogleChat     *GoogleChat             `mapstructure:"googleChat" json:"googleChat,omitempty"`
	API            API                     `mapstructure:"api" json:"api,omitempty"`
	WorkerCount    int                     `mapstructure:"worker" json:"worker,omitempty"`
	DBFile         string                  `mapstructure:"dbfile" json:"dbfile,omitempty"`
	Metrics        Metrics                 `mapstructure:"metrics" json:"metrics,omitempty"`
	REST           REST                    `mapstructure:"rest" json:"rest,omitempty"`
	ReportFilter   ReportFilter            `mapstructure:"reportFilter" json:"reportFilter,omitempty"`
	Redis          Redis                   `mapstructure:"redis" json:"redis,omitempty"`
	Profiling      Profiling               `mapstructure:"profiling" json:"profiling,omitempty"`
	EmailReports   EmailReports            `mapstructure:"emailReports" json:"emailReports,omitempty"`
	LeaderElection LeaderElection          `mapstructure:"leaderElection" json:"leaderElection,omitempty"`
	K8sClient      K8sClient               `mapstructure:"k8sClient" json:"k8sClient,omitempty"`
	Logging        Logging                 `mapstructure:"logging" json:"logging,omitempty"`
	Database       Database                `mapstructure:"database" json:"database,omitempty"`
	SourceConfig   map[string]SourceConfig `mapstructure:"sourceConfig" json:"sourceConfig,omitempty"`
	Templates      Templates               `mapstructure:"templates" json:"templates,omitempty"`
}
