package cmd

import (
	"context"
	"flag"
	"log"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/listener"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func newRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run PolicyReporter Watcher & HTTP Metrics Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Load(cmd)
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

			if resolver.HasTargets() && c.LeaderElection.Enabled {
				elector, err := resolver.LeaderElectionClient()
				if err != nil {
					return err
				}

				elector.RegisterOnStart(func(c context.Context) {
					klog.Info("started leadership")

					resolver.RegisterSendResultListener()
				}).RegisterOnNew(func(currentID, lockID string) {
					if currentID != lockID {
						klog.Infof("leadership by %s", currentID)
					}
				}).RegisterOnStop(func() {
					klog.Info("stopped leadership")

					resolver.EventPublisher().UnregisterListener(listener.NewResults)
				})

				g.Go(func() error {
					return elector.Run(cmd.Context())
				})
			} else if resolver.HasTargets() {
				resolver.RegisterSendResultListener()
			}

			g.Go(server.Start)

			g.Go(func() error {
				stop := make(chan struct{})
				defer close(stop)

				return client.Run(stop)
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
	cmd.PersistentFlags().String("lease-name", "policy-reporter", "name of the LeaseLock")

	flag.Parse()

	return cmd
}
