package config_test

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/kyverno/policy-reporter/pkg/config"
)

func createCMD() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Flags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.Flags().StringP("config", "c", "", "target configuration file")
	cmd.Flags().IntP("port", "p", 8080, "http port for the optional rest api")
	cmd.Flags().StringP("dbfile", "d", "sqlite-database.db", "path to the SQLite DB File")
	cmd.Flags().BoolP("metrics-enabled", "m", false, "Enable Policy Reporter's Metrics API")
	cmd.Flags().BoolP("rest-enabled", "r", false, "Enable Policy Reporter's REST API")
	cmd.Flags().Bool("profile", false, "Enable application profiling with pprof")
	cmd.Flags().StringP("template-dir", "t", "./templates", "template directory for email reports")

	return cmd
}

func Test_Load(t *testing.T) {
	cmd := createCMD()

	_ = cmd.Flags().Set("kubeconfig", "./config")
	_ = cmd.Flags().Set("port", "8081")
	_ = cmd.Flags().Set("rest-enabled", "1")
	_ = cmd.Flags().Set("metrics-enabled", "1")
	_ = cmd.Flags().Set("profile", "1")
	_ = cmd.Flags().Set("template-dir", "/app/templates")
	_ = cmd.Flags().Set("dbfile", "")

	c, err := config.Load(cmd)
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	if c.Kubeconfig != "./config" {
		t.Errorf("Unexpected TemplateDir Config: %s", c.Kubeconfig)
	}
	if c.API.Port != 8081 {
		t.Errorf("Unexpected Port Config: %d", c.API.Port)
	}
	if c.REST.Enabled != true {
		t.Errorf("Unexpected REST Config: %v", c.REST.Enabled)
	}
	if c.Metrics.Enabled != true {
		t.Errorf("Unexpected Metrics Config: %v", c.Metrics.Enabled)
	}
	if c.Profiling.Enabled != true {
		t.Errorf("Unexpected Profiling Config: %v", c.Profiling.Enabled)
	}
	if c.EmailReports.Templates.Dir != "/app/templates" {
		t.Errorf("Unexpected TemplateDir Config: %s", c.EmailReports.Templates.Dir)
	}
	if c.DBFile != "sqlite-database.db" {
		t.Errorf("Unexpected DBFile Config: %s", c.DBFile)
	}
}
