package config

import "github.com/kyverno/policy-reporter/pkg/target"

type Target[T any] struct {
	Config          *T                `mapstructure:"config"`
	Name            string            `mapstructure:"name"`
	MinimumPriority string            `mapstructure:"minimumPriority"`
	Filter          TargetFilter      `mapstructure:"filter"`
	SecretRef       string            `mapstructure:"secretRef"`
	MountedSecret   string            `mapstructure:"mountedSecret"`
	Sources         []string          `mapstructure:"sources"`
	CustomFields    map[string]string `mapstructure:"customFields"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
	Channels        []*Target[T]      `mapstructure:"channels"`
	Valid           bool              `mapstructure:"-"`
}

func (config *Target[T]) MapBaseParent(parent *Target[T]) {
	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}
}

func (config *Target[T]) ClientOptions() target.ClientOptions {
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

type WebhookOptions struct {
	Webhook     string            `mapstructure:"webhook"`
	SkipTLS     bool              `mapstructure:"skipTLS"`
	Certificate string            `mapstructure:"certificate"`
	Headers     map[string]string `mapstructure:"headers"`
}

type HostOptions struct {
	Host        string            `mapstructure:"host"`
	SkipTLS     bool              `mapstructure:"skipTLS"`
	Certificate string            `mapstructure:"certificate"`
	Headers     map[string]string `mapstructure:"headers"`
}

type TelegramOptions struct {
	WebhookOptions `mapstructure:",squash"`
	Token          string `mapstructure:"token"`
	ChatID         string `mapstructure:"chatID"`
}

type SlackOptions struct {
	WebhookOptions `mapstructure:",squash"`
	Channel        string `mapstructure:"channel"`
}

type LokiOptions struct {
	HostOptions `mapstructure:",squash"`
	Path        string `mapstructure:"path"`
}

type ElasticsearchOptions struct {
	HostOptions `mapstructure:",squash"`
	Index       string `mapstructure:"index"`
	Rotation    string `mapstructure:"rotation"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	APIKey      string `mapstructure:"apiKey"`
}

type S3Options struct {
	AWSConfig            `mapstructure:",squash"`
	Prefix               string `mapstructure:"prefix"`
	Bucket               string `mapstructure:"bucket"`
	BucketKeyEnabled     bool   `mapstructure:"bucketKeyEnabled"`
	KmsKeyID             string `mapstructure:"kmsKeyId"`
	ServerSideEncryption string `mapstructure:"serverSideEncryption"`
	PathStyle            bool   `mapstructure:"pathStyle"`
}

type KinesisOptions struct {
	AWSConfig  `mapstructure:",squash"`
	StreamName string `mapstructure:"streamName"`
}

type SecurityHubOptions struct {
	AWSConfig `mapstructure:",squash"`
	AccountID string `mapstructure:"accountId"`
}

type GCSOptions struct {
	Credentials string `mapstructure:"credentials"`
	Prefix      string `mapstructure:"prefix"`
	Bucket      string `mapstructure:"bucket"`
}

type Targets struct {
	Loki          *Target[LokiOptions]          `mapstructure:"loki"`
	Elasticsearch *Target[ElasticsearchOptions] `mapstructure:"elasticsearch"`
	Slack         *Target[SlackOptions]         `mapstructure:"slack"`
	Discord       *Target[WebhookOptions]       `mapstructure:"discord"`
	Teams         *Target[WebhookOptions]       `mapstructure:"teams"`
	UI            *Target[WebhookOptions]       `mapstructure:"ui"`
	Webhook       *Target[WebhookOptions]       `mapstructure:"webhook"`
	GoogleChat    *Target[WebhookOptions]       `mapstructure:"googleChat"`
	Telegram      *Target[TelegramOptions]      `mapstructure:"telegram"`
	S3            *Target[S3Options]            `mapstructure:"s3"`
	Kinesis       *Target[KinesisOptions]       `mapstructure:"kinesis"`
	SecurityHub   *Target[SecurityHubOptions]   `mapstructure:"securityHub"`
	GCS           *Target[GCSOptions]           `mapstructure:"gcs"`
}
