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
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

// TargetFactory manages target creation
type TargetFactory struct {
	secretClient secrets.Client
}

// LokiClients resolver method
func createClients[T any](name string, config *Target[T], mapper func(*Target[T], *Target[T]) target.Client) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	if config.Config == nil {
		config.Config = new(T)
	}

	setFallback(&config.Name, name)

	if client := mapper(config, &Target[T]{Config: new(T)}); client != nil {
		clients = append(clients, client)
		config.Valid = true
	}

	for i, channel := range config.Channels {
		setFallback(&config.Name, fmt.Sprintf("%s Channel %d", name, i+1))

		if channel.Config == nil {
			channel.Config = new(T)
		}

		if client := mapper(channel, config); client != nil {
			clients = append(clients, client)
			channel.Valid = true
		}
	}

	return clients
}

// LokiClients resolver method
func (f *TargetFactory) CreateClients(config *Targets) []target.Client {
	clients := make([]target.Client, 0)
	if config == nil {
		return clients
	}

	clients = append(clients, createClients("Loki", config.Loki, f.createLokiClient)...)
	clients = append(clients, createClients("Elasticsearch", config.Elasticsearch, f.createElasticsearchClient)...)
	clients = append(clients, createClients("Slack", config.Slack, f.createSlackClient)...)
	clients = append(clients, createClients("Discord", config.Discord, f.createDiscordClient)...)
	clients = append(clients, createClients("Teams", config.Teams, f.createTeamsClient)...)
	clients = append(clients, createClients("GoogleChat", config.GoogleChat, f.createGoogleChatClient)...)
	clients = append(clients, createClients("Telegram", config.Telegram, f.createTelegramClient)...)
	clients = append(clients, createClients("Webhook", config.Webhook, f.createWebhookClient)...)
	clients = append(clients, createClients("S3", config.S3, f.createS3Client)...)
	clients = append(clients, createClients("Kinesis", config.Kinesis, f.createKinesisClient)...)
	clients = append(clients, createClients("SecurityHub", config.SecurityHub, f.createSecurityHub)...)
	clients = append(clients, createClients("GoogleCloudStorage", config.GCS, f.createGCSClient)...)

	return clients
}

func (f *TargetFactory) createSlackClient(config, parent *Target[SlackOptions]) target.Client {
	if config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Config.Webhook == "" && config.Config.Channel == "" {
		return nil
	}

	setFallback(&config.Config.Webhook, parent.Config.Webhook)

	if config.Config.Webhook == "" {
		return nil
	}

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return slack.NewClient(slack.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Config.Webhook,
		Channel:       config.Config.Channel,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient("", false),
	})
}

