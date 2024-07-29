package config_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var targets = target.Targets{
	Loki: &target.Config[target.LokiOptions]{
		Config: &target.LokiOptions{
			HostOptions: target.HostOptions{
				Host:    "http://localhost:3100",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.LokiOptions]{
			{
				CustomFields: map[string]string{"label2": "value2"},
			},
		},
	},
	Elasticsearch: &target.Config[target.ElasticsearchOptions]{
		Config: &target.ElasticsearchOptions{
			HostOptions: target.HostOptions{
				Host:    "http://localhost:9200",
				SkipTLS: true,
			},
			Index:    "policy-reporter",
			Rotation: "daily",
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.ElasticsearchOptions]{{}},
	},
	Slack: &target.Config[target.SlackOptions]{
		Config: &target.SlackOptions{
			WebhookOptions: target.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.SlackOptions]{{
			Config: &target.SlackOptions{
				WebhookOptions: target.WebhookOptions{
					Webhook: "http://localhost:9200",
				},
			},
		}, {
			Config: &target.SlackOptions{
				Channel: "general",
			},
		}},
	},
	Discord: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://discord:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.WebhookOptions]{{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	Teams: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://hook.teams:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.WebhookOptions]{{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	GoogleChat: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://localhost:900/webhook",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.WebhookOptions]{{}},
	},
	Telegram: &target.Config[target.TelegramOptions]{
		Config: &target.TelegramOptions{
			WebhookOptions: target.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
			Token:  "XXX",
			ChatID: "123456",
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.TelegramOptions]{{
			Config: &target.TelegramOptions{
				ChatID: "1234567",
			},
		}},
	},
	Webhook: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://localhost:8080",
			SkipTLS: true,
			Headers: map[string]string{
				"X-Custom": "Header",
			},
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.WebhookOptions]{{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:8081",
				Headers: map[string]string{
					"X-Custom-2": "Header",
				},
			},
		}},
	},
	S3: &target.Config[target.S3Options]{
		Config: &target.S3Options{
			AWSConfig: target.AWSConfig{
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
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.S3Options]{{}},
	},
	Kinesis: &target.Config[target.KinesisOptions]{
		Config: &target.KinesisOptions{
			AWSConfig: target.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			StreamName: "policy-reporter",
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.KinesisOptions]{{}},
	},
	SecurityHub: &target.Config[target.SecurityHubOptions]{
		Config: &target.SecurityHubOptions{
			AWSConfig: target.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			AccountID: "AccountID",
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.SecurityHubOptions]{{}},
	},
	GCS: &target.Config[target.GCSOptions]{
		Config: &target.GCSOptions{
			Credentials: `{"token": "token", "type": "authorized_user"}`,
			Bucket:      "test",
			Prefix:      "prefix",
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.GCSOptions]{{}},
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
	SourceConfig: map[string]config.SourceConfig{
		"test": {
			CustomID: config.CustomID{
				Enabled: true,
				Fields:  []string{"resource"},
			},
		},
		"default": {},
	},
}

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	if count := len(resolver.TargetClients().Clients()); count != 25 {
		t.Errorf("Expected 25 Clients, got %d", count)
	}
}

func Test_ResolveHasTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	if !resolver.HasTargets() {
		t.Errorf("Expected 'true'")
	}
}

func Test_ResolveSkipExistingOnStartup(t *testing.T) {
	testConfig := &config.Config{
		Targets: target.Targets{
			Loki: &target.Config[target.LokiOptions]{
				Config: &target.LokiOptions{
					HostOptions: target.HostOptions{
						Host: "http://localhost:3100",
					},
				},
				SkipExisting:    true,
				MinimumPriority: "debug",
			},
			Elasticsearch: &target.Config[target.ElasticsearchOptions]{
				Config: &target.ElasticsearchOptions{
					HostOptions: target.HostOptions{
						Host: "http://localhost:9200",
					},
				},
				SkipExisting:    true,
				MinimumPriority: "debug",
			},
		},
	}

	t.Run("Resolve false", func(t *testing.T) {
		testConfig.Targets.Elasticsearch.SkipExisting = false

		resolver := config.NewResolver(testConfig, &rest.Config{})

		if resolver.SkipExistingOnStartup() == true {
			t.Error("Expected SkipExistingOnStartup to be false if one Client has SkipExistingOnStartup false configured")
		}
	})

	t.Run("Resolve true", func(t *testing.T) {
		testConfig.Targets.Elasticsearch.SkipExisting = true

		resolver := config.NewResolver(testConfig, &rest.Config{})

		if resolver.SkipExistingOnStartup() == false {
			t.Error("Expected SkipExistingOnStartup to be true if all Client has SkipExistingOnStartup true configured")
		}
	})
}

