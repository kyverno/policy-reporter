package config

import (
	"context"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
	"github.com/kyverno/policy-reporter/pkg/target/http"
	"github.com/kyverno/policy-reporter/pkg/target/kinesis"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
	"github.com/kyverno/policy-reporter/pkg/target/ui"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

// TargetFactory manages target creation
type TargetFactory struct {
	secretClient secrets.Client
	namespace    string
}

// LokiClients resolver method
func (f *TargetFactory) LokiClients(config Loki, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Loki"
	}
	if config.Path == "" {
		config.Path = "/api/prom/push"
	}

	if loki := f.createLokiClient(config, Loki{}, logger); loki != nil {
		clients = append(clients, loki)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Loki Channel %d", i+1)
		}

		if loki := f.createLokiClient(channel, config, logger); loki != nil {
			clients = append(clients, loki)
		}
	}

	return clients
}

// ElasticsearchClients resolver method
func (f *TargetFactory) ElasticsearchClients(config Elasticsearch, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Elasticsearch"
	}

	if es := f.createElasticsearchClient(config, Elasticsearch{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Elasticsearch Channel %d", i+1)
		}

		if es := f.createElasticsearchClient(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// SlackClients resolver method
func (f *TargetFactory) SlackClients(config Slack, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Slack"
	}

	if es := f.createSlackClient(config, Slack{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Slack Channel %d", i+1)
		}

		if es := f.createSlackClient(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// DiscordClients resolver method
func (f *TargetFactory) DiscordClients(config Discord, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Discord"
	}

	if es := f.createDiscordClient(config, Discord{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Discord Channel %d", i+1)
		}

		if es := f.createDiscordClient(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// TeamsClients resolver method
func (f *TargetFactory) TeamsClients(config Teams, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Teams"
	}

	if es := f.createTeamsClient(config, Teams{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Teams Channel %d", i+1)
		}

		if es := f.createTeamsClient(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// WebhookClients resolver method
func (f *TargetFactory) WebhookClients(config Webhook, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Webhook"
	}

	if es := f.createWebhookClient(config, Webhook{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Webhook Channel %d", i+1)
		}

		if es := f.createWebhookClient(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// UIClient resolver method
func (f *TargetFactory) UIClient(config UI, logger *zap.Logger) target.Client {
	if config.Host == "" {
		return nil
	}

	logger.Info("UI configured")

	return ui.NewClient(ui.Options{
		ClientOptions: target.ClientOptions{
			Name:                  "UI",
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(TargetFilter{}, config.MinimumPriority, config.Sources),
		},
		Host:       config.Host,
		HTTPClient: http.NewClient(config.Certificate, config.SkipTLS),
	})
}

// S3Clients resolver method
func (f *TargetFactory) S3Clients(config S3, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "S3"
	}

	if es := f.createS3Client(config, S3{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("S3 Channel %d", i+1)
		}

		if es := f.createS3Client(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

// KinesisClients resolver method
func (f *TargetFactory) KinesisClients(config Kinesis, logger *zap.Logger) []target.Client {
	clients := make([]target.Client, 0)
	if config.Name == "" {
		config.Name = "Kinesis"
	}

	if es := f.createKinesisClient(config, Kinesis{}, logger); es != nil {
		clients = append(clients, es)
	}
	for i, channel := range config.Channels {
		if channel.Name == "" {
			channel.Name = fmt.Sprintf("Kinesis Channel %d", i+1)
		}

		if es := f.createKinesisClient(channel, config, logger); es != nil {
			clients = append(clients, es)
		}
	}

	return clients
}

func (f *TargetFactory) createSlackClient(config Slack, parent Slack, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	logger.Sugar().Infof("%s configured", config.Name)

	return slack.NewClient(slack.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		Webhook:      config.Webhook,
		CustomFields: config.CustomFields,
		HTTPClient:   http.NewClient("", false),
	})
}

func (f *TargetFactory) createLokiClient(config Loki, parent Loki, logger *zap.Logger) target.Client {
	if config.SecretRef != "" {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Host == "" && parent.Host == "" {
		return nil
	} else if config.Host == "" {
		config.Host = parent.Host
	}

	if config.Certificate == "" {
		config.Certificate = parent.Certificate
	}

	if !config.SkipTLS {
		config.SkipTLS = parent.SkipTLS
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if config.Path == "" {
		config.Path = parent.Path
	}

	logger.Sugar().Infof("%s configured", config.Name)

	return loki.NewClient(loki.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
		},
		Host:         config.Host + config.Path,
		CustomLabels: config.CustomLabels,
		HTTPClient:   http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createElasticsearchClient(config Elasticsearch, parent Elasticsearch, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Host == "" && parent.Host == "" {
		return nil
	} else if config.Host == "" {
		config.Host = parent.Host
	}

	if config.Certificate == "" {
		config.Certificate = parent.Certificate
	}

	if !config.SkipTLS {
		config.SkipTLS = parent.SkipTLS
	}

	if config.Username == "" {
		config.Username = parent.Username
	}

	if config.Password == "" {
		config.Password = parent.Password
	}

	if config.Index == "" && parent.Index == "" {
		config.Index = "policy-reporter"
	} else if config.Index == "" {
		config.Index = parent.Index
	}

	if config.Rotation == "" && parent.Rotation == "" {
		config.Rotation = elasticsearch.Daily
	} else if config.Rotation == "" {
		config.Rotation = parent.Rotation
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	logger.Sugar().Infof("%s configured", config.Name)

	return elasticsearch.NewClient(elasticsearch.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		Host:         config.Host,
		Username:     config.Username,
		Password:     config.Password,
		Rotation:     config.Rotation,
		Index:        config.Index,
		CustomFields: config.CustomFields,
		HTTPClient:   http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createDiscordClient(config Discord, parent Discord, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Webhook == "" {
		return nil
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	logger.Sugar().Infof("%s configured", config.Name)

	return discord.NewClient(discord.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		Webhook:      config.Webhook,
		CustomFields: config.CustomFields,
		HTTPClient:   http.NewClient("", false),
	})
}

func (f *TargetFactory) createTeamsClient(config Teams, parent Teams, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Webhook == "" {
		return nil
	}

	if config.Certificate == "" {
		config.Certificate = parent.Certificate
	}

	if !config.SkipTLS {
		config.SkipTLS = parent.SkipTLS
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	if !config.SkipTLS {
		config.SkipTLS = parent.SkipTLS
	}

	logger.Sugar().Infof("%s configured", config.Name)

	return teams.NewClient(teams.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		Webhook:      config.Webhook,
		CustomFields: config.CustomFields,
		HTTPClient:   http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createWebhookClient(config Webhook, parent Webhook, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Host == "" {
		return nil
	}

	if config.Certificate == "" {
		config.Certificate = parent.Certificate
	}

	if !config.SkipTLS {
		config.SkipTLS = parent.SkipTLS
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

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

	logger.Sugar().Infof("%s configured", config.Name)

	return webhook.NewClient(webhook.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		Host:         config.Host,
		Headers:      config.Headers,
		CustomFields: config.CustomFields,
		HTTPClient:   http.NewClient(config.Certificate, config.SkipTLS),
	})
}

func (f *TargetFactory) createS3Client(config S3, parent S3, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Endpoint == "" && parent.Endpoint == "" {
		return nil
	} else if config.Endpoint == "" {
		config.Endpoint = parent.Endpoint
	}

	sugar := logger.Sugar()

	if config.AccessKeyID == "" && parent.AccessKeyID == "" {
		sugar.Errorf("%s.AccessKeyID has not been declared", config.Name)
		return nil
	} else if config.AccessKeyID == "" {
		config.AccessKeyID = parent.AccessKeyID
	}

	if config.SecretAccessKey == "" && parent.SecretAccessKey == "" {
		sugar.Errorf("%s.SecretAccessKey has not been declared", config.Name)
		return nil
	} else if config.SecretAccessKey == "" {
		config.SecretAccessKey = parent.SecretAccessKey
	}

	if config.Region == "" && parent.Region == "" {
		sugar.Errorf("%s.Region has not been declared", config.Name)
		return nil
	} else if config.Region == "" {
		config.Region = parent.Region
	}

	if config.Bucket == "" && parent.Bucket == "" {
		sugar.Errorf("%s.Bucket has not been declared", config.Name)
		return nil
	} else if config.Bucket == "" {
		config.Bucket = parent.Bucket
	}

	if config.Prefix == "" && parent.Prefix == "" {
		config.Prefix = "policy-reporter"
	} else if config.Prefix == "" {
		config.Prefix = parent.Prefix
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	s3Client := helper.NewS3Client(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.Bucket,
		config.PathStyle,
	)

	sugar.Infof("%s configured", config.Name)

	return s3.NewClient(s3.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		S3:           s3Client,
		CustomFields: config.CustomFields,
		Prefix:       config.Prefix,
	})
}

func (f *TargetFactory) createKinesisClient(config Kinesis, parent Kinesis, logger *zap.Logger) target.Client {
	if config.SecretRef != "" && f.secretClient != nil {
		f.mapSecretValues(&config, config.SecretRef, logger)
	}

	if config.Endpoint == "" && parent.Endpoint == "" {
		return nil
	} else if config.Endpoint == "" {
		config.Endpoint = parent.Endpoint
	}

	sugar := logger.Sugar()

	if config.AccessKeyID == "" && parent.AccessKeyID == "" {
		sugar.Errorf("%s.AccessKeyID has not been declared", config.Name)
		return nil
	} else if config.AccessKeyID == "" {
		config.AccessKeyID = parent.AccessKeyID
	}

	if config.SecretAccessKey == "" && parent.SecretAccessKey == "" {
		sugar.Errorf("%s.SecretAccessKey has not been declared", config.Name)
		return nil
	} else if config.SecretAccessKey == "" {
		config.SecretAccessKey = parent.SecretAccessKey
	}

	if config.Region == "" && parent.Region == "" {
		sugar.Errorf("%s.Region has not been declared", config.Name)
		return nil
	} else if config.Region == "" {
		config.Region = parent.Region
	}

	if config.StreamName == "" && parent.StreamName == "" {
		sugar.Errorf("%s.StreamName has not been declared", config.Name)
		return nil
	} else if config.StreamName == "" {
		config.StreamName = parent.StreamName
	}

	if config.MinimumPriority == "" {
		config.MinimumPriority = parent.MinimumPriority
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}

	kinesisClient := helper.NewKinesisClient(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.Region,
		config.Endpoint,
		config.StreamName,
	)

	sugar.Infof("%s configured", config.Name)

	return kinesis.NewClient(kinesis.Options{
		ClientOptions: target.ClientOptions{
			Name:                  config.Name,
			SkipExistingOnStartup: config.SkipExisting,
			ResultFilter:          createResultFilter(config.Filter, config.MinimumPriority, config.Sources),
			ReportFilter:          createReprotFilter(config.Filter),
			Logger:                logger,
		},
		CustomFields: config.CustomFields,
		Kinesis:      kinesisClient,
	})
}

func (f *TargetFactory) mapSecretValues(config any, ref string, logger *zap.Logger) {
	values, err := f.secretClient.Get(context.Background(), ref)
	if err != nil {
		logger.Warn("failed to get secret reference", zap.Error(err))
		return
	}

	switch c := config.(type) {
	case *Loki:
		if values.Host != "" {
			c.Host = values.Host
		}

	case *Slack:
		if values.Webhook != "" {
			c.Webhook = values.Webhook
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

	case *Kinesis:
		if values.AccessKeyID != "" {
			c.AccessKeyID = values.AccessKeyID
		}
		if values.SecretAccessKey != "" {
			c.SecretAccessKey = values.SecretAccessKey
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

func createReprotFilter(filter TargetFilter) *report.ReportFilter {
	return target.NewReportFilter(
		ToRuleSet(filter.ReportLabels),
	)
}

func NewTargetFactory(namespace string, secretClient secrets.Client) *TargetFactory {
	return &TargetFactory{namespace: namespace, secretClient: secretClient}
}
