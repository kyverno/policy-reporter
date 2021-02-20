package cmd

import (
	"flag"
	"log"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/config"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type PolicySeverity = string

const (
	Fail  PolicySeverity = "fail"
	Warn  PolicySeverity = "warn"
	Error PolicySeverity = "error"
	Pass  PolicySeverity = "pass"
	Skip  PolicySeverity = "skip"
)

func NewCLI() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "policyreporter",
		SilenceUsage: true,
		Short:        "Kyverno Policy API",
		Long:         `Kyverno Policy API and Monitoring`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := LoadConfig(cmd)
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c)

			client, err := resolver.KubernetesClient()
			if err != nil {
				return err
			}

			loki := resolver.LokiClient()

			if loki != nil {
				go client.WatchRuleValidation(func(r report.Result) {
					go loki.Send(r)
				})
			}

			metrics, err := resolver.Metrics()
			if err != nil {
				return err
			}

			go metrics.GenerateMetrics()

			http.Handle("/metrics", promhttp.Handler())
			http.ListenAndServe(":2112", nil)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().StringP("loki", "l", "", "loki host: http://loki:3100")

	flag.Parse()

	return rootCmd
}

func LoadConfig(cmd *cobra.Command) (*config.Config, error) {
	v := viper.New()
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
		log.Println("no config provided")
	}

	if flag := cmd.Flags().Lookup("loki"); flag != nil {
		v.BindPFlag("loki.host", flag)
	}
	if flag := cmd.Flags().Lookup("kubeconfig"); flag != nil {
		v.BindPFlag("kubeconfig", flag)
	}

	c := &config.Config{}

	err := v.Unmarshal(c)

	return c, err
}
