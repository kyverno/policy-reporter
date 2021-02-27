package cmd

import (
	"flag"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/config"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run PolicyReporter Watcher & HTTP Metrics Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c)

			client, err := resolver.PolicyReportClient()
			if err != nil {
				return err
			}

			policyMetrics, err := resolver.PolicyReportMetrics()
			if err != nil {
				return err
			}

			clusterPolicyMetrics, err := resolver.ClusterPolicyReportMetrics()
			if err != nil {
				return err
			}

			g := new(errgroup.Group)

			g.Go(policyMetrics.GenerateMetrics)

			g.Go(clusterPolicyMetrics.GenerateMetrics)

			g.Go(func() error {
				targets := resolver.TargetClients()

				if len(targets) == 0 {
					return nil
				}

				return client.WatchPolicyReportResults(func(r report.Result, e bool) {
					for _, t := range targets {
						go func(target target.Client, result report.Result, preExisted bool) {
							if preExisted && target.SkipExistingOnStartup() {
								return
							}

							target.Send(result)
						}(t, r, e)
					}
				}, resolver.SkipExistingOnStartup())
			})

			g.Go(func() error {
				http.Handle("/metrics", promhttp.Handler())

				return http.ListenAndServe(":2112", nil)
			})

			return g.Wait()
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")

	cmd.PersistentFlags().String("loki", "", "loki host: http://loki:3100")
	cmd.PersistentFlags().String("loki-minimum-priority", "", "Minimum Priority to send Results to Loki (info < warning < error)")
	cmd.PersistentFlags().Bool("loki-skip-existing-on-startup", false, "Skip Results created before PolicyReporter started. Prevent duplicated sending after new deployment")

	flag.Parse()

	return cmd
}
