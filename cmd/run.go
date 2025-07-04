package cmd

import (
	"context"
	"errors"
	"flag"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyverno/policy-reporter/pkg/api"
	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	v2 "github.com/kyverno/policy-reporter/pkg/api/v2"
	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func newRunCMD(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run PolicyReporter Watcher & HTTP Metrics Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Load(cmd)
			if err != nil {
				return err
			}
			c.Version = version

			var k8sConfig *rest.Config
			if c.K8sClient.Kubeconfig != "" {
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.K8sClient.Kubeconfig)
			} else {
				k8sConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				return err
			}

			k8sConfig.QPS = c.K8sClient.QPS
			k8sConfig.Burst = c.K8sClient.Burst

			readinessProbe := config.NewReadinessProbe(c)
			defer readinessProbe.Close()

			resolver := config.NewResolver(c, k8sConfig)
			logger, err := resolver.Logger()
			if err != nil {
				return err
			}

			var wgClient, orClient report.PolicyReportClient

			orClient, err = resolver.OpenReportsClient()
			if err != nil {
				return err
			}
			wgClient, err = resolver.WGPolicyReportClient()
			if err != nil {
				return err
			}

			if wgClient == nil && orClient == nil {
				return errors.New("no valid reporting API group found in the cluster")
			}

			secretInformer, err := resolver.SecretInformer()
			if err != nil {
				return err
			}

			g := &errgroup.Group{}

			var store *database.Store
			servOptions := []api.ServerOption{
				api.WithPort(c.API.Port),
				api.WithHealthChecks([]api.HealthCheck{
					func() error {
						if wgClient == nil {
							return nil
						}
						if !wgClient.HasSynced() {
							return errors.New("wginformer not ready")
						}
						return nil
					},
					func() error {
						if orClient == nil {
							return nil
						}
						if !orClient.HasSynced() {
							return errors.New("orinformer not ready")
						}
						return nil
					},
				}),
			}

			if c.REST.Enabled {
				db := resolver.Database()
				if db == nil {
					return errors.New("unable to create database connection")
				}
				defer db.Close()

				store, err = resolver.Store(db)
				if err != nil {
					return err
				}

				nsClient, err := resolver.NamespaceClient()
				if err != nil {
					return err
				}

				if !c.LeaderElection.Enabled || store.IsSQLite() {
					store.PrepareDatabase(cmd.Context())
					resolver.RegisterStoreListener(cmd.Context(), store)
				}

				logger.Info("REST api enabled")
				servOptions = append(servOptions, v1.WithAPI(store, resolver.TargetClients(), resolver.ViolationsReporter()), v2.WithAPI(store, nsClient, c.Targets))
			}

			if c.Metrics.Enabled {
				logger.Info("metrics enabled")
				resolver.RegisterMetricsListener()
				servOptions = append(servOptions, api.WithMetrics())
			}

			if c.Profiling.Enabled {
				logger.Info("pprof profiling enabled")
				servOptions = append(servOptions, api.WithProfiling())
			}

			if !resolver.ResultCache().Shared() {
				logger.Debug("register new result listener")
				resolver.RegisterNewResultsListener()
			}

			if resolver.EnableLeaderElection() {
				elector, err := resolver.LeaderElectionClient()
				if err != nil {
					return err
				}

				elector.RegisterOnStart(func(ctx context.Context) {
					logger.Info("started leadership")

					if c.REST.Enabled && !store.IsSQLite() {
						store.PrepareDatabase(cmd.Context())
						logger.Debug("register database persistence")
						resolver.RegisterStoreListener(ctx, store)

						if readinessProbe.Running() {
							logger.Debug("trigger informer restart")
							if orClient != nil {
								orClient.Stop()
							}

							if wgClient != nil {
								wgClient.Stop()
							}
						}
					}

					resolver.RegisterSendResultListener()

					readinessProbe.Ready()
				}).RegisterOnNew(func(currentID, lockID string) {
					if currentID != lockID {
						logger.Info("leadership", zap.String("leader", currentID))
						readinessProbe.Ready()
						return
					}
				}).RegisterOnStop(func() {
					logger.Info("stopped leadership")

					if !store.IsSQLite() {
						resolver.EventPublisher().UnregisterListener(listener.Store)
					}

					if resolver.HasTargets() {
						resolver.UnregisterSendResultListener()
					}
				})

				g.Go(func() error {
					return elector.Run(cmd.Context())
				})
			} else {
				resolver.RegisterSendResultListener()
				readinessProbe.Ready()
			}

			server, err := resolver.Server(cmd.Context(), servOptions)
			if err != nil {
				return err
			}

			g.Go(server.Start)

			if c.CRD.TargetConfig {
				g.Go(func() error {
					readinessProbe.Wait()

					stop := make(chan struct{})
					client, err := resolver.TargetConfigClient()
					if err != nil {
						return err
					}

					client.Run(stop)

					<-stop

					return nil
				})
			}
			if wgClient != nil {
				g.Go(func() error {
					logger.Info("wait for wgpolicy informer")
					readinessProbe.Wait()

					logger.Info("start wgpolicy client", zap.Int("worker", c.WorkerCount))
					for {
						stop := make(chan struct{})
						if err := wgClient.Run(c.WorkerCount, stop); err != nil {
							logger.Error("wgpolicy informer client error", zap.Error(err))
						}

						logger.Debug("wgpolicy informer restarts")
					}
				})
			}

			if orClient != nil {
				g.Go(func() error {
					logger.Info("wait for openreports informer")
					readinessProbe.Wait()

					logger.Info("start openreports client", zap.Int("worker", c.WorkerCount))
					for {
						stop := make(chan struct{})
						if err := orClient.Run(c.WorkerCount, stop); err != nil {
							logger.Error("openreports informer client error", zap.Error(err))
						}

						logger.Debug("openreports informer restarts")
					}
				})
			}

			g.Go(func() error {
				collection := resolver.TargetClients()
				if !c.CRD.TargetConfig && !collection.UsesSecrets() {
					return nil
				}

				logger.Info("wait for secret informer")
				readinessProbe.Wait()

				stop := make(chan struct{})
				if err := secretInformer.Sync(collection, stop); err != nil {
					logger.Error("secret informer error", zap.Error(err))

					return err
				}

				<-stop

				return nil
			})

			return g.Wait()
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().IntP("port", "p", 8001, "http port for the optional rest api")
	cmd.PersistentFlags().StringP("dbfile", "d", "sqlite-database-v2.db", "path to the SQLite DB File")
	cmd.PersistentFlags().BoolP("metrics-enabled", "m", false, "Enable Policy Reporter's Metrics API")
	cmd.PersistentFlags().BoolP("rest-enabled", "r", false, "Enable Policy Reporter's REST API")
	cmd.PersistentFlags().Bool("profile", false, "Enable application profiling with pprof")
	cmd.PersistentFlags().String("lease-name", "policy-reporter", "name of the LeaseLock")
	cmd.PersistentFlags().String("pod-name", "policy-reporter", "name of the pod, used for leaderelection")
	cmd.PersistentFlags().StringP("template-dir", "t", "./templates", "template directory")
	cmd.PersistentFlags().Int("worker", 5, "amount of queue worker")
	cmd.PersistentFlags().Float32("qps", 20, "K8s RESTClient QPS")
	cmd.PersistentFlags().Int("burst", 50, "K8s RESTClient burst")

	flag.Parse()

	return cmd
}
