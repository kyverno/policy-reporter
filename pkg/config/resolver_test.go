package config_test

import (
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/config"
)

var testConfig = &config.Config{
	Loki: struct {
		Host            string "mapstructure:\"host\""
		SkipExisting    bool   "mapstructure:\"skipExistingOnStartup\""
		MinimumPriority string "mapstructure:\"minimumPriority\""
	}{
		Host:            "http://localhost:3100",
		SkipExisting:    true,
		MinimumPriority: "debug",
	},
	Elasticsearch: struct {
		Host            string "mapstructure:\"host\""
		Index           string "mapstructure:\"index\""
		Rotation        string "mapstructure:\"rotation\""
		SkipExisting    bool   "mapstructure:\"skipExistingOnStartup\""
		MinimumPriority string "mapstructure:\"minimumPriority\""
	}{
		Host:            "http://localhost:9200",
		Index:           "policy-reporter",
		Rotation:        "dayli",
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

func Test_ResolveTargets(t *testing.T) {
	resolver := config.NewResolver(testConfig)

	clients := resolver.TargetClients()
	if count := len(clients); count != 2 {
		t.Errorf("Expected 2 Clients, got %d", count)
	}
}

func Test_ResolveSkipExistingOnStartup(t *testing.T) {
	var testConfig = &config.Config{
		Loki: struct {
			Host            string "mapstructure:\"host\""
			SkipExisting    bool   "mapstructure:\"skipExistingOnStartup\""
			MinimumPriority string "mapstructure:\"minimumPriority\""
		}{
			Host:            "http://localhost:3100",
			SkipExisting:    true,
			MinimumPriority: "debug",
		},
		Elasticsearch: struct {
			Host            string "mapstructure:\"host\""
			Index           string "mapstructure:\"index\""
			Rotation        string "mapstructure:\"rotation\""
			SkipExisting    bool   "mapstructure:\"skipExistingOnStartup\""
			MinimumPriority string "mapstructure:\"minimumPriority\""
		}{
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
		Loki: struct {
			Host            string "mapstructure:\"host\""
			SkipExisting    bool   "mapstructure:\"skipExistingOnStartup\""
			MinimumPriority string "mapstructure:\"minimumPriority\""
		}{
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
		Elasticsearch: struct {
			Host            string "mapstructure:\"host\""
			Index           string "mapstructure:\"index\""
			Rotation        string "mapstructure:\"rotation\""
			SkipExisting    bool   "mapstructure:\"skipExistingOnStartup\""
			MinimumPriority string "mapstructure:\"minimumPriority\""
		}{
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