func (f *TargetFactory) createLokiClient(config, parent *Target[LokiOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Config.Host == "" && parent.Config.Host == "" {
		return nil
	}

	setFallback(&config.Config.Path, "/api/prom/push")
	setFallback(&config.Config.Host, parent.Config.Host)
	setFallback(&config.Config.Certificate, parent.Config.Certificate)
	setFallback(&config.Config.Path, parent.Config.Path)
	setBool(&config.Config.SkipTLS, parent.Config.SkipTLS)

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return loki.NewClient(loki.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Config.Host + config.Config.Path,
		CustomLabels:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createElasticsearchClient(config, parent *Target[ElasticsearchOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Config.Host == "" && parent.Config.Host == "" {
		return nil
	}

	setFallback(&config.Config.Host, parent.Config.Host)
	setFallback(&config.Config.Certificate, parent.Config.Certificate)
	setBool(&config.Config.SkipTLS, parent.Config.SkipTLS)
	setFallback(&config.Config.Username, parent.Config.Username)
	setFallback(&config.Config.Password, parent.Config.Password)
	setFallback(&config.Config.APIKey, parent.Config.APIKey)
	setFallback(&config.Config.Index, parent.Config.Index, "policy-reporter")
	setFallback(&config.Config.Rotation, parent.Config.Rotation, elasticsearch.Daily)

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return elasticsearch.NewClient(elasticsearch.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Config.Host,
		Username:      config.Config.Username,
		Password:      config.Config.Password,
		ApiKey:        config.Config.APIKey,
		Rotation:      config.Config.Rotation,
		Index:         config.Config.Index,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createDiscordClient(config, parent *Target[WebhookOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return discord.NewClient(discord.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Config.Webhook,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createTeamsClient(config, parent *Target[WebhookOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return teams.NewClient(teams.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Config.Webhook,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createWebhookClient(config, parent *Target[WebhookOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return webhook.NewClient(webhook.Options{
		ClientOptions: config.ClientOptions(),
		Host:          config.Config.Webhook,
		Headers:       config.Config.Headers,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createTelegramClient(config, parent *Target[TelegramOptions]) target.Client {
	if config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Config.Token, parent.Config.Token)

	if config.Config.ChatID == "" || config.Config.Token == "" {
		return nil
	}

	setFallback(&config.Config.Webhook, parent.Config.Webhook)
	setFallback(&config.Config.Certificate, parent.Config.Certificate)
	setBool(&config.Config.SkipTLS, parent.Config.SkipTLS)

	config.MapBaseParent(parent)

	if len(parent.Config.Headers) > 0 {
		headers := map[string]string{}
		for header, value := range parent.Config.Headers {
			headers[header] = value
		}
		for header, value := range config.Config.Headers {
			headers[header] = value
		}

		config.Config.Headers = headers
	}

	host := "https://api.telegram.org"
	if config.Config.Webhook != "" {
		host = strings.TrimSuffix(config.Config.Webhook, "/")
	}

	zap.S().Infof("%s configured", config.Name)

	return telegram.NewClient(telegram.Options{
		ClientOptions: config.ClientOptions(),
		Host:          fmt.Sprintf("%s/bot%s/sendMessage", host, config.Config.Token),
		ChatID:        config.Config.ChatID,
		Headers:       config.Config.Headers,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createGoogleChatClient(config, parent *Target[WebhookOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return googlechat.NewClient(googlechat.Options{
		ClientOptions: config.ClientOptions(),
		Webhook:       config.Config.Webhook,
		Headers:       config.Config.Headers,
		CustomFields:  config.CustomFields,
		HTTPClient:    http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
	})
}

func (f *TargetFactory) createS3Client(config, parent *Target[S3Options]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	config.Config.MapAWSParent(parent.Config.AWSConfig)
	if config.Config.Endpoint == "" && !hasAWSIdentity() {
		return nil
	}

	sugar := zap.S()

	if err := checkAWSConfig(config.Name, config.Config.AWSConfig, parent.Config.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	setFallback(&config.Config.Bucket, parent.Config.Bucket)
	if config.Config.Bucket == "" {
		sugar.Errorf("%s.Bucket has not been declared", config.Name)
		return nil
	}

	setFallback(&config.Config.Region, os.Getenv("AWS_REGION"))
	setFallback(&config.Config.Prefix, parent.Config.Prefix, "policy-reporter")
	setFallback(&config.Config.KmsKeyID, parent.Config.KmsKeyID)
	setFallback(&config.Config.ServerSideEncryption, parent.Config.ServerSideEncryption)
	setBool(&config.Config.BucketKeyEnabled, parent.Config.BucketKeyEnabled)

	config.MapBaseParent(parent)

	s3Client := helper.NewS3Client(
		config.Config.AccessKeyID,
		config.Config.SecretAccessKey,
		config.Config.Region,
		config.Config.Endpoint,
		config.Config.Bucket,
		config.Config.PathStyle,
		helper.WithKMS(config.Config.BucketKeyEnabled, &config.Config.KmsKeyID, &config.Config.ServerSideEncryption),
	)

	sugar.Infof("%s configured", config.Name)

	return s3.NewClient(s3.Options{
		ClientOptions: config.ClientOptions(),
		S3:            s3Client,
		CustomFields:  config.CustomFields,
		Prefix:        config.Config.Prefix,
	})
}

func (f *TargetFactory) createKinesisClient(config, parent *Target[KinesisOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	config.Config.MapAWSParent(parent.Config.AWSConfig)
	if config.Config.Endpoint == "" {
		return nil
	}

	sugar := zap.S()
	if err := checkAWSConfig(config.Name, config.Config.AWSConfig, parent.Config.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	setFallback(&config.Config.StreamName, parent.Config.StreamName)
	if config.Config.StreamName == "" {
		sugar.Errorf("%s.StreamName has not been declared", config.Name)
		return nil
	}

	config.MapBaseParent(parent)

	kinesisClient := helper.NewKinesisClient(
		config.Config.AccessKeyID,
		config.Config.SecretAccessKey,
		config.Config.Region,
		config.Config.Endpoint,
		config.Config.StreamName,
	)

	sugar.Infof("%s configured", config.Name)

	return kinesis.NewClient(kinesis.Options{
		ClientOptions: config.ClientOptions(),
		CustomFields:  config.CustomFields,
		Kinesis:       kinesisClient,
	})
}

func (f *TargetFactory) createSecurityHub(config, parent *Target[SecurityHubOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Config.AccountID, parent.Config.AccountID)
	if config.Config.AccountID == "" {
		return nil
	}

	sugar := zap.S()
	if err := checkAWSConfig(config.Name, config.Config.AWSConfig, parent.Config.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	config.Config.MapAWSParent(parent.Config.AWSConfig)
	config.MapBaseParent(parent)

	client := helper.NewHubClient(
		config.Config.AccessKeyID,
		config.Config.SecretAccessKey,
		config.Config.Region,
		config.Config.Endpoint,
	)

	sugar.Infof("%s configured", config.Name)

	return securityhub.NewClient(securityhub.Options{
		ClientOptions: config.ClientOptions(),
		CustomFields:  config.CustomFields,
		Client:        client,
		AccountID:     config.Config.AccountID,
		Region:        config.Config.Region,
	})
}

func (f *TargetFactory) createGCSClient(config, parent *Target[GCSOptions]) target.Client {
	if config == nil || config.Config == nil {
		return nil
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Config.Bucket, parent.Config.Bucket)
	if config.Config.Bucket == "" {
		return nil
	}

	sugar := zap.S()

	setFallback(&config.Config.Credentials, parent.Config.Credentials)
	if config.Config.Credentials == "" {
		sugar.Errorf("%s.Credentials has not been declared", config.Name)
		return nil
	}

	setFallback(&config.Config.Prefix, parent.Config.Prefix, "policy-reporter")

	config.MapBaseParent(parent)

	gcsClient := helper.NewGCSClient(
		context.Background(),
		config.Config.Credentials,
		config.Config.Bucket,
	)
	if gcsClient == nil {
		return nil
	}

	sugar.Infof("%s configured", config.Name)

	return gcs.NewClient(gcs.Options{
		ClientOptions: config.ClientOptions(),
		Client:        gcsClient,
		CustomFields:  config.CustomFields,
		Prefix:        config.Config.Prefix,
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
	case *Target[LokiOptions]:
		if values.Host != "" {
			c.Config.Host = values.Host
		}

	case *Target[SlackOptions]:
		if values.Webhook != "" {
			c.Config.Webhook = values.Webhook
			c.Config.Channel = values.Channel
		}

	case *Target[WebhookOptions]:
		if values.Webhook != "" {
			c.Config.Webhook = values.Webhook
		}
		if values.Token != "" {
			if c.Config.Headers == nil {
				c.Config.Headers = make(map[string]string)
			}

			c.Config.Headers["Authorization"] = values.Token
		}

	case *Target[ElasticsearchOptions]:
		if values.Host != "" {
			c.Config.Host = values.Host
		}
		if values.Username != "" {
			c.Config.Username = values.Username
		}
		if values.Password != "" {
			c.Config.Password = values.Password
		}
		if values.APIKey != "" {
			c.Config.APIKey = values.APIKey
		}

	case *Target[S3Options]:
		if values.AccessKeyID != "" {
			c.Config.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.Config.SecretAccessKey = values.SecretAccessKey
		}
		if values.KmsKeyID != "" {
			c.Config.KmsKeyID = values.KmsKeyID
		}

	case *Target[KinesisOptions]:
		if values.AccessKeyID != "" {
			c.Config.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.Config.SecretAccessKey = values.SecretAccessKey
		}

	case *Target[SecurityHubOptions]:
		if values.AccessKeyID != "" {
			c.Config.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.Config.SecretAccessKey = values.SecretAccessKey
		}
		if values.AccountID != "" {
			c.Config.AccountID = values.AccessKeyID
		}

	case *Target[GCSOptions]:
		if values.Credentials != "" {
			c.Config.Credentials = values.Credentials
		}

	case *Target[TelegramOptions]:
		if values.Token != "" {
			c.Config.Token = values.Token
		}
		if values.Host != "" {
			c.Config.Webhook = values.Host
		}
	}
}

func NewTargetFactory(secretClient secrets.Client) *TargetFactory {
	return &TargetFactory{secretClient: secretClient}
}

func mapWebhookTarget(config, parent *Target[WebhookOptions]) {
	setFallback(&config.Config.Webhook, parent.Config.Webhook)
	setFallback(&config.Config.Certificate, parent.Config.Certificate)
	setBool(&config.Config.SkipTLS, parent.Config.SkipTLS)

	config.MapBaseParent(parent)

	if len(parent.Config.Headers) > 0 {
		headers := map[string]string{}
		for header, value := range parent.Config.Headers {
			headers[header] = value
		}
		for header, value := range config.Config.Headers {
			headers[header] = value
		}

		config.Config.Headers = headers
	}
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
