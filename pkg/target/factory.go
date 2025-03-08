package target

type ValueFilter struct {
	Include  []string       `mapstructure:"include"`
	Exclude  []string       `mapstructure:"exclude"`
	Selector map[string]any `mapstructure:"selector"`
}

type Filter struct {
	Namespaces   ValueFilter `mapstructure:"namespaces"`
	Status       ValueFilter `mapstructure:"status"`
	Severities   ValueFilter `mapstructure:"severities"`
	Policies     ValueFilter `mapstructure:"policies"`
	Sources      ValueFilter `mapstructure:"sources"`
	ReportLabels ValueFilter `mapstructure:"reportLabels"`
}

type Config[T any] struct {
	Config          *T                `mapstructure:"config"`
	Name            string            `mapstructure:"name"`
	MinimumSeverity string            `mapstructure:"minimumSeverity"`
	Filter          Filter            `mapstructure:"filter"`
	SecretRef       string            `mapstructure:"secretRef"`
	MountedSecret   string            `mapstructure:"mountedSecret"`
	Sources         []string          `mapstructure:"sources"`
	CustomFields    map[string]string `mapstructure:"customFields"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup"`
	Channels        []*Config[T]      `mapstructure:"channels"`
	Valid           bool              `mapstructure:"-"`
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

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"accessKeyId"`
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
	ChatID         string `mapstructure:"chatId"`
}

type SlackOptions struct {
	WebhookOptions `mapstructure:",squash"`
	Channel        string `mapstructure:"channel"`
}

type LokiOptions struct {
	HostOptions `mapstructure:",squash"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Path        string `mapstructure:"path"`
}

type ElasticsearchOptions struct {
	HostOptions `mapstructure:",squash"`
	Index       string `mapstructure:"index"`
	Rotation    string `mapstructure:"rotation"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	APIKey      string `mapstructure:"apiKey"`
	TypelessAPI bool   `mapstructure:"typelessApi"`
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
	AWSConfig      `mapstructure:",squash"`
	AccountID      string `mapstructure:"accountId"`
	ProductName    string `mapstructure:"productName"`
	CompanyName    string `mapstructure:"companyName"`
	DelayInSeconds int    `mapstructure:"delayInSeconds"`
	Synchronize    bool   `mapstructure:"synchronize"`
}

type GCSOptions struct {
	Credentials string `mapstructure:"credentials"`
	Prefix      string `mapstructure:"prefix"`
	Bucket      string `mapstructure:"bucket"`
}

type PagerDutyOptions struct {
	APIToken  string `mapstructure:"apiToken"`
	ServiceID string `mapstructure:"serviceId"`
}

type Targets struct {
	Loki          *Config[LokiOptions]          `mapstructure:"loki"`
	Elasticsearch *Config[ElasticsearchOptions] `mapstructure:"elasticsearch"`
	Slack         *Config[SlackOptions]         `mapstructure:"slack"`
	Discord       *Config[WebhookOptions]       `mapstructure:"discord"`
	Teams         *Config[WebhookOptions]       `mapstructure:"teams"`
	Webhook       *Config[WebhookOptions]       `mapstructure:"webhook"`
	GoogleChat    *Config[WebhookOptions]       `mapstructure:"googleChat"`
	Telegram      *Config[TelegramOptions]      `mapstructure:"telegram"`
	S3            *Config[S3Options]            `mapstructure:"s3"`
	Kinesis       *Config[KinesisOptions]       `mapstructure:"kinesis"`
	SecurityHub   *Config[SecurityHubOptions]   `mapstructure:"securityHub"`
	GCS           *Config[GCSOptions]           `mapstructure:"gcs"`
	PagerDuty     *Config[PagerDutyOptions]    `mapstructure:"pagerduty"`
}

type Factory interface {
	CreateClients(config *Targets) *Collection
	CreateLokiTarget(config, parent *Config[LokiOptions]) *Target
	CreateElasticsearchTarget(config, parent *Config[ElasticsearchOptions]) *Target
	CreateSlackTarget(config, parent *Config[SlackOptions]) *Target
	CreateDiscordTarget(config, parent *Config[WebhookOptions]) *Target
	CreateTeamsTarget(config, parent *Config[WebhookOptions]) *Target
	CreateWebhookTarget(config, parent *Config[WebhookOptions]) *Target
	CreateTelegramTarget(config, parent *Config[TelegramOptions]) *Target
	CreateGoogleChatTarget(config, parent *Config[WebhookOptions]) *Target
	CreateS3Target(config, parent *Config[S3Options]) *Target
	CreateKinesisTarget(config, parent *Config[KinesisOptions]) *Target
	CreateSecurityHubTarget(config, parent *Config[SecurityHubOptions]) *Target
	CreateGCSTarget(config, parent *Config[GCSOptions]) *Target
}
