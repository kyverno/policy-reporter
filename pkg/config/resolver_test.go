package config_test

import (
	"context"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var testConfig = &config.Config{
	Loki: &config.Loki{
		Host: "http://localhost:3100",
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		SkipTLS: true,
		Channels: []*config.Loki{
			{
				TargetBaseOptions: config.TargetBaseOptions{
					CustomFields: map[string]string{"label2": "value2"},
				},
			},
		},
	},
	Elasticsearch: &config.Elasticsearch{
		Host:     "http://localhost:9200",
		Index:    "policy-reporter",
		Rotation: "daily",
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		SkipTLS:  true,
		Channels: []*config.Elasticsearch{{}},
	},
	Slack: &config.Slack{
		Webhook: "http://hook.slack:80",
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		Channels: []*config.Slack{{
			Webhook: "http://localhost:9200",
		}, {
			Channel: "general",
		}},
	},
	Discord: &config.Discord{
		Webhook: "http://hook.discord:80",
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		Channels: []*config.Discord{{
			Webhook: "http://localhost:9200",
		}},
	},
	Teams: &config.Teams{
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		Webhook: "http://hook.teams:80",
		SkipTLS: true,
		Channels: []*config.Teams{{
			Webhook: "http://localhost:9200",
		}},
	},
	UI: &config.UI{
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Host: "http://localhost:8080",
	},
	Webhook: &config.Webhook{
		Host: "http://localhost:8080",
		Headers: map[string]string{
			"X-Custom": "Header",
		},
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		SkipTLS: true,
		Channels: []*config.Webhook{{
			Host: "http://localhost:8081",
			Headers: map[string]string{
				"X-Custom-2": "Header",
			},
		}},
	},
	S3: &config.S3{
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		AWSConfig: config.AWSConfig{
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
		Channels:             []*config.S3{{}},
	},
	Kinesis: &config.Kinesis{
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		AWSConfig: config.AWSConfig{
			AccessKeyID:     "AccessKey",
			SecretAccessKey: "SecretAccessKey",
			Endpoint:        "https://yds.serverless.yandexcloud.net",
			Region:          "ru-central1",
		},
		StreamName: "policy-reporter",
		Channels:   []*config.Kinesis{{}},
	},
	SecurityHub: &config.SecurityHub{
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		AWSConfig: config.AWSConfig{
			AccessKeyID:     "AccessKey",
			SecretAccessKey: "SecretAccessKey",
			Endpoint:        "https://yds.serverless.yandexcloud.net",
			Region:          "ru-central1",
		},
		AccountID: "AccountID",
		Channels:  []*config.SecurityHub{{}},
	},
	GCS: &config.GCS{
		TargetBaseOptions: config.TargetBaseOptions{
			SkipExisting:    true,
			MinimumPriority: "debug",
			CustomFields:    map[string]string{"field": "value"},
		},
		Credentials: `{"token": "token", "type": "authorized_user"}`,
		Bucket:      "test",
		Prefix:      "prefix",
		Channels:    []*config.GCS{{}},
	},
	EmailReports: config.EmailReports{
		Templates: config.EmailTemplates{
			Dir: "../../templates",
		},
		SMTP: config.SMTP{
			Host:       "localhost",
			Port:       465,
			Username:   "policy-reporter@kyverno.io",
			Password:   "password",
			From:       "policy-reporter@kyverno.io",
			Encryption: "ssl/tls",
		},
	},
	Telegram: &config.Telegram{
		Token:  "XXX",
		ChatID: "123456",
		Channels: []*config.Telegram{
			{
				ChatID: "1234567",
			},
		},
	},
	GoogleChat: &config.GoogleChat{
		Webhook:  "http://localhost:900/webhook",
		Channels: []*config.GoogleChat{{}},
	},
}

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	if count := len(resolver.TargetClients()); count != 26 {
		t.Errorf("Expected 26 Clients, got %d", count)
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
		Loki: &config.Loki{
			Host: "http://localhost:3100",
			TargetBaseOptions: config.TargetBaseOptions{
				SkipExisting:    true,
				MinimumPriority: "debug",
			},
		},
		Elasticsearch: &config.Elasticsearch{
			Host: "http://localhost:9200",
			TargetBaseOptions: config.TargetBaseOptions{
				SkipExisting:    true,
				MinimumPriority: "debug",
			},
		},
	}

	t.Run("Resolve false", func(t *testing.T) {
		testConfig.Elasticsearch.SkipExisting = false

		resolver := config.NewResolver(testConfig, &rest.Config{})

		if resolver.SkipExistingOnStartup() == true {
			t.Error("Expected SkipExistingOnStartup to be false if one Client has SkipExistingOnStartup false configured")
		}
	})

	t.Run("Resolve true", func(t *testing.T) {
		testConfig.Elasticsearch.SkipExisting = true

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

	store1, err := resolver.PolicyReportStore(db)
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	store2, _ := resolver.PolicyReportStore(db)
	if store1 != store2 {
		t.Error("A second call resolver.PolicyReportClient() should return the cached first client")
	}
}

func Test_ResolveAPIServer(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	server := resolver.APIServer(context.Background(), func() bool { return true })
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

func Test_ResolveMapper(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	mapper1 := resolver.Mapper()
	if mapper1 == nil {
		t.Error("Error: Should return Mapper")
	}

	mapper2 := resolver.Mapper()
	if mapper1 != mapper2 {
		t.Error("A second call resolver.Mapper() should return the cached first cache")
	}
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

func Test_ResolveEnableLeaderElection(t *testing.T) {
	t.Run("general disabled", func(t *testing.T) {
		resolver := config.NewResolver(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: false},
			Loki:           &config.Loki{Host: "localhost:3100"},
			Database:       config.Database{Type: database.MySQL},
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
			Loki:           &config.Loki{Host: "localhost:3100"},
			DBFile:         "test.db",
		}, &rest.Config{})

		if !resolver.EnableLeaderElection() {
			t.Error("leaderelection should be enabled if general enabled and targets configured")
		}
	})
}
