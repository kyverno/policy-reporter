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
