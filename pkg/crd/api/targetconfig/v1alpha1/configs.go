package v1alpha1

import "github.com/kyverno/policy-reporter/pkg/filters"

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"accessKeyId" json:"accessKeyId"`
	SecretAccessKey string `mapstructure:"secretAccessKey" json:"secretAccessKey"`
	// +optional
	Region string `mapstructure:"region" json:"region"`
	// +optional
	Endpoint string `mapstructure:"endpoint" json:"endpoint"`
}

type KeepaliveConfig struct {
	// +optional
	Interval string `mapstructure:"interval" json:"interval"` // Duration string like "5m"
	// +optional
	Params map[string]string `mapstructure:"params" json:"params"`
}

type WebhookOptions struct {
	Webhook string `mapstructure:"webhook" json:"webhook"`
	// +optional
	SkipTLS bool `mapstructure:"skipTLS" json:"skipTLS"`
	// +optional
	Certificate string `mapstructure:"certificate" json:"certificate"`
	// +optional
	Headers map[string]string `mapstructure:"headers" json:"headers"`
	// +optional
	Keepalive *KeepaliveConfig `mapstructure:"keepalive" json:"keepalive"`
}

type JiraOptions struct {
	ProjectKey string `mapstructure:"projectKey" json:"projectKey"`
	// +optional
	Host string `mapstructure:"host" json:"host"`
	// +optional
	APIToken string `mapstructure:"apiToken" json:"apiToken"`
	// +optional
	Username string `mapstructure:"username" json:"username"`
	// +optional
	Password string `mapstructure:"password" json:"password"`
	// +optional
	IssueType string `mapstructure:"issueType" json:"issueType"`
	// +optional
	SummaryTemplate string `mapstructure:"summaryTemplate" json:"summaryTemplate"`
	// +optional
	APIVersion string `mapstructure:"apiVersion" json:"apiVersion"`
	// +optional
	Labels []string `mapstructure:"labels" json:"labels"`
	// +optional
	Components []string `mapstructure:"components" json:"components"`
	// +optional
	SkipTLS bool `mapstructure:"skipTLS" json:"skipTLS"`
	// +optional
	Certificate string `mapstructure:"certificate" json:"certificate"`
}

type HostOptions struct {
	Host string `mapstructure:"host" json:"host"`
	// +optional
	SkipTLS bool `mapstructure:"skipTLS" json:"skipTLS"`
	// +optional
	Certificate string `mapstructure:"certificate" json:"certificate"`
	// +optional
	Headers map[string]string `mapstructure:"headers" json:"headers"`
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

type SplunkOptions struct {
	HostOptions `mapstructure:",squash" json:",inline"`

	Token string `mapstructure:"token" json:"token"`
}

type LokiOptions struct {
	HostOptions `mapstructure:",squash" json:",inline"`
	// +optional
	Username string `mapstructure:"username" json:"username"`
	// +optional
	Password string `mapstructure:"password" json:"password"`
	// +optional
	Path string `mapstructure:"path" json:"path"`
}

type ElasticsearchOptions struct {
	HostOptions `mapstructure:",squash" json:",inline"`
	Index       string `mapstructure:"index" json:"index"`
	// +optional
	Rotation string `mapstructure:"rotation" json:"rotation"`
	// +optional
	Username string `mapstructure:"username" json:"username"`
	// +optional
	Password string `mapstructure:"password" json:"password"`
	// +optional
	APIKey string `mapstructure:"apiKey" json:"apiKey"`
	// +optional
	TypelessAPI bool `mapstructure:"typelessApi" json:"typelessApi"`
}

type S3Options struct {
	AWSConfig `mapstructure:",squash" json:",inline"`
	// +optional
	Prefix string `mapstructure:"prefix" json:"prefix"`
	Bucket string `mapstructure:"bucket" json:"bucket"`
	// +optional
	BucketKeyEnabled bool `mapstructure:"bucketKeyEnabled" json:"bucketKeyEnabled"`
	// +optional
	KmsKeyID string `mapstructure:"kmsKeyId" json:"kmsKeyId"`
	// +optional
	ServerSideEncryption string `mapstructure:"serverSideEncryption" json:"serverSideEncryption"`
	// +optional
	PathStyle bool `mapstructure:"pathStyle" json:"pathStyle"`
}

type KinesisOptions struct {
	AWSConfig  `mapstructure:",squash" json:",inline"`
	StreamName string `mapstructure:"streamName" json:"streamName"`
}

type SecurityHubOptions struct {
	AWSConfig   `mapstructure:",squash" json:",inline"`
	AccountID   string `mapstructure:"accountId" json:"accountId"`
	ProductName string `mapstructure:"productName" json:"productName"`
	// +optional
	CompanyName string `mapstructure:"companyName" json:"companyName"`
	// +optional
	DelayInSeconds int `mapstructure:"delayInSeconds" json:"delayInSeconds"`
	// +optional
	Synchronize bool `mapstructure:"synchronize" json:"synchronize"`
}

type GCSOptions struct {
	Credentials string `mapstructure:"credentials" json:"credentials"`
	Prefix      string `mapstructure:"prefix" json:"prefix"`
	Bucket      string `mapstructure:"bucket" json:"bucket"`
}

type Config struct {
	// +optional
	Name string `mapstructure:"name" json:"name"`
	// +optional
	MinimumSeverity string `mapstructure:"minimumSeverity" json:"minimumSeverity"`
	// +optional
	Filter filters.Filter `mapstructure:"filter" json:"filter"`
	// +optional
	SecretRef string `mapstructure:"secretRef" json:"secretRef"`
	// +optional
	MountedSecret string `mapstructure:"mountedSecret" json:"mountedSecret"`
	// +optional
	Sources []string `mapstructure:"sources" json:"sources"`
	// +optional
	CustomFields map[string]string `mapstructure:"customFields" json:"customFields"`
	// +optional
	// SkipExisting bool `mapstructure:"skipExistingOnStartup" json:"skipExistingOnStartup"`
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
