package config_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/config"
)

func TestInitialReportsConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name         string
		rest, strict bool
		database     string
		valid        bool
	}{
		{"default", false, false, "", true},
		{"disabled external", true, false, "postgres", true},
		{"sqlite", true, true, "sqlite", true},
		{"default database", true, true, "", true},
		{"REST disabled", false, true, "sqlite", false},
		{"postgres", true, true, "postgres", false},
		{"mysql", true, true, "mysql", false},
		{"mariadb", true, true, "mariadb", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := config.Config{REST: config.REST{Enabled: tc.rest, WaitForInitialReports: tc.strict}, Database: config.Database{Type: tc.database}}
			if err := c.ValidateInitialReports(); (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
}
