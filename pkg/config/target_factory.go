package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
	"github.com/kyverno/policy-reporter/pkg/target/gcs"
	"github.com/kyverno/policy-reporter/pkg/target/googlechat"
	"github.com/kyverno/policy-reporter/pkg/target/http"
	"github.com/kyverno/policy-reporter/pkg/target/kinesis"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
	"github.com/kyverno/policy-reporter/pkg/target/securityhub"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
	"github.com/kyverno/policy-reporter/pkg/target/telegram"
	"github.com/kyverno/policy-reporter/pkg/target/ui"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

// TargetFactory manages target creation
type TargetFactory struct {
	secretClient secrets.Client
}

// LokiClients resolver method
func (f *TargetFactory) LokiClients(config *Loki) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Loki")
	setFallback(&config.Path, "/api/prom/push")

	if loki := f.createLokiClient(config, &Loki{}); loki != nil {
		clients = append(clients, loki)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Loki Channel %d", i+1))

		if loki := f.createLokiClient(channel, config); loki != nil {
			clients = append(clients, loki)
		}
	}

	return clients
}

// ElasticsearchClients resolver method
func (f *TargetFactory) ElasticsearchClients(config *Elasticsearch) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Elasticsearch")

	if es := f.createElasticsearchClient(config, &Elasticsearch{}); es != nil {
		clients = append(clients, es)
	}

	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Elasticsearch Channel %d", i+1))

		if es := f.createElasticsearchClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// SlackClients resolver method
