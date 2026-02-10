package config_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var targets = target.Targets{
	Loki: &targetconfig.Config[v1alpha1.LokiOptions]{
		Config: &v1alpha1.LokiOptions{
			HostOptions: v1alpha1.HostOptions{
				Host:    "http://localhost:3100",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.LokiOptions]{
			{
				CustomFields: map[string]string{"label2": "value2"},
			},
		},
	},
	Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{
		Config: &v1alpha1.ElasticsearchOptions{
			HostOptions: v1alpha1.HostOptions{
				Host:    "http://localhost:9200",
				SkipTLS: true,
			},
			Index:    "policy-reporter",
			Rotation: "daily",
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.ElasticsearchOptions]{{}},
	},
	Slack: &targetconfig.Config[v1alpha1.SlackOptions]{
		Config: &v1alpha1.SlackOptions{
			WebhookOptions: v1alpha1.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.SlackOptions]{{
			Config: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:9200",
				},
			},
		}, {
			Config: &v1alpha1.SlackOptions{
				Channel: "general",
			},
		}},
	},
	Discord: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://discord:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.WebhookOptions]{{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	Teams: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://hook.teams:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.WebhookOptions]{{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	GoogleChat: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://localhost:900/webhook",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.WebhookOptions]{{}},
	},
	Telegram: &targetconfig.Config[v1alpha1.TelegramOptions]{
		Config: &v1alpha1.TelegramOptions{
			WebhookOptions: v1alpha1.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
			Token:  "XXX",
			ChatID: "123456",
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.TelegramOptions]{{
			Config: &v1alpha1.TelegramOptions{
				ChatID: "1234567",
			},
		}},
	},
	Webhook: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://localhost:8080",
			SkipTLS: true,
			Headers: map[string]string{
				"X-Custom": "Header",
			},
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.WebhookOptions]{{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:8081",
				Headers: map[string]string{
					"X-Custom-2": "Header",
				},
			},
		}},
	},
	S3: &targetconfig.Config[v1alpha1.S3Options]{
		Config: &v1alpha1.S3Options{
			AWSConfig: v1alpha1.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			Bucket:               "test",
			BucketKeyEnabled:     false,
			KmsKeyID:             "",
			ServerSideEncryption: "",
			PathStyle:            true,
			Prefix:               "prefix",
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.S3Options]{{}},
	},
	Kinesis: &targetconfig.Config[v1alpha1.KinesisOptions]{
		Config: &v1alpha1.KinesisOptions{
			AWSConfig: v1alpha1.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			StreamName: "policy-reporter",
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.KinesisOptions]{{}},
	},
	SecurityHub: &targetconfig.Config[v1alpha1.SecurityHubOptions]{
		Config: &v1alpha1.SecurityHubOptions{
			AWSConfig: v1alpha1.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			AccountID: "AccountId",
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.SecurityHubOptions]{{}},
	},
	GCS: &targetconfig.Config[v1alpha1.GCSOptions]{
		Config: &v1alpha1.GCSOptions{
			Credentials: `{"token": "token", "type": "authorized_user"}`,
			Bucket:      "test",
			Prefix:      "prefix",
		},
		SkipExisting:    true,
		MinimumSeverity: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.GCSOptions]{{}},
	},
}

var testConfig = &config.Config{
	Templates: config.Templates{
		Dir: "../../templates",
	},
	EmailReports: config.EmailReports{
		SMTP: config.SMTP{
			Host:       "localhost",
			Port:       465,
			Username:   "policy-reporter@kyverno.io",
			Password:   "password",
			From:       "policy-reporter@kyverno.io",
			Encryption: "ssl/tls",
		},
	},
	Targets: targets,
	Logging: config.Logging{
		Development: true,
	},
	SourceConfig: []config.SourceConfig{
		{
			Selector: config.SourceSelector{
				Source: "test",
			},
			CustomID: config.CustomID{
				Enabled: true,
				Fields:  []string{"resource"},
			},
		},
		{
			Selector: config.SourceSelector{
				Source: "default",
			},
		},
	},
}

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	assert.Equal(t, resolver.TargetClients().Length(), 25)
}

func Test_ResolveHasTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	assert.True(t, resolver.HasTargets())
}

