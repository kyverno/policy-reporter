package config

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Load(cmd *cobra.Command) (*Config, error) {
	v := viper.New()

	v.SetDefault("leaderElection.releaseOnCancel", true)
	v.SetDefault("leaderElection.leaseDuration", 15)
	v.SetDefault("leaderElection.renewDeadline", 10)
	v.SetDefault("leaderElection.retryPeriod", 2)

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
		log.Println("[INFO] No configuration file found")
	}

	if flag := cmd.Flags().Lookup("kubeconfig"); flag != nil {
		v.BindPFlag("kubeconfig", flag)
	}

	if flag := cmd.Flags().Lookup("port"); flag != nil {
		v.BindPFlag("api.port", flag)
	}

	if flag := cmd.Flags().Lookup("rest-enabled"); flag != nil {
		v.BindPFlag("rest.enabled", flag)
	}

	if flag := cmd.Flags().Lookup("metrics-enabled"); flag != nil {
		v.BindPFlag("metrics.enabled", flag)
	}

	if flag := cmd.Flags().Lookup("profile"); flag != nil {
		v.BindPFlag("profiling.enabled", flag)
	}

	if flag := cmd.Flags().Lookup("dbfile"); flag != nil {
		v.BindPFlag("dbfile", flag)
	}

	if flag := cmd.Flags().Lookup("template-dir"); flag != nil {
		v.BindPFlag("emailReports.templates.dir", flag)
	}

	if flag := cmd.Flags().Lookup("lease-name"); flag != nil {
		v.BindPFlag("leaderElection.lockName", flag)
	}

	if err := v.BindEnv("leaderElection.podName", "POD_NAME"); err != nil {
		log.Printf("[WARNING] failed to bind env POD_NAME")
	}
	if err := v.BindEnv("leaderElection.namespace", "POD_NAMESPACE"); err != nil {
		log.Printf("[WARNING] failed to bind env POD_NAMESPACE")
	}

	c := &Config{}

	err := v.Unmarshal(c)

	if c.DBFile == "" {
		c.DBFile = "sqlite-database.db"
	}

	return c, err
}
