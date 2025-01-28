package v1alpha1

import "github.com/kyverno/policy-reporter/pkg/filters"

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"accessKeyId" json:"accessKeyId"`
	SecretAccessKey string `mapstructure:"secretAccessKey" json:"secretAccessKey"`
	Region          string `mapstructure:"region" json:"region"`
	Endpoint        string `mapstructure:"endpoint" json:"endpoint"`
}

type WebhookOptions struct {
	Webhook     string            `mapstructure:"webhook" json:"webhook"`
	SkipTLS     bool              `mapstructure:"skipTLS" json:"skipTLS"`
	Certificate string            `mapstructure:"certificate" json:"certificate"`
	Headers     map[string]string `mapstructure:"headers" json:"headers"`
}

type HostOptions struct {
	Host        string            `mapstructure:"host" json:"host"`
	SkipTLS     bool              `mapstructure:"skipTLS" json:"skipTLS"`
	Certificate string            `mapstructure:"certificate" json:"certificate"`
	Headers     map[string]string `mapstructure:"headers" json:"headers"`
}

type TelegramOptions struct {
	WebhookOptions `mapstructure:",squash" json:",inline"`
	Token          string `mapstructure:"token" json:"token"`
	ChatID         string `mapstructure:"chatId" json:"chatId"`
}

type SlackOptions struct {
	WebhookOptions `mapstructure:",squash" json:",inline"`
	Channel        string `mapstructure:"channel" json:"channel"`
}

type LokiOptions struct {
	HostOptions `mapstructure:",squash" json:",inline"`
	Username    string `mapstructure:"username" json:"username"`
	Password    string `mapstructure:"password" json:"password"`
	Path        string `mapstructure:"path" json:"path"`
}

type ElasticsearchOptions struct {
	HostOptions `mapstructure:",squash" json:",inline"`
	Index       string `mapstructure:"index" json:"index"`
	Rotation    string `mapstructure:"rotation" json:"rotation"`
	Username    string `mapstructure:"username" json:"username"`
	Password    string `mapstructure:"password" json:"password"`
	APIKey      string `mapstructure:"apiKey" json:"apiKey"`
	TypelessAPI bool   `mapstructure:"typelessApi" json:"typelessApi"`
}

type S3Options struct {
	AWSConfig            `mapstructure:",squash" json:",inline"`
	Prefix               string `mapstructure:"prefix" json:"prefix"`
	Bucket               string `mapstructure:"bucket" json:"bucket"`
	BucketKeyEnabled     bool   `mapstructure:"bucketKeyEnabled" json:"bucketKeyEnabled"`
	KmsKeyID             string `mapstructure:"kmsKeyId" json:"kmsKeyId"`
	ServerSideEncryption string `mapstructure:"serverSideEncryption" json:"serverSideEncryption"`
	PathStyle            bool   `mapstructure:"pathStyle" json:"pathStyle"`
}

type KinesisOptions struct {
	AWSConfig  `mapstructure:",squash" json:",inline"`
	StreamName string `mapstructure:"streamName" json:"streamName"`
}

type SecurityHubOptions struct {
	AWSConfig      `mapstructure:",squash" json:",inline"`
	AccountID      string `mapstructure:"accountId" json:"accountId"`
	ProductName    string `mapstructure:"productName" json:"productName"`
	CompanyName    string `mapstructure:"companyName" json:"companyName"`
	DelayInSeconds int    `mapstructure:"delayInSeconds" json:"delayInSeconds"`
	Synchronize    bool   `mapstructure:"synchronize" json:"synchronize"`
}

type GCSOptions struct {
	Credentials string `mapstructure:"credentials" json:"credentials"`
	Prefix      string `mapstructure:"prefix" json:"prefix"`
	Bucket      string `mapstructure:"bucket" json:"bucket"`
}

type Config[T any] struct {
	Config          *T                `mapstructure:"config" json:"config"`
	Name            string            `mapstructure:"name" json:"name"`
	MinimumSeverity string            `mapstructure:"minimumSeverity" json:"minimumSeverity"`
	Filter          filters.Filter    `mapstructure:"filter" json:"filter"`
	SecretRef       string            `mapstructure:"secretRef" json:"secretRef"`
	MountedSecret   string            `mapstructure:"mountedSecret" json:"mountedSecret"`
	Sources         []string          `mapstructure:"sources" json:"sources"`
	CustomFields    map[string]string `mapstructure:"customFields" json:"customFields"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup" json:"skipExistingOnStartup"`
	Channels        []*Config[T]      `mapstructure:"channels" json:"channels"`
	Valid           bool              `mapstructure:"-" json:"-"`
}

type ConfigStrict struct {
	Name            string            `mapstructure:"name" json:"name"`
	MinimumSeverity string            `mapstructure:"minimumSeverity" json:"minimumSeverity"`
	Filter          filters.Filter    `mapstructure:"filter" json:"filter"`
	SecretRef       string            `mapstructure:"secretRef" json:"secretRef"`
	MountedSecret   string            `mapstructure:"mountedSecret" json:"mountedSecret"`
	Sources         []string          `mapstructure:"sources" json:"sources"`
	CustomFields    map[string]string `mapstructure:"customFields" json:"customFields"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup" json:"skipExistingOnStartup"`
}

func (config *Config[T]) MapBaseParent(parent *Config[T]) {
	if config.MinimumSeverity == "" {
		config.MinimumSeverity = parent.MinimumSeverity
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}
}

func (config *Config[T]) Secret() string {
	return config.SecretRef
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