func Test_ResolveSkipExistingOnStartup(t *testing.T) {
	testConfig := &config.Config{
		Targets: target.Targets{
			Loki: &targetconfig.Config[v1alpha1.LokiOptions]{
				Config: &v1alpha1.LokiOptions{
					HostOptions: v1alpha1.HostOptions{
						Host: "http://localhost:3100",
					},
				},
				SkipExisting:    true,
				MinimumSeverity: "debug",
			},
			Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{
				Config: &v1alpha1.ElasticsearchOptions{
					HostOptions: v1alpha1.HostOptions{
						Host: "http://localhost:9200",
					},
				},
				SkipExisting:    true,
				MinimumSeverity: "debug",
			},
		},
	}

	t.Run("Resolve false", func(t *testing.T) {
		testConfig.Targets.Elasticsearch.SkipExisting = false

		resolver := config.NewResolver(testConfig, &rest.Config{})

		assert.False(t, resolver.SkipExistingOnStartup(), "Expected SkipExistingOnStartup to be false if one Client has SkipExistingOnStartup false configured")
	})

	t.Run("Resolve true", func(t *testing.T) {
		testConfig.Targets.Elasticsearch.SkipExisting = true

		resolver := config.NewResolver(testConfig, &rest.Config{})

		assert.True(t, resolver.SkipExistingOnStartup(), "Expected SkipExistingOnStartup to be true if all Client has SkipExistingOnStartup true configured")
	})
}

func Test_ResolvePolicyClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})

	client1, err := resolver.OpenReportsClient()
	assert.Nil(t, err)

	client2, _ := resolver.OpenReportsClient()

	assert.Equal(t, client1, client2, "A second call resolver.PolicyReportClient() should return the cached first client")
}

func Test_ResolveSecretInformer(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})

	informer, err := resolver.SecretInformer()
	assert.Nil(t, err)
	assert.NotNil(t, informer)
}

func Test_ResolveSecretInformerWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.SecretInformer()
	assert.NotNil(t, err, "Error: 'host must be a URL or a host:port pair' was expected")
}

func Test_ResolveLeaderElectionClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})

	client1, err := resolver.LeaderElectionClient()
	assert.Nil(t, err)

	client2, _ := resolver.LeaderElectionClient()

	assert.Equal(t, client1, client2, "A second call resolver.LeaderElectionClient() should return the cached first client")
}

func Test_ResolvePolicyStore(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})
	db := resolver.Database()
	defer db.Close()

	store1, err := resolver.Store(db)
	assert.Nil(t, err)

	store2, _ := resolver.Store(db)
	assert.Equal(t, store1, store2, "A second call resolver.Store() should return the cached first client")
}

func Test_ResolveAPIServer(t *testing.T) {
	resolver := config.NewResolver(&config.Config{
		API: config.API{
			BasicAuth: config.BasicAuth{Username: "user", Password: "password"},
		},
	}, &rest.Config{})

	server, _ := resolver.Server(context.Background(), nil)
	assert.NotNil(t, server)
}

func Test_ResolveCache(t *testing.T) {
	t.Run("InMemory", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})

		cache1 := resolver.ResultCache()
		assert.NotNil(t, cache1)

		assert.Equal(t, cache1, resolver.ResultCache(), "A second call resolver.ResultCache() should return the cached first client")
	})

	t.Run("Redis", func(t *testing.T) {
		redisConfig := &config.Config{
			Redis: config.Redis{
				Enabled: true,
				Address: "localhost:6379",
			},
		}

		resolver := config.NewResolver(redisConfig, &rest.Config{})

		assert.NotNil(t, resolver.ResultCache())
	})

	t.Run("RedisWithSkipTLS", func(t *testing.T) {
		redisConfig := &config.Config{
			Redis: config.Redis{
				Enabled: true,
				Address: "localhost:6380",
				SkipTLS: true,
			},
		}

		resolver := config.NewResolver(redisConfig, &rest.Config{})

		assert.NotNil(t, resolver.ResultCache())
	})

	t.Run("RedisWithCertificate", func(t *testing.T) {
		redisConfig := &config.Config{
			Redis: config.Redis{
				Enabled:     true,
				Address:     "localhost:6380",
				Certificate: "/tmp/non-existing-cert.pem",
			},
		}

		resolver := config.NewResolver(redisConfig, &rest.Config{})

		assert.NotNil(t, resolver.ResultCache())
	})
}

