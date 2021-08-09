package cmd

import (
	"context"
	"flag"
	"net/http"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
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

			client, err := resolver.PolicyReportClient(ctx)
			if err != nil {
				return err
			}

			client.RegisterCallback(metrics.CreateMetricsCallback())

			targets := resolver.TargetClients()

			if len(targets) > 0 {
				client.RegisterPolicyResultCallback(func(r report.Result, e bool) {
					for _, t := range targets {
						go func(target target.Client, result report.Result, preExisted bool) {
							if preExisted && target.SkipExistingOnStartup() {
								return
							}

							target.Send(result)
						}(t, r, e)
					}
				})

				client.RegisterPolicyResultWatcher(resolver.SkipExistingOnStartup())
			}

			errorChan := make(chan error)

			go func() { errorChan <- client.StartWatching() }()

			go func() { errorChan <- resolver.APIServer(ctx).Start() }()

			go func() {
				http.Handle("/metrics", promhttp.Handler())

				errorChan <- http.ListenAndServe(":2112", nil)
			}()

			return <-errorChan
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().IntP("apiPort", "a", 8080, "http port for the optional rest api")

	flag.Parse()

	return cmd
}
