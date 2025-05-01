package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/filters"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
	"github.com/kyverno/policy-reporter/pkg/target/gcs"
	"github.com/kyverno/policy-reporter/pkg/target/googlechat"
	"github.com/kyverno/policy-reporter/pkg/target/http"
	"github.com/kyverno/policy-reporter/pkg/target/jira"
	"github.com/kyverno/policy-reporter/pkg/target/kinesis"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
	"github.com/kyverno/policy-reporter/pkg/target/provider/aws"
	gs "github.com/kyverno/policy-reporter/pkg/target/provider/gcs"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
	"github.com/kyverno/policy-reporter/pkg/target/securityhub"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
	"github.com/kyverno/policy-reporter/pkg/target/telegram"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
	"github.com/kyverno/policy-reporter/pkg/target/alertmanager"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

// TargetFactory manages target creation
type TargetFactory struct {
	secretClient  secrets.Client
	filterFactory *target.ResultFilterFactory
}

// LokiClients resolver method
func createClients[T any](name string, config *v1alpha1.Config[T], mapper func(*v1alpha1.Config[T], *v1alpha1.Config[T]) *target.Target) []*target.Target {
	clients := make([]*target.Target, 0)
	if config == nil {
		return clients
	}

	if config.Config == nil {
		config.Config = new(T)
	}

	setFallback(&config.Name, name)

	if client := mapper(config, &v1alpha1.Config[T]{Config: new(T)}); client != nil {
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
func (f *TargetFactory) CreateClients(config *target.Targets) *target.Collection {
	targets := make([]*target.Target, 0)
	if config == nil {
		return target.NewCollection()
	}

	targets = append(targets, createClients("Loki", config.Loki, f.CreateLokiTarget)...)
	targets = append(targets, createClients("Elasticsearch", config.Elasticsearch, f.CreateElasticsearchTarget)...)
	targets = append(targets, createClients("Slack", config.Slack, f.CreateSlackTarget)...)
	targets = append(targets, createClients("Discord", config.Discord, f.CreateDiscordTarget)...)
	targets = append(targets, createClients("Teams", config.Teams, f.CreateTeamsTarget)...)
	targets = append(targets, createClients("GoogleChat", config.GoogleChat, f.CreateGoogleChatTarget)...)
	targets = append(targets, createClients("Jira", config.Jira, f.CreateJiraTarget)...)
	targets = append(targets, createClients("Telegram", config.Telegram, f.CreateTelegramTarget)...)
	targets = append(targets, createClients("Webhook", config.Webhook, f.CreateWebhookTarget)...)
	targets = append(targets, createClients("S3", config.S3, f.CreateS3Target)...)
	targets = append(targets, createClients("Kinesis", config.Kinesis, f.CreateKinesisTarget)...)
	targets = append(targets, createClients("SecurityHub", config.SecurityHub, f.CreateSecurityHubTarget)...)
	targets = append(targets, createClients("GoogleCloudStorage", config.GCS, f.CreateGCSTarget)...)
	targets = append(targets, createClients("AlertManager", config.AlertManager, f.CreateAlertManagerTarget)...)

	return target.NewCollection(targets...)
}

func (f *TargetFactory) CreateSingleClient(tc *v1alpha1.TargetConfig) (*target.Target, error) {
	if tc.Spec.S3 != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.S3), f.CreateS3Target)), nil
	} else if tc.Spec.Webhook != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Webhook), f.CreateWebhookTarget)), nil
	} else if tc.Spec.GCS != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.GCS), f.CreateGCSTarget)), nil
	} else if tc.Spec.ElasticSearch != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.ElasticSearch), f.CreateElasticsearchTarget)), nil
	} else if tc.Spec.Telegram != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Telegram), f.CreateTelegramTarget)), nil
	} else if tc.Spec.Kinesis != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Kinesis), f.CreateKinesisTarget)), nil
	} else if tc.Spec.SecurityHub != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.SecurityHub), f.CreateSecurityHubTarget)), nil
	} else if tc.Spec.Loki != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Loki), f.CreateLokiTarget)), nil
	} else if tc.Spec.Slack != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Slack), f.CreateSlackTarget)), nil
	} else if tc.Spec.Teams != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Teams), f.CreateTeamsTarget)), nil
	} else if tc.Spec.Jira != nil {
		return helper.First(createClients(tc.Name, createConfig(tc, tc.Spec.Jira), f.CreateJiraTarget)), nil
	}
	return nil, fmt.Errorf("invalid target type passed")
}

