package cmd

import (
	"log"

	"github.com/fjogeleit/policy-reporter/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCLI creates a new instance of the root CLI
func NewCLI() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "policyreporter",
		Short: "Generates PolicyReport Metrics and Send Results to different targets",
		Long: `Generates Prometheus Metrics from PolicyReports, ClusterPolicyReports and PolicyReportResults.
		Sends notifications to different targets like Grafana's Loki.`,
	}

	rootCmd.AddCommand(newRunCMD())
	rootCmd.AddCommand(newSendCMD())

	return rootCmd
}

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	v := viper.New()

	v.SetDefault("namespace", "policy-reporter")
	v.SetDefault("api.port", 8080)

	cfgFile := ""

	configFlag := cmd.Flags().Lookup("config")
	if configFlag != nil {
		cfgFile = configFlag.Value.String()
	}

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName("config")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Println("[INFO] No target configuration file found")
	}

	if flag := cmd.Flags().Lookup("loki"); flag != nil {
		v.BindPFlag("loki.host", flag)
	}
	if flag := cmd.Flags().Lookup("loki-minimum-priority"); flag != nil {
		v.BindPFlag("loki.minimumPriority", flag)
	}
	if flag := cmd.Flags().Lookup("loki-skip-existing-on-startup"); flag != nil {
		v.BindPFlag("loki.skipExistingOnStartup", flag)
	}

	if flag := cmd.Flags().Lookup("kubeconfig"); flag != nil {
		v.BindPFlag("kubeconfig", flag)
	}

	if flag := cmd.Flags().Lookup("apiPort"); flag != nil {
		v.BindPFlag("api.port", flag)
		v.SetDefault("api.enabled", true)
	}

	c := &config.Config{}

	err := v.Unmarshal(c)

	return c, err
}