func (f *TargetFactory) SlackClients(config *Slack) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Slack")

	if es := f.createSlackClient(config, &Slack{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Slack Channel %d", i+1))

		if es := f.createSlackClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// DiscordClients resolver method
func (f *TargetFactory) DiscordClients(config *Discord) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Discord")

	if es := f.createDiscordClient(config, &Discord{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Discord Channel %d", i+1))

		if es := f.createDiscordClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// TeamsClients resolver method
func (f *TargetFactory) TeamsClients(config *Teams) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Teams")

	if es := f.createTeamsClient(config, &Teams{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Teams Channel %d", i+1))

		if es := f.createTeamsClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// WebhookClients resolver method
func (f *TargetFactory) WebhookClients(config *Webhook) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Webhook")

	if es := f.createWebhookClient(config, &Webhook{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Webhook Channel %d", i+1))

		if es := f.createWebhookClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// UIClient resolver method
func (f *TargetFactory) UIClient(config *UI) target.Client {
	if config == nil || config.Host == "" {
		return nil
	}

	setFallback(&config.Name, "UI")

	zap.L().Info("UI configured")

	return ui.NewClient(ui.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Host,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

// S3Clients resolver method
func (f *TargetFactory) S3Clients(config *S3) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "S3")

	if es := f.createS3Client(config, &S3{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("S3 Channel %d", i+1))

		if es := f.createS3Client(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// KinesisClients resolver method
func (f *TargetFactory) KinesisClients(config *Kinesis) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Kinesis")

	if es := f.createKinesisClient(config, &Kinesis{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Kinesis Channel %d", i+1))

		if es := f.createKinesisClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// SecurityHub resolver method
func (f *TargetFactory) SecurityHubs(config *SecurityHub) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "SecurityHub")

	if es := f.createSecurityHub(config, &SecurityHub{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("SecurityHub Channel %d", i+1))

		if es := f.createSecurityHub(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// GCSClients resolver method
func (f *TargetFactory) GCSClients(config *GCS) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "GoogleCloudStorage")

	if es := f.createGCSClient(config, &GCS{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("GCS Channel %d", i+1))

		if es := f.createGCSClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// TelegramClients resolver method
func (f *TargetFactory) TelegramClients(config *Telegram) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "Telegram")

	if es := f.createTelegramClient(config, &Telegram{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("Telegram Channel %d", i+1))

		if es := f.createTelegramClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// GoogleChatClients resolver method
func (f *TargetFactory) GoogleChatClients(config *GoogleChat) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	setFallback(&config.Name, "GoogleChat")

	if es := f.createGoogleChatClient(config, &GoogleChat{}); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("GoogleChat Channel %d", i+1))

		if es := f.createGoogleChatClient(channel, config); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

func (f *TargetFactory) createSlackClient(config, parent *Slack) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Webhook == "" && config.Channel == "" {
		return nil
	}

	setFallback(&config.Webhook, parent.Webhook)

	if config.Webhook == "" {
		return nil
	}

	config.MapBaseParent(parent.TargetBaseOptions)

	zap.S().Infof("%s configured", config.Name)

	return slack.NewClient(slack.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Webhook,
		Channel:       config.Channel,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient("", false),
	})
}

func (f *TargetFactory) createLokiClient(config, parent *Loki) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Host == "" && parent.Host == "" {
		return nil
	}

	setFallback(&config.Host, parent.Host)
	setFallback(&config.Certificate, parent.Certificate)
	setFallback(&config.Path, parent.Path)
	setBool(&config.SkipTLS, parent.SkipTLS)

	config.MapBaseParent(parent.TargetBaseOptions)

	zap.S().Infof("%s configured", config.Name)

	if config.CustomFields == nil {
		config.CustomFields = make(map[string]string)
	}

	if config.CustomLabels != nil {
		for k, v := range config.CustomLabels {
			config.CustomFields[k] = v
		}
	}

	return loki.NewClient(loki.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Host + config.Path,
		CustomLabels:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createElasticsearchClient(config, parent *Elasticsearch) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Host == "" && parent.Host == "" {
		return nil
	}

	setFallback(&config.Host, parent.Host)
	setFallback(&config.Certificate, parent.Certificate)
	setBool(&config.SkipTLS, parent.SkipTLS)
	setFallback(&config.Username, parent.Username)
	setFallback(&config.Password, parent.Password)
	setFallback(&config.Index, parent.Index, "policy-reporter")
	setFallback(&config.Rotation, parent.Rotation, elasticsearch.Daily)

	config.MapBaseParent(parent.TargetBaseOptions)

	zap.S().Infof("%s configured", config.Name)

	return elasticsearch.NewClient(elasticsearch.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Host,
		Username:      config.Username,
		Password:      config.Password,
		Rotation:      config.Rotation,
		Index:         config.Index,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createDiscordClient(config, parent *Discord) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Webhook == "" {
		return nil
	}

	config.MapBaseParent(parent.TargetBaseOptions)

	zap.S().Infof("%s configured", config.Name)

	return discord.NewClient(discord.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Webhook,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient("", false),
	})
}

func (f *TargetFactory) createTeamsClient(config, parent *Teams) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Webhook == "" {
		return nil
	}

	setFallback(&config.Certificate, parent.Certificate)
	setBool(&config.SkipTLS, parent.SkipTLS)

	config.MapBaseParent(parent.TargetBaseOptions)

	zap.S().Infof("%s configured", config.Name)

	return teams.NewClient(teams.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Webhook,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createWebhookClient(config, parent *Webhook) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Host == "" {
		return nil
	}

	setFallback(&config.Certificate, parent.Certificate)
	setBool(&config.SkipTLS, parent.SkipTLS)
	config.MapBaseParent(parent.TargetBaseOptions)

	if len(parent.Headers) > 0 {
		headers := map[string]string{}
		for header, value := range parent.Headers {
			headers[header] = value
		}
		for header, value := range config.Headers {
			headers[header] = value
		}

		config.Headers = headers
	}

	zap.S().Infof("%s configured", config.Name)

	return webhook.NewClient(webhook.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Host,
		Headers:       config.Headers,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createTelegramClient(config, parent *Telegram) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Token, parent.Token)

	if config.ChatID == "" || config.Token == "" {
		return nil
	}

	setFallback(&config.Host, parent.Host)
	setFallback(&config.Certificate, parent.Certificate)
	setBool(&config.SkipTLS, parent.SkipTLS)

	config.MapBaseParent(parent.TargetBaseOptions)

	if len(parent.Headers) > 0 {
		headers := map[string]string{}
		for header, value := range parent.Headers {
			headers[header] = value
		}
		for header, value := range config.Headers {
			headers[header] = value
		}

		config.Headers = headers
	}

	host := "https://api.telegram.org"
	if config.Host != "" {
		host = strings.TrimSuffix(config.Host, "/")
	}

	zap.S().Infof("%s configured", config.Name)

	return telegram.NewClient(telegram.Options{
		ClientOptions: config.ClientOptions(),
		Host:          fmt.Sprintf("%s/bot%s/sendMessage", host, config.Token),
		ChatID:        config.ChatID,
		Headers:       config.Headers,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createGoogleChatClient(config, parent *GoogleChat) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Webhook, parent.Webhook)

	if config.Webhook == "" {
		return nil
	}

	setFallback(&config.Certificate, parent.Certificate)
	setBool(&config.SkipTLS, parent.SkipTLS)
	config.MapBaseParent(parent.TargetBaseOptions)

	if len(parent.Headers) > 0 {
		headers := map[string]string{}
		for header, value := range parent.Headers {
			headers[header] = value
		}
		for header, value := range config.Headers {
			headers[header] = value
		}

		config.Headers = headers
	}

	zap.S().Infof("%s configured", config.Name)

	return googlechat.NewClient(googlechat.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Webhook,
		Headers:       config.Headers,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createS3Client(config, parent *S3) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	config.MapAWSParent(parent.AWSConfig)
	if config.Endpoint == "" && !hasAWSIdentity() {
		return nil
	}

	sugar := zap.S()

	if err := checkAWSConfig(config.Name, config.AWSConfig, parent.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	setFallback(&config.Bucket, parent.Bucket)
	if config.Bucket == "" {
		sugar.Errorf("%s.Bucket has not been declared", config.Name)
		return nil
	}

	setFallback(&config.Prefix, parent.Prefix, "policy-reporter")
	setFallback(&config.KmsKeyID, parent.KmsKeyID)
	setFallback(&config.ServerSideEncryption, parent.ServerSideEncryption)
	setBool(&config.BucketKeyEnabled, parent.BucketKeyEnabled)

	config.MapBaseParent(parent.TargetBaseOptions)

	s3Client := helper.NewS3Client(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.Bucket,
		config.PathStyle,
		helper.WithKMS(&config.BucketKeyEnabled, &config.KmsKeyID, &config.ServerSideEncryption),
	)

	sugar.Infof("%s configured", config.Name)

	return s3.NewClient(s3.Options{
		ClientOptions: config.ClientOptions(),
		S3:            s3Client,
		CustomFields:  config.CustomFields,
		Prefix:        config.Prefix,
	})
}

func (f *TargetFactory) createKinesisClient(config, parent *Kinesis) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	config.MapAWSParent(parent.AWSConfig)
	if config.Endpoint == "" {
		return nil
	}

	sugar := zap.S()
	if err := checkAWSConfig(config.Name, config.AWSConfig, parent.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	setFallback(&config.StreamName, parent.StreamName)
	if config.StreamName == "" {
		sugar.Errorf("%s.StreamName has not been declared", config.Name)
		return nil
	}

	config.MapBaseParent(parent.TargetBaseOptions)

	kinesisClient := helper.NewKinesisClient(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.StreamName,
	)

	sugar.Infof("%s configured", config.Name)

	return kinesis.NewClient(kinesis.Options{
		ClientOptions: config.ClientOptions(),
		CustomFields:  config.CustomFields,
		Kinesis:       kinesisClient,
	})
}

func (f *TargetFactory) createSecurityHub(config, parent *SecurityHub) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.AccountID, parent.AccountID)
	if config.AccountID == "" {
		return nil
	}

	sugar := zap.S()
	if err := checkAWSConfig(config.Name, config.AWSConfig, parent.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	config.MapAWSParent(parent.AWSConfig)
	config.MapBaseParent(parent.TargetBaseOptions)

	client := helper.NewHubClient(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
	)

	sugar.Infof("%s configured", config.Name)

	return securityhub.NewClient(securityhub.Options{
		ClientOptions: config.ClientOptions(),
		CustomFields:  config.CustomFields,
		Client:        client,
		AccountID:     config.AccountID,
		Region:        config.Region,
	})
}

func (f *TargetFactory) createGCSClient(config, parent *GCS) target.Client {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Bucket, parent.Bucket)
	if config.Bucket == "" {
		return nil
	}

	sugar := zap.S()

	setFallback(&config.Credentials, parent.Credentials)
	if config.Credentials == "" {
		sugar.Errorf("%s.Credentials has not been declared", config.Name)
		return nil
	}

	setFallback(&config.Prefix, parent.Prefix, "policy-reporter")

	config.MapBaseParent(parent.TargetBaseOptions)

	gcsClient := helper.NewGCSClient(
		context.Background(),
		config.Credentials,
		config.Bucket,
	)
	if gcsClient == nil {
		return nil
	}

	sugar.Infof("%s configured", config.Name)

	return gcs.NewClient(gcs.Options{
		ClientOptions: config.ClientOptions(),
		Client:        gcsClient,
		CustomFields:  config.CustomFields,
		Prefix:        config.Prefix,
	})
}

func (f *TargetFactory) mapSecretValues(config any, ref, mountedSecret string) {
	values := secrets.Values{}

	if ref != "" {
		secretValues, err := f.secretClient.Get(context.Background(), ref)
		values = secretValues
		if err != nil {
			zap.L().Warn("failed to get secret reference", zap.Error(err))
			return
		}
	}

	if mountedSecret != "" {
		file, err := os.ReadFile(mountedSecret)
		if err != nil {
			zap.L().Warn("failed to get mounted secret", zap.Error(err))
			return
		}
		err = json.Unmarshal(file, &values)
		if err != nil {
			zap.L().Warn("failed to unmarshal mounted secret", zap.Error(err))
			return
		}
	}

	switch c := config.(type) {
	case *Loki:
		if values.Host != "" {
			c.Host = values.Host
		}

	case *Slack:
		if values.Webhook != "" {
			c.Webhook = values.Webhook
			c.Channel = values.Channel
		}

	case *Discord:
		if values.Webhook != "" {
			c.Webhook = values.Webhook
		}

	case *Teams:
		if values.Webhook != "" {
			c.Webhook = values.Webhook
		}

	case *Elasticsearch:
		if values.Host != "" {
			c.Host = values.Host
		}
		if values.Username != "" {
			c.Username = values.Username
		}
		if values.Password != "" {
			c.Password = values.Password
		}

	case *S3:
		if values.AccessKeyID != "" {
			c.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.SecretAccessKey = values.SecretAccessKey
		}
		if values.KmsKeyID != "" {
			c.KmsKeyID = values.KmsKeyID
		}

	case *Kinesis:
		if values.AccessKeyID != "" {
			c.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.SecretAccessKey = values.SecretAccessKey
		}

	case *SecurityHub:
		if values.AccessKeyID != "" {
			c.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.SecretAccessKey = values.SecretAccessKey
		}
		if values.AccountID != "" {
			c.AccountID = values.AccessKeyID
		}

	case *GCS:
		if values.Credentials != "" {
			c.Credentials = values.Credentials
		}

	case *Webhook:
		if values.Host != "" {
			c.Host = values.Host
		}
		if values.Token != "" {
			if c.Headers == nil {
				c.Headers = make(map[string]string)
			}

			c.Headers["Authorization"] = values.Token
		}
	case *Telegram:
		if values.Token != "" {
			c.Token = values.Token
		}
		if values.Host != "" {
			c.Host = values.Host
		}
	case *GoogleChat:
		if values.Webhook != "" {
			c.Webhook = values.Webhook
		}
	}
}

func createResultFilter(filter TargetFilter, minimumPriority string, sources []string) *report.ResultFilter {
	return target.NewResultFilter(
		ToRuleSet(filter.Namespaces),
		ToRuleSet(filter.Priorities),
		ToRuleSet(filter.Policies),
		minimumPriority,
		sources,
	)
}

func createReportFilter(filter TargetFilter) *report.ReportFilter {
	return target.NewReportFilter(
		ToRuleSet(filter.ReportLabels),
	)
}

func NewTargetFactory(secretClient secrets.Client) *TargetFactory {
	return &TargetFactory{secretClient: secretClient}
}

func hasAWSIdentity() bool {
	arn := os.Getenv("AWS_ROLE_ARN")
	file := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	return arn != "" && file != ""
}

func checkAWSConfig(name string, config AWSConfig, parent AWSConfig) error {
	noEnvConfig := !hasAWSIdentity()

	if noEnvConfig && (config.AccessKeyID == "" && parent.AccessKeyID == "") {
		return fmt.Errorf("%s.AccessKeyID has not been declared", name)
	}

	if noEnvConfig && (config.SecretAccessKey == "" && parent.SecretAccessKey == "") {
		return fmt.Errorf("%s.SecretAccessKey has not been declared", name)
	}

	if config.Region == "" && parent.Region == "" {
		return fmt.Errorf("%s.Region has not been declared", name)
	}

	return nil
}

func setFallback(config *string, parents ...string) {
	if *config == "" {
		for _, p := range parents {
			if p != "" {
				*config = p
				return
			}
		}
	}
}

func setBool(config *bool, parent bool) {
	if *config == false {
		*config = parent
	}
}