func (f *TargetFactory) CreateSlackTarget(config, parent *v1alpha1.Config[v1alpha1.SlackOptions]) *target.Target {
	if config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
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

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Slack,
		Config:       config,
		ParentConfig: parent,
		Client: slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Channel:      config.Config.Channel,
			Webhook:      config.Config.Webhook,
			CustomFields: config.CustomFields,
			Headers:      config.Config.Headers,
			HTTPClient:   http.NewClient("", false),
		}),
	}
}

func (f *TargetFactory) CreateLokiTarget(config, parent *v1alpha1.Config[v1alpha1.LokiOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Config.Host == "" && parent.Config.Host == "" {
		return nil
	}

	setFallback(&config.Config.Path, "/loki/api/v1/push")
	setFallback(&config.Config.Host, parent.Config.Host)
	setFallback(&config.Config.Certificate, parent.Config.Certificate)
	setFallback(&config.Config.Path, parent.Config.Path)
	setFallback(&config.Config.Username, parent.Config.Username)
	setFallback(&config.Config.Password, parent.Config.Password)
	setBool(&config.Config.SkipTLS, parent.Config.SkipTLS)

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Loki,
		Config:       config,
		ParentConfig: parent,
		Client: loki.NewClient(loki.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Host:         config.Config.Host + config.Config.Path,
			CustomFields: config.CustomFields,
			Username:     config.Config.Username,
			Password:     config.Config.Password,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
			Headers:      config.Config.Headers,
		}),
	}
}

