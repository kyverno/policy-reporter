package cmd

import (
	"context"
	"flag"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/config"
	"github.com/fjogeleit/policy-reporter/pkg/metrics"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

			var k8sConfig *rest.Config
			if c.Kubeconfig != "" {
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
			} else {
				k8sConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				return err
			}

			ctx := context.Background()

			resolver := config.NewResolver(c, k8sConfig)

			pClient, err := resolver.PolicyReportClient(ctx)
			if err != nil {
				return err
			}
			cpClient, err := resolver.ClusterPolicyReportClient(ctx)
			if err != nil {
				return err
			}
			rClient, err := resolver.PolicyResultClient(ctx)
			if err != nil {
				return err
			}

			cpClient.RegisterCallback(metrics.CreateClusterPolicyReportMetricsCallback())
			pClient.RegisterCallback(metrics.CreatePolicyReportMetricsCallback())

			targets := resolver.TargetClients()

			if len(targets) > 0 {
				rClient.RegisterPolicyResultCallback(func(r report.Result, e bool) {
					for _, t := range targets {
						go func(target target.Client, result report.Result, preExisted bool) {
							if preExisted && target.SkipExistingOnStartup() {
								return
							}

							target.Send(result)
						}(t, r, e)
					}
				})

				rClient.RegisterPolicyResultWatcher(resolver.SkipExistingOnStartup())
			}

			g := new(errgroup.Group)

			if c.API.Enabled {
				g.Go(resolver.APIServer().Start)
			}

			g.Go(cpClient.StartWatching)
			g.Go(pClient.StartWatching)
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
	cmd.PersistentFlags().IntP("apiPort", "a", 0, "http port for the optional rest api")

	cmd.PersistentFlags().String("loki", "", "loki host: http://loki:3100")
	cmd.PersistentFlags().String("loki-minimum-priority", "", "Minimum Priority to send Results to Loki (info < warning < error)")
	cmd.PersistentFlags().Bool("loki-skip-existing-on-startup", false, "Skip Results created before PolicyReporter started. Prevent duplicated sending after new deployment")

	flag.Parse()

	return cmd
}