func Test_ResolveReportFilter(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	assert.NotNil(t, resolver.ReportFilter())
}

func Test_ResolveClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.OpenReportsClient()
	assert.NotNil(t, err, "Error: 'host must be a URL or a host:port pair' was expected")
}

func Test_ResolveLeaderElectionWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.LeaderElectionClient()
	assert.NotNil(t, err, "Error: 'host must be a URL or a host:port pair' was expected")
}

func Test_ResolveCRDClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.OpenreportsCRClient()
	assert.NotNil(t, err, "Error: 'host must be a URL or a host:port pair' was expected")
}

func Test_ResolveSecretClient(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	assert.NotNil(t, resolver.SecretClient())
}

func Test_ResolveSecretCClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	client := resolver.SecretClient()
	assert.Nil(t, client, "Error: 'host must be a URL or a host:port pair' was expected")
}

func Test_RegisterStoreListener(t *testing.T) {
	t.Run("Register StoreListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterStoreListener(context.Background(), report.NewPolicyReportStore())

		assert.Len(t, resolver.EventPublisher().GetListener(), 1, "Expected one Listener to be registered")
	})
}

func Test_RegisterMetricsListener(t *testing.T) {
	t.Run("Register MetricsListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterMetricsListener()

		assert.Len(t, resolver.EventPublisher().GetListener(), 1, "Expected one Listener to be registered")
	})
}

func Test_RegisterSendResultListener(t *testing.T) {
	t.Run("Register SendResultListener with Targets", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.Logger()
		resolver.RegisterSendResultListener()

		assert.Len(t, resolver.EventPublisher().GetListener(), 1, "Expected one Listener to be registered")
	})
}

func Test_SMTP(t *testing.T) {
	t.Run("SMTP", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})

		assert.NotNil(t, resolver.SMTPServer())
	})
	t.Run("EmailClient", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})

		assert.NotNil(t, resolver.EmailClient())
	})
}

func Test_ResolveLogger(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	logger1, _ := resolver.Logger()
	assert.NotNil(t, logger1)

	logger2, _ := resolver.Logger()
	assert.NotNil(t, logger2)

	assert.Equal(t, logger1, logger2, "A second call resolver.Logger() should return the cached first cache")
}

func Test_ResolveEnableLeaderElection(t *testing.T) {
	t.Run("general disabled", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: false},
			Targets: target.Targets{
				Loki: &targetconfig.Config[v1alpha1.LokiOptions]{
					Config: &v1alpha1.LokiOptions{
						HostOptions: v1alpha1.HostOptions{
							Host: "http://localhost:3100",
						},
					},
				},
			},
			Database: config.Database{Type: database.MySQL},
		}, &rest.Config{})

		assert.False(t, resolver.EnableLeaderElection(), "leaderelection should be not enabled if its general disabled")
	})

	t.Run("no pushes and SQLite Database", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: true},
			Database:       config.Database{Type: database.SQLite},
			DBFile:         "test.db",
		}, &rest.Config{})

		assert.False(t, resolver.EnableLeaderElection(), "leaderelection should be not enabled if no pushes configured and SQLite is used")
	})

	t.Run("enabled if pushes defined", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: true},
			Database:       config.Database{Type: database.SQLite},
			Targets: target.Targets{
				Loki: &targetconfig.Config[v1alpha1.LokiOptions]{
					Config: &v1alpha1.LokiOptions{
						HostOptions: v1alpha1.HostOptions{
							Host: "http://localhost:3100",
						},
					},
				},
			},
			DBFile: "test.db",
		}, &rest.Config{})

		assert.True(t, resolver.EnableLeaderElection(), "leaderelection should be enabled if general enabled and targets configured")
	})
}

func Test_ResolveCustomIDGenerators(t *testing.T) {
	resolver := config.NewResolver(testConfig, nil)

	generators := resolver.ReconditionerConfigs()
	assert.Len(t, generators, 1, "only enabled custom id config should be mapped")
}

func Test_ResolveTargetCollection(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	collection := resolver.TargetClients()
	assert.NotNil(t, collection)

	assert.Equal(t, collection, resolver.TargetClients(), "A second call resolver.TargetClients() should return the cached first cache")
}