func (f *TargetFactory) CreateElasticsearchTarget(config, parent *v1alpha1.Config[v1alpha1.ElasticsearchOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
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
	setBool(&config.Config.TypelessAPI, parent.Config.TypelessAPI)

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Elasticsearch,
		Config:       config,
		ParentConfig: parent,
		Client: elasticsearch.NewClient(elasticsearch.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Host:         config.Config.Host,
			Username:     config.Config.Username,
			Password:     config.Config.Password,
			ApiKey:       config.Config.APIKey,
			Rotation:     config.Config.Rotation,
			Index:        config.Config.Index,
			TypelessApi:  config.Config.TypelessAPI,
			CustomFields: config.CustomFields,
			Headers:      config.Config.Headers,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateDiscordTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Discord,
		Config:       config,
		ParentConfig: parent,
		Client: discord.NewClient(discord.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Webhook:      config.Config.Webhook,
			CustomFields: config.CustomFields,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateTeamsTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Teams,
		Config:       config,
		ParentConfig: parent,
		Client: teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Webhook:      config.Config.Webhook,
			CustomFields: config.CustomFields,
			Headers:      config.Config.Headers,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateWebhookTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Webhook,
		Config:       config,
		ParentConfig: parent,
		Client: webhook.NewClient(webhook.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Host:         config.Config.Webhook,
			Headers:      config.Config.Headers,
			CustomFields: config.CustomFields,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateTelegramTarget(config, parent *v1alpha1.Config[v1alpha1.TelegramOptions]) *target.Target {
	if config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
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

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Telegram,
		Config:       config,
		ParentConfig: parent,
		Client: telegram.NewClient(telegram.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Host:         fmt.Sprintf("%s/bot%s/sendMessage", host, config.Config.Token),
			ChatID:       config.Config.ChatID,
			Headers:      config.Config.Headers,
			CustomFields: config.CustomFields,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateGoogleChatTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	mapWebhookTarget(config, parent)

	if config.Config.Webhook == "" {
		return nil
	}

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.GoogleChat,
		Config:       config,
		ParentConfig: parent,
		Client: googlechat.NewClient(googlechat.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Webhook:      config.Config.Webhook,
			Headers:      config.Config.Headers,
			CustomFields: config.CustomFields,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateJiraTarget(config, parent *v1alpha1.Config[v1alpha1.JiraOptions]) *target.Target {
	if config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	if config.Config.Host == "" {
		return nil
	}

	setFallback(&config.Config.Certificate, parent.Config.Certificate)
	setBool(&config.Config.SkipTLS, parent.Config.SkipTLS)
	setFallback(&config.Config.Username, parent.Config.Username)
	setFallback(&config.Config.Password, parent.Config.Password)
	setFallback(&config.Config.APIToken, parent.Config.APIToken)
	setFallback(&config.Config.ProjectKey, parent.Config.ProjectKey)
	setFallback(&config.Config.IssueType, parent.Config.IssueType)

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Jira,
		Config:       config,
		ParentConfig: parent,
		Client: jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Host:         config.Config.Host,
			Username:     config.Config.Username,
			Password:     config.Config.Password,
			APIToken:     config.Config.APIToken,
			ProjectKey:   config.Config.ProjectKey,
			IssueType:    config.Config.IssueType,
			SkipTLS:      config.Config.SkipTLS,
			Certificate:  config.Config.Certificate,
			CustomFields: config.CustomFields,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) CreateS3Target(config, parent *v1alpha1.Config[v1alpha1.S3Options]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Config.Bucket, parent.Config.Bucket)
	if config.Config.Bucket == "" {
		return nil
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

	setFallback(&config.Config.Region, os.Getenv("AWS_REGION"))
	setFallback(&config.Config.Prefix, parent.Config.Prefix, "policy-reporter")
	setFallback(&config.Config.KmsKeyID, parent.Config.KmsKeyID)
	setFallback(&config.Config.ServerSideEncryption, parent.Config.ServerSideEncryption)
	setBool(&config.Config.BucketKeyEnabled, parent.Config.BucketKeyEnabled)

	config.MapBaseParent(parent)

	s3Client := aws.NewS3Client(
		config.Config.AccessKeyID,
		config.Config.SecretAccessKey,
		config.Config.Region,
		config.Config.Endpoint,
		config.Config.Bucket,
		config.Config.PathStyle,
		aws.WithKMS(config.Config.BucketKeyEnabled, &config.Config.KmsKeyID, &config.Config.ServerSideEncryption),
	)

	sugar.Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.S3,
		Config:       config,
		ParentConfig: parent,
		Client: s3.NewClient(s3.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			S3:           s3Client,
			CustomFields: config.CustomFields,
			Prefix:       config.Config.Prefix,
		}),
	}
}

func (f *TargetFactory) CreateKinesisTarget(config, parent *v1alpha1.Config[v1alpha1.KinesisOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
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

	setFallback(&config.Config.Region, os.Getenv("AWS_REGION"))

	config.MapBaseParent(parent)

	kinesisClient := aws.NewKinesisClient(
		config.Config.AccessKeyID,
		config.Config.SecretAccessKey,
		config.Config.Region,
		config.Config.Endpoint,
		config.Config.StreamName,
	)

	sugar.Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.Kinesis,
		Config:       config,
		ParentConfig: parent,
		Client: kinesis.NewClient(kinesis.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			CustomFields: config.CustomFields,
			Kinesis:      kinesisClient,
		}),
	}
}

func (f *TargetFactory) CreateSecurityHubTarget(config, parent *v1alpha1.Config[v1alpha1.SecurityHubOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
	}

	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(config, config.SecretRef, config.MountedSecret)
	}

	setFallback(&config.Config.AccountID, parent.Config.AccountID)
	if config.Config.AccountID == "" {
		return nil
	}
	setFallback(&config.Config.Region, os.Getenv("AWS_REGION"))

	sugar := zap.S()
	if err := checkAWSConfig(config.Name, config.Config.AWSConfig, parent.Config.AWSConfig); err != nil {
		sugar.Error(err)

		return nil
	}

	config.Config.MapAWSParent(parent.Config.AWSConfig)
	config.MapBaseParent(parent)

	setFallback(&config.Config.ProductName, parent.Config.ProductName, "Policy Reporter")
	setFallback(&config.Config.CompanyName, parent.Config.CompanyName, "Kyverno")
	setInt(&config.Config.DelayInSeconds, parent.Config.DelayInSeconds)

	client := aws.NewHubClient(
		config.Config.AccessKeyID,
		config.Config.SecretAccessKey,
		config.Config.Region,
		config.Config.Endpoint,
	)

	zap.L().Info(config.Name+" configured", zap.Bool("synchronize", config.Config.Synchronize))

	hub := securityhub.NewClient(securityhub.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
			ReportFilter:          createReportFilter(config.Filter),
		},
		CustomFields: config.CustomFields,
		Client:       client,
		AccountID:    config.Config.AccountID,
		ProductName:  config.Config.ProductName,
		CompanyName:  config.Config.CompanyName,
		Region:       config.Config.Region,
		Delay:        time.Duration(config.Config.DelayInSeconds) * time.Second,
		Synchronize:  config.Config.Synchronize,
	})

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.SecurityHub,
		Config:       config,
		ParentConfig: parent,
		Client:       hub,
	}
}

func (f *TargetFactory) CreateGCSTarget(config, parent *v1alpha1.Config[v1alpha1.GCSOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
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

	gcsClient := gs.NewClient(
		context.Background(),
		config.Config.Credentials,
		config.Config.Bucket,
	)
	if gcsClient == nil {
		return nil
	}

	sugar.Infof("%s configured", config.Name)
	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.GCS,
		Config:       config,
		ParentConfig: parent,
		Client: gcs.NewClient(gcs.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Client:       gcsClient,
			CustomFields: config.CustomFields,
			Prefix:       config.Config.Prefix,
		}),
	}
}