func Test_ResolvePolicyClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})

	client1, err := resolver.PolicyReportClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, _ := resolver.PolicyReportClient()
	if client1 != client2 {
		t.Error("A second call resolver.PolicyReportClient() should return the cached first client")
	}
}

func Test_ResolveLeaderElectionClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})

	client1, err := resolver.LeaderElectionClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, _ := resolver.LeaderElectionClient()
	if client1 != client2 {
		t.Error("A second call resolver.LeaderElectionClient() should return the cached first client")
	}
}

func Test_ResolvePolicyStore(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})
	db := resolver.Database()
	defer db.Close()

	store1, err := resolver.Store(db)
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	store2, _ := resolver.Store(db)
	if store1 != store2 {
		t.Error("A second call resolver.PolicyReportClient() should return the cached first client")
	}
}

func Test_ResolveAPIServer(t *testing.T) {
	resolver := config.NewResolver(&config.Config{
		API: config.API{
			BasicAuth: config.BasicAuth{Username: "user", Password: "password"},
		},
	}, &rest.Config{})

	server, _ := resolver.Server(context.Background(), nil)
	if server == nil {
		t.Error("Error: Should return API Server")
	}
}

func Test_ResolveCache(t *testing.T) {
	t.Run("InMemory", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})

		cache1 := resolver.ResultCache()
		if cache1 == nil {
			t.Error("Error: Should return ResultCache")
		}

		cache2 := resolver.ResultCache()
		if cache1 != cache2 {
			t.Error("A second call resolver.ResultCache() should return the cached first cache")
		}
	})

	t.Run("Redis", func(t *testing.T) {
		redisConfig := &config.Config{
			Redis: config.Redis{
				Enabled: true,
				Address: "localhost:6379",
			},
		}

		resolver := config.NewResolver(redisConfig, &rest.Config{})

		cache1 := resolver.ResultCache()
		if cache1 == nil {
			t.Error("Error: Should return ResultCache")
		}
	})
}

func Test_ResolveReportFilter(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	filter := resolver.ReportFilter()
	if filter == nil {
		t.Error("Error: Should return Filter")
	}
}

func Test_ResolveClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.PolicyReportClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolveLeaderElectionWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.LeaderElectionClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolveCRDClient(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	_, err := resolver.CRDClient()
	if err != nil {
		t.Error("unexpected error")
	}
}

func Test_ResolveCRDClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.CRDClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolveSecretClient(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	client := resolver.SecretClient()
	if client == nil {
		t.Error("unexpected error")
	}
}

func Test_ResolveSecretCClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	client := resolver.SecretClient()
	if client != nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_RegisterStoreListener(t *testing.T) {
	t.Run("Register StoreListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterStoreListener(context.Background(), report.NewPolicyReportStore())

		if len(resolver.EventPublisher().GetListener()) != 1 {
			t.Error("Expected one Listener to be registered")
		}
	})
}

func Test_RegisterMetricsListener(t *testing.T) {
	t.Run("Register MetricsListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterMetricsListener()

		if len(resolver.EventPublisher().GetListener()) != 1 {
			t.Error("Expected one Listener to be registered")
		}
	})
}

func Test_RegisterSendResultListener(t *testing.T) {
	t.Run("Register SendResultListener with Targets", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterSendResultListener()

		if len(resolver.EventPublisher().GetListener()) != 1 {
			t.Error("Expected one Listener to be registered")
		}
	})
	t.Run("Register SendResultListener without Targets", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{}, &rest.Config{})

		resolver.RegisterSendResultListener()

		if len(resolver.EventPublisher().GetListener()) != 0 {
			t.Error("Expected no Listener to be registered because no target exists")
		}
	})
}

