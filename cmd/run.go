package cmd

import (
	"errors"
	"flag"
	"log"

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

			resolver := config.NewResolver(c, k8sConfig)

			client, err := resolver.PolicyReportClient()
			if err != nil {
				return err
			}

			server := resolver.APIServer(client.HasSynced)

			resolver.RegisterSendResultListener()

			g := &errgroup.Group{}

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

				log.Println("[INFO] REST api enabled")
				resolver.RegisterStoreListener(store)
				server.RegisterV1Handler(store)
			}

			if c.Metrics.Enabled {
				log.Println("[INFO] metrics enabled")
				resolver.RegisterMetricsListener()
				server.RegisterMetricsHandler()
			}

			if c.Profiling.Enabled {
				log.Println("[INFO] pprof profiling enabled")
				server.RegisterProfilingHandler()
			}

			g.Go(server.Start)

			stop := make(chan struct{})
			defer close(stop)

			eventChan, err := client.Run(stop)
			if err != nil {
				return err
			}

			g.Go(func() error {
				resolver.EventPublisher().Publish(eventChan)

				return errors.New("event publisher stopped")
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
	cmd.PersistentFlags().Bool("profile", false, "Enable application profiling with pprof")

	flag.Parse()

	return cmd
}
