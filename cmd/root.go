package cmd

import (
	"flag"
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
		Use:   "run",
		Short: "Kyverno Policy API",
		Long:  `Kyverno Policy API and Monitoring`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := LoadConfig(cmd)
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c)

			client, err := resolver.PolicyReportClient()
			if err != nil {
				return err
			}

			loki := resolver.LokiClient()

			if loki != nil {
				go client.WatchRuleValidation(func(r report.Result) {
					go loki.Send(r)
				}, c.Loki.SkipExisting)
			}

			policyMetrics, err := resolver.PolicyReportMetrics()
			if err != nil {
				return err
			}

			go policyMetrics.GenerateMetrics()

			clusterPolicyMetrics, err := resolver.ClusterPolicyReportMetrics()
			if err != nil {
				return err
			}

			go clusterPolicyMetrics.GenerateMetrics()

			http.Handle("/metrics", promhttp.Handler())
			http.ListenAndServe(":2112", nil)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")

	rootCmd.PersistentFlags().String("loki", "", "loki host: http://loki:3100")
	rootCmd.PersistentFlags().String("loki-minimum-priority", "", "Minimum Priority to send Results to Loki (info < warning < error)")
	rootCmd.PersistentFlags().Bool("loki-skip-exising-on-startup", false, "Skip Results created before PolicyReporter started. Prevent duplicated sending after new deployment")

	flag.Parse()

	return rootCmd
}

func LoadConfig(cmd *cobra.Command) (*config.Config, error) {
	v := viper.New()

	v.SetDefault("namespace", "policy-reporter")

	v.AutomaticEnv()

	if flag := cmd.Flags().Lookup("loki"); flag != nil {
		v.BindPFlag("loki.host", flag)
	}
	if flag := cmd.Flags().Lookup("loki-minimum-priority"); flag != nil {
		v.BindPFlag("loki.minimumPriority", flag)
	}
	if flag := cmd.Flags().Lookup("loki-skip-exising-on-startup"); flag != nil {
		v.BindPFlag("loki.skipExistingOnStartup", flag)
	}

	if flag := cmd.Flags().Lookup("kubeconfig"); flag != nil {
		v.BindPFlag("kubeconfig", flag)
	}

	c := &config.Config{}

	err := v.Unmarshal(c)

	return c, err
}