func (f *TargetFactory) CreateAlertManagerTarget(config, parent *v1alpha1.Config[v1alpha1.AlertManagerOptions]) *target.Target {
	if config == nil || config.Config == nil {
		return nil
	}

	if (parent.SecretRef != "" && f.secretClient != nil) || parent.MountedSecret != "" {
		f.mapSecretValues(parent, parent.SecretRef, parent.MountedSecret)
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

	config.MapBaseParent(parent)

	zap.S().Infof("%s configured", config.Name)

	return &target.Target{
		ID:           uuid.NewString(),
		Type:         target.AlertManager,
		Config:       config,
		ParentConfig: parent,
		Client: alertmanager.NewClient(alertmanager.Options{
			ClientOptions: target.ClientOptions{
				Name:                  config.Name,
				SkipExistingOnStartup: config.SkipExisting,
				ResultFilter:          f.createResultFilter(config.Filter, config.MinimumSeverity, config.Sources),
				ReportFilter:          createReportFilter(config.Filter),
			},
			Host:         config.Config.Host,
			Headers:      config.Config.Headers,
			CustomFields: config.CustomFields,
			HTTPClient:   http.NewClient(config.Config.Certificate, config.Config.SkipTLS),
		}),
	}
}

func (f *TargetFactory) createResultFilter(filter filters.Filter, minimumSeverity string, sources []string) *report.ResultFilter {
	sourceFilter := filter.Sources
	if len(sources) > 0 {
		sourceFilter = filters.ValueFilter{Include: sources}
	}

	return f.filterFactory.CreateFilter(
		validate.RuleSets{
			Include:  filter.Namespaces.Include,
			Exclude:  filter.Namespaces.Exclude,
			Selector: filter.Namespaces.Selector,
		},
		ToRuleSet(filter.Severities),
		ToRuleSet(filter.Status),
		ToRuleSet(filter.Policies),
		ToRuleSet(sourceFilter),
		minimumSeverity,
	)
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
	case *v1alpha1.Config[v1alpha1.LokiOptions]:
		if values.Host != "" {
			c.Config.Host = values.Host
		}

	case *v1alpha1.Config[v1alpha1.SlackOptions]:
		if values.Webhook != "" {
			c.Config.Webhook = values.Webhook
		}
		if values.Channel != "" {
			c.Config.Channel = values.Channel
		}

	case *v1alpha1.Config[v1alpha1.WebhookOptions]:
		if values.Webhook != "" {
			c.Config.Webhook = values.Webhook
		}
		if values.Token != "" {
			if c.Config.Headers == nil {
				c.Config.Headers = make(map[string]string)
			}

			c.Config.Headers["Authorization"] = values.Token
		}

	case *v1alpha1.Config[v1alpha1.ElasticsearchOptions]:
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

	case *v1alpha1.Config[v1alpha1.S3Options]:
		if values.AccessKeyID != "" {
			c.Config.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.Config.SecretAccessKey = values.SecretAccessKey
		}
		if values.KmsKeyID != "" {
			c.Config.KmsKeyID = values.KmsKeyID
		}

	case *v1alpha1.Config[v1alpha1.KinesisOptions]:
		if values.AccessKeyID != "" {
			c.Config.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.Config.SecretAccessKey = values.SecretAccessKey
		}

	case *v1alpha1.Config[v1alpha1.SecurityHubOptions]:
		if values.AccessKeyID != "" {
			c.Config.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.Config.SecretAccessKey = values.SecretAccessKey
		}
		if values.AccountID != "" {
			c.Config.AccountID = values.AccountID
		}

	case *v1alpha1.Config[v1alpha1.GCSOptions]:
		if values.Credentials != "" {
			c.Config.Credentials = values.Credentials
		}

	case *v1alpha1.Config[v1alpha1.TelegramOptions]:
		if values.Token != "" {
			c.Config.Token = values.Token
		}
		if values.Webhook != "" {
			c.Config.Webhook = values.Webhook
		}
		if values.Host != "" {
			c.Config.Webhook = values.Host
		}
	}
}

