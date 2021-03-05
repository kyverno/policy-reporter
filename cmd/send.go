package cmd

import (
	"context"
	"flag"
	"sync"

	"github.com/fjogeleit/policy-reporter/pkg/config"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newSendCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send all current PolicyReportResults to the configured targets",
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

			resolver := config.NewResolver(c, k8sConfig)

			client, err := resolver.PolicyResultClient(context.Background())
			if err != nil {
				return err
			}

			clients := resolver.TargetClients()

			if len(clients) == 0 {
				return nil
			}

			results, err := client.FetchPolicyResults()
			if err != nil {
				return err
			}

			wg := sync.WaitGroup{}
			wg.Add(len(results) * len(clients))

			for _, result := range results {
				for _, client := range clients {
					go func(c target.Client, r report.Result) {
						c.Send(r)
						wg.Done()
					}(client, result)
				}
			}

			wg.Wait()

			return err
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")

	cmd.PersistentFlags().String("loki", "", "loki host: http://loki:3100")
	cmd.PersistentFlags().String("loki-minimum-priority", "", "Minimum Priority to send Results to Loki (info < warning < error)")

	flag.Parse()

	return cmd
}
