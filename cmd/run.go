package cmd

import (
	"context"
	"errors"
	"flag"

	"golang.org/x/sync/errgroup"

	"github.com/kyverno/policy-reporter/pkg/config"
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

			client, err := resolver.PolicyReportClient()
			if err != nil {
				return err
			}

			resolver.RegisterSendResultListener()

			g := &errgroup.Group{}

			server := resolver.APIServer(client.GetFoundResources())

			if c.REST.Enabled {
				db, err := resolver.Database()
				if err != nil {
					return err
				}
				defer db.Close()

				store, err := resolver.PolicyReportStore(db)
				if err != nil {
					return err
				}

				resolver.RegisterStoreListener(store)
				server.RegisterV1Handler(store)
			}

			if c.Metrics.Enabled {
				resolver.RegisterMetricsListener()
				server.RegisterMetricsHandler()
			}

			g.Go(server.Start)

			g.Go(func() error {
				eventChan := client.WatchPolicyReports(ctx)

				resolver.EventPublisher().Publish(eventChan)

				return errors.New("event publisher stoped")
			})

			return g.Wait()
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().IntP("port", "p", 8080, "http port for the optional rest api")
	cmd.PersistentFlags().StringP("dbfile", "d", "sqlite-database.db", "path to the SQLite DB File")
	cmd.PersistentFlags().BoolP("metrics-enabled", "m", false, "Enable Policy Reporter's Metrics API")
	cmd.PersistentFlags().BoolP("rest-enabled", "r", false, "Enable Policy Reporter's REST API")

	flag.Parse()

	return cmd
}
