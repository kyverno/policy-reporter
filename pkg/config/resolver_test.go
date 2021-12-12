package config_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/report"
	"k8s.io/client-go/rest"
)

var testConfig = &config.Config{
	Loki: config.Loki{
		Host:            "http://localhost:3100",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	Elasticsearch: config.Elasticsearch{
		Host:            "http://localhost:9200",
		Index:           "policy-reporter",
		Rotation:        "dayli",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	Slack: config.Slack{
		Webhook:         "http://hook.slack:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	Discord: config.Discord{
		Webhook:         "http://hook.discord:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	Teams: config.Teams{
		Webhook:         "http://hook.teams:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	UI: config.UI{
		Host:            "http://localhost:8080",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	S3: config.S3{
		AccessKeyID:     "AccessKey",
		SecretAccessKey: "SecretAccessKey",
		Bucket:          "test",
		SkipExisting:    true,
		MinimumPriority: "debug",
		Endpoint:        "https://storage.yandexcloud.net",
		Region:          "ru-central1",
	},
}

func Test_ResolveTarget(t *testing.T) {
	resolver := config.NewResolver(testConfig, nil)

	t.Run("Loki", func(t *testing.T) {
		client := resolver.LokiClient()
		if client == nil {
			t.Error("Expected Client, got nil")
		}

		client2 := resolver.LokiClient()
		if client != client2 {
			t.Error("Error: Should reuse first instance")
		}
	})
	t.Run("Elasticsearch", func(t *testing.T) {
		client := resolver.ElasticsearchClient()
		if client == nil {
			t.Error("Expected Client, got nil")
		}

		client2 := resolver.ElasticsearchClient()
		if client != client2 {
			t.Error("Error: Should reuse first instance")
		}
	})
	t.Run("Slack", func(t *testing.T) {
		client := resolver.SlackClient()
		if client == nil {
			t.Error("Expected Client, got nil")
		}

		client2 := resolver.SlackClient()
		if client != client2 {
			t.Error("Error: Should reuse first instance")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		client := resolver.DiscordClient()
		if client == nil {
			t.Error("Expected Client, got nil")
		}

		client2 := resolver.DiscordClient()
		if client != client2 {
			t.Error("Error: Should reuse first instance")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		client := resolver.TeamsClient()
		if client == nil {
			t.Error("Expected Client, got nil")
		}

		client2 := resolver.TeamsClient()
		if client != client2 {
			t.Error("Error: Should reuse first instance")
		}
	})
	t.Run("S3", func(t *testing.T) {
		client := resolver.S3Client()
		if client == nil {
			t.Error("Expected Client, got nil")
		}

		client2 := resolver.S3Client()
		if client != client2 {
			t.Error("Error: Should reuse first instance")
		}
	})
}

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, nil)

	clients := resolver.TargetClients()
	if count := len(clients); count != 7 {
		t.Errorf("Expected 7 Clients, got %d", count)
	}
}

func Test_ResolveSkipExistingOnStartup(t *testing.T) {
	var testConfig = &config.Config{
		Loki: config.Loki{
			Host:            "http://localhost:3100",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Elasticsearch: config.Elasticsearch{
			Host:            "http://localhost:9200",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
	}

	t.Run("Resolve false", func(t *testing.T) {
		testConfig.Elasticsearch.SkipExisting = false

		resolver := config.NewResolver(testConfig, nil)

		if resolver.SkipExistingOnStartup() == true {
			t.Error("Expected SkipExistingOnStartup to be false if one Client has SkipExistingOnStartup false configured")
		}
	})

	t.Run("Resolve true", func(t *testing.T) {
		testConfig.Elasticsearch.SkipExisting = true

		resolver := config.NewResolver(testConfig, nil)

		if resolver.SkipExistingOnStartup() == false {
			t.Error("Expected SkipExistingOnStartup to be true if all Client has SkipExistingOnStartup true configured")
		}
	})
}

func Test_ResolveTargetWithoutHost(t *testing.T) {
	config2 := &config.Config{
		Loki: config.Loki{
			Host:            "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Elasticsearch: config.Elasticsearch{
			Host:            "",
			Index:           "policy-reporter",
			Rotation:        "dayli",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Slack: config.Slack{
			Webhook:         "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Discord: config.Discord{
			Webhook:         "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Teams: config.Teams{
			Webhook:         "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		S3: config.S3{
			Endpoint:        "",
			Region:          "",
			AccessKeyID:     "",
			SecretAccessKey: "",
			Bucket:          "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
	}

	t.Run("Loki", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if resolver.LokiClient() != nil {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Elasticsearch", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if resolver.ElasticsearchClient() != nil {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Slack", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if resolver.SlackClient() != nil {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if resolver.DiscordClient() != nil {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if resolver.TeamsClient() != nil {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("S3.Endoint", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if resolver.S3Client() != nil {
			t.Error("Expected Client to be nil if no endpoint is configured")
		}
	})
	t.Run("S3.AccessKey", func(t *testing.T) {
		config2.S3.Endpoint = "https://storage.yandexcloud.net"

		resolver := config.NewResolver(config2, nil)

		if resolver.S3Client() != nil {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("S3.AccessKey", func(t *testing.T) {
		config2.S3.Endpoint = "https://storage.yandexcloud.net"

		resolver := config.NewResolver(config2, nil)

		if resolver.S3Client() != nil {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		config2.S3.AccessKeyID = "access"

		resolver := config.NewResolver(config2, nil)

		if resolver.S3Client() != nil {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("S3.Region", func(t *testing.T) {
		config2.S3.SecretAccessKey = "secret"

		resolver := config.NewResolver(config2, nil)

		if resolver.S3Client() != nil {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("S3.Bucket", func(t *testing.T) {
		config2.S3.Region = "ru-central1"

		resolver := config.NewResolver(config2, nil)

		if resolver.S3Client() != nil {
			t.Error("Expected Client to be nil if no bucket is configured")
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

func Test_ResolvePolicyStore(t *testing.T) {
	resolver := config.NewResolver(&config.Config{DBFile: "test.db"}, &rest.Config{})
	db, _ := resolver.Database()
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

	server := resolver.APIServer(make(map[string]string))
	if server == nil {
		t.Error("Error: Should return API Server")
	}
}

func Test_ResolveCache(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	cache1 := resolver.ResultCache()
	if cache1 == nil {
		t.Error("Error: Should return ResultCache")
	}

	cache2 := resolver.ResultCache()
	if cache1 != cache2 {
		t.Error("A second call resolver.ResultCache() should return the cached first cache")
	}
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

func Test_ResolveClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(testConfig, k8sConfig)

	_, err := resolver.PolicyReportClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_RegisterStoreListener(t *testing.T) {
	t.Run("Register StoreListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterStoreListener(report.NewPolicyReportStore())

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