// what are those parameters ?
func NewFactory(secretClient secrets.Client, filterFactory *target.ResultFilterFactory) target.Factory {
	return &TargetFactory{secretClient: secretClient, filterFactory: filterFactory}
}

func mapWebhookTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) {
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
	irsaARN := os.Getenv("AWS_ROLE_ARN")
	irsaFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	podIdentityFile := os.Getenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE")
	podIdentityURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")

	return (irsaARN != "" && irsaFile != "") || (podIdentityFile != "" && podIdentityURI != "")
}

func checkAWSConfig(name string, config v1alpha1.AWSConfig, parent v1alpha1.AWSConfig) error {
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

func setInt(config *int, parent int) {
	if *config == 0 {
		*config = parent
	}
}

func createReportFilter(filter filters.Filter) *report.ReportFilter {
	return target.NewReportFilter(
		ToRuleSet(filter.ReportLabels),
		ToRuleSet(filter.Sources),
	)
}

func ToRuleSet(filter filters.ValueFilter) validate.RuleSets {
	return validate.RuleSets{
		Include: filter.Include,
		Exclude: filter.Exclude,
	}
}

func createConfig[T any](tc *v1alpha1.TargetConfig, config *T) *v1alpha1.Config[T] {
	return &v1alpha1.Config[T]{
		Name:            tc.Spec.Name,
		MinimumSeverity: tc.Spec.MinimumSeverity,
		Filter:          tc.Spec.Filter,
		SecretRef:       tc.Spec.SecretRef,
		SkipExisting:    true, // todo: introduce the skip existing feature
		CustomFields:    tc.Spec.CustomFields,
		MountedSecret:   tc.Spec.MountedSecret,
		Sources:         tc.Spec.Sources,
		Config:          config,
	}
}
