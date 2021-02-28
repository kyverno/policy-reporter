package config_test

import (
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/config"
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
		Webhook:         "http://localhost:80",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
}

func Test_ResolveLokiClient(t *testing.T) {
	resolver := config.NewResolver(testConfig)

	client := resolver.LokiClient()
	if client == nil {
		t.Error("Expected Client, got nil")
	}

	client2 := resolver.LokiClient()
	if client != client2 {
		t.Error("Error: Should reuse first instance")
	}
}

func Test_ResolveElasticSearchClient(t *testing.T) {
	resolver := config.NewResolver(testConfig)

	client := resolver.ElasticsearchClient()
	if client == nil {
		t.Error("Expected Client, got nil")
	}

	client2 := resolver.ElasticsearchClient()
	if client != client2 {
		t.Error("Error: Should reuse first instance")
	}
}

func Test_ResolveSlackClient(t *testing.T) {
	resolver := config.NewResolver(testConfig)

	client := resolver.SlackClient()
	if client == nil {
		t.Error("Expected Client, got nil")
	}

	client2 := resolver.SlackClient()
	if client != client2 {
		t.Error("Error: Should reuse first instance")
	}
}

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig)

	clients := resolver.TargetClients()
	if count := len(clients); count != 3 {
		t.Errorf("Expected 3 Clients, got %d", count)
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
			Index:           "policy-reporter",
			Rotation:        "dayli",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
	}

	t.Run("Resolve false", func(t *testing.T) {
		testConfig.Elasticsearch.SkipExisting = false

		resolver := config.NewResolver(testConfig)

		if resolver.SkipExistingOnStartup() == true {
			t.Error("Expected SkipExistingOnStartup to be false if one Client has SkipExistingOnStartup false configured")
		}
	})

	t.Run("Resolve true", func(t *testing.T) {
		testConfig.Elasticsearch.SkipExisting = true

		resolver := config.NewResolver(testConfig)

		if resolver.SkipExistingOnStartup() == false {
			t.Error("Expected SkipExistingOnStartup to be true if all Client has SkipExistingOnStartup true configured")
		}
	})
}

func Test_ResolveLokiClientWithoutHost(t *testing.T) {
	config2 := &config.Config{
		Loki: config.Loki{
			Host:            "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
	}

	resolver := config.NewResolver(config2)
	resolver.Reset()

	if resolver.LokiClient() != nil {
		t.Error("Expected Client to be nil if no host is configured")
	}
}

func Test_ResolveElasticsearchClientWithoutHost(t *testing.T) {
	config2 := &config.Config{
		Elasticsearch: config.Elasticsearch{
			Host:            "",
			Index:           "policy-reporter",
			Rotation:        "dayli",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
	}

	resolver := config.NewResolver(config2)

	if resolver.ElasticsearchClient() != nil {
		t.Error("Expected Client to be nil if no host is configured")
	}
}

func Test_ResolveSlackClientWithoutHost(t *testing.T) {
	config2 := &config.Config{
		Slack: config.Slack{
			Webhook:         "",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
	}

	resolver := config.NewResolver(config2)

	if resolver.SlackClient() != nil {
		t.Error("Expected Client to be nil if no host is configured")
	}
}
