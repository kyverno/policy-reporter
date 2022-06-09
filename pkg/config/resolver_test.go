package config_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/redis"
	"github.com/kyverno/policy-reporter/pkg/report"
	"k8s.io/client-go/rest"
)

var testConfig = &config.Config{
	Loki: config.Loki{
		Host:            "http://localhost:3100",
		SkipExisting:    true,
		MinimumPriority: "debug",
		CustomLabels:    map[string]string{"label": "value"},
		Channels: []config.Loki{
			{
				CustomLabels: map[string]string{"label2": "value2"},
			},
		},
	},
	Elasticsearch: config.Elasticsearch{
		Host:            "http://localhost:9200",
		Index:           "policy-reporter",
		Rotation:        "dayli",
		SkipExisting:    true,
		MinimumPriority: "debug",
		Channels:        []config.Elasticsearch{{}},
	},
	Slack: config.Slack{
		Webhook:         "http://hook.slack:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
		Channels: []config.Slack{{
			Webhook: "http://localhost:9200",
		}},
	},
	Discord: config.Discord{
		Webhook:         "http://hook.discord:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
		Channels: []config.Discord{{
			Webhook: "http://localhost:9200",
		}},
	},
	Teams: config.Teams{
		Webhook:         "http://hook.teams:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
		Channels: []config.Teams{{
			Webhook: "http://localhost:9200",
		}},
	},
	UI: config.UI{
		Host:            "http://localhost:8080",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	Webhook: config.Webhook{
		Host: "http://localhost:8080",
		Headers: map[string]string{
			"X-Custom": "Header",
		},
		SkipExisting:    true,
		MinimumPriority: "debug",
		Channels: []config.Webhook{{
			Host: "http://localhost:8081",
			Headers: map[string]string{
				"X-Custom-2": "Header",
			},
		}},
	},
	S3: config.S3{
		AccessKeyID:     "AccessKey",
		SecretAccessKey: "SecretAccessKey",
		Bucket:          "test",
		SkipExisting:    true,
		MinimumPriority: "debug",
		Endpoint:        "https://storage.yandexcloud.net",
		Region:          "ru-central1",
		Prefix:          "prefix",
		Channels:        []config.S3{{}},
	},
}

func Test_ResolveTarget(t *testing.T) {
	resolver := config.NewResolver(testConfig, nil)

	t.Run("Loki", func(t *testing.T) {
		clients := resolver.LokiClients()
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}

	})
	t.Run("Elasticsearch", func(t *testing.T) {
		clients := resolver.ElasticsearchClients()
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("Slack", func(t *testing.T) {
		clients := resolver.SlackClients()
		if len(clients) != 2 {
			t.Error("Expected Client, got nil")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		clients := resolver.DiscordClients()
		if len(clients) != 2 {
			t.Error("Expected Client, got nil")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		clients := resolver.TeamsClients()
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("Webhook", func(t *testing.T) {
		clients := resolver.WebhookClients()
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("S3", func(t *testing.T) {
		clients := resolver.S3Clients()
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
}

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig, nil)

	clients := resolver.TargetClients()
	if count := len(clients); count != 15 {
		t.Errorf("Expected 15 Clients, got %d", count)
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
		Webhook: config.Webhook{
			Host:            "",
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

		if len(resolver.LokiClients()) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Elasticsearch", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if len(resolver.ElasticsearchClients()) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Slack", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if len(resolver.SlackClients()) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if len(resolver.DiscordClients()) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if len(resolver.TeamsClients()) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Webhook", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if len(resolver.WebhookClients()) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("S3.Endoint", func(t *testing.T) {
		resolver := config.NewResolver(config2, nil)

		if len(resolver.S3Clients()) != 0 {
			t.Error("Expected Client to be nil if no endpoint is configured")
		}
	})
	t.Run("S3.AccessKey", func(t *testing.T) {
		config2.S3.Endpoint = "https://storage.yandexcloud.net"

		resolver := config.NewResolver(config2, nil)

		if len(resolver.S3Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("S3.AccessKey", func(t *testing.T) {
		config2.S3.Endpoint = "https://storage.yandexcloud.net"

		resolver := config.NewResolver(config2, nil)

		if len(resolver.S3Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		config2.S3.AccessKeyID = "access"

		resolver := config.NewResolver(config2, nil)

		if len(resolver.S3Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("S3.Region", func(t *testing.T) {
		config2.S3.SecretAccessKey = "secret"

		resolver := config.NewResolver(config2, nil)

		if len(resolver.S3Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("S3.Bucket", func(t *testing.T) {
		config2.S3.Region = "ru-central1"

		resolver := config.NewResolver(config2, nil)

		if len(resolver.S3Clients()) != 0 {
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
	t.Run("InMemory", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		_, ok := resolver.ResultCache().(*cache.InMemoryCache)
		if !ok {
			t.Error("Expected Cache to be InMemory Cache")
		}

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
		var redisConfig = &config.Config{
			Redis: config.Redis{
				Enabled: true,
				Address: "localhost:6379",
			},
		}

		resolver := config.NewResolver(redisConfig, &rest.Config{})
		_, ok := resolver.ResultCache().(*redis.RedisCache)
		if !ok {
			t.Error("Expected Cache to be Redis Cache")
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