func Test_SummaryReportServices(t *testing.T) {
	t.Run("Generator", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		generator, err := resolver.SummaryGenerator()
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if generator == nil {
			t.Error("Should return Generator Pointer")
		}
	})
	t.Run("Generator.Error", func(t *testing.T) {
		k8sConfig := &rest.Config{}
		k8sConfig.Host = "invalid/url"

		resolver := config.NewResolver(testConfig, k8sConfig)

		_, err := resolver.SummaryGenerator()
		if err == nil {
			t.Error("Error: 'host must be a URL or a host:port pair' was expected")
		}
	})
	t.Run("Reporter", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		reporter := resolver.SummaryReporter()
		if reporter == nil {
			t.Error("Should return Reporter Pointer")
		}
	})
}

func Test_ViolationReportServices(t *testing.T) {
	t.Run("Generator", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		generator, err := resolver.ViolationsGenerator()
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if generator == nil {
			t.Error("Should return Generator Pointer")
		}
	})
	t.Run("Generator.Error", func(t *testing.T) {
		k8sConfig := &rest.Config{}
		k8sConfig.Host = "invalid/url"

		resolver := config.NewResolver(testConfig, k8sConfig)

		_, err := resolver.ViolationsGenerator()
		if err == nil {
			t.Error("Error: 'host must be a URL or a host:port pair' was expected")
		}
	})
	t.Run("Reporter", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		reporter := resolver.ViolationsReporter()
		if reporter == nil {
			t.Error("Should return Reporter Pointer")
		}
	})
}

func Test_SMTP(t *testing.T) {
	t.Run("SMTP", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		smtp := resolver.SMTPServer()
		if smtp == nil {
			t.Error("Should return SMTP Pointer")
		}
	})
	t.Run("EmailClient", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		client := resolver.EmailClient()
		if client == nil {
			t.Error("Should return EmailClient Pointer")
		}
	})
}

func Test_ResolveLogger(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	logger1, _ := resolver.Logger()
	if logger1 == nil {
		t.Error("Error: Should return Logger")
	}

	logger2, _ := resolver.Logger()
	if logger1 != logger2 {
		t.Error("A second call resolver.Mapper() should return the cached first cache")
	}
}

func Test_Logger(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	logger, _ := resolver.Logger()

	assert.NotNil(t, logger)
}

func Test_ResolveEnableLeaderElection(t *testing.T) {
	t.Run("general disabled", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: false},
			Targets: target.Targets{
				Loki: &target.Config[target.LokiOptions]{
					Config: &target.LokiOptions{
						HostOptions: target.HostOptions{
							Host: "http://localhost:3100",
						},
					},
				},
			},
			Database: config.Database{Type: database.MySQL},
		}, &rest.Config{})

		if resolver.EnableLeaderElection() {
			t.Error("leaderelection should be not enabled if its general disabled")
		}
	})

	t.Run("no pushes and SQLite Database", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: true},
			Database:       config.Database{Type: database.SQLite},
			DBFile:         "test.db",
		}, &rest.Config{})

		if resolver.EnableLeaderElection() {
			t.Error("leaderelection should be not enabled if no pushes configured and SQLite is used")
		}
	})

	t.Run("enabled if pushes defined", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: true},
			Database:       config.Database{Type: database.SQLite},
			Targets: target.Targets{
				Loki: &target.Config[target.LokiOptions]{
					Config: &target.LokiOptions{
						HostOptions: target.HostOptions{
							Host: "http://localhost:3100",
						},
					},
				},
			},
			DBFile: "test.db",
		}, &rest.Config{})

		if !resolver.EnableLeaderElection() {
			t.Error("leaderelection should be enabled if general enabled and targets configured")
		}
	})
}

func Test_ResolveCustomIDGenerators(t *testing.T) {
	resolver := config.NewResolver(testConfig, nil)

	generators := resolver.CustomIDGenerators()
	assert.Equal(t, 1, len(generators), "only enabled custom id config should be mapped")
}
