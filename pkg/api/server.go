package api

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	pprof "net/http/pprof"
)

// Server for the Lifecycle and optional HTTP REST API
type Server interface {
	// Start the HTTP Server
	Start() error
	// Shutdown the HTTP Sever
	Shutdown(ctx context.Context) error
	// RegisterLifecycleHandler adds healthy and readiness APIs
	RegisterLifecycleHandler()
	// RegisterMetricsHandler adds the optional metrics endpoint
	RegisterMetricsHandler()
	// RegisterV1Handler adds the optional v1 REST APIs
	RegisterV1Handler(v1.PolicyReportFinder)
	// RegisterProfilingHandler adds the optional pprof profiling APIs
	RegisterProfilingHandler()
}

type httpServer struct {
	http    http.Server
	mux     *http.ServeMux
	targets []target.Client
	synced  func() bool
}

func (s *httpServer) RegisterLifecycleHandler() {
	s.mux.HandleFunc("/healthz", HealthzHandler(s.synced))
	s.mux.HandleFunc("/ready", ReadyHandler(s.synced))
}

func (s *httpServer) RegisterV1Handler(finder v1.PolicyReportFinder) {
	s.mux.HandleFunc("/v1/targets", Gzip(v1.TargetsHandler(s.targets)))
	s.mux.HandleFunc("/v1/categories", Gzip(v1.CategoryListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaces", Gzip(v1.NamespaceListHandler(finder)))
	s.mux.HandleFunc("/v1/rule-status-count", Gzip(v1.RuleStatusCountHandler(finder)))

	s.mux.HandleFunc("/v1/namespaced-resources/policies", Gzip(v1.NamespacedResourcesPolicyListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/rules", Gzip(v1.NamespacedResourcesRuleListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/kinds", Gzip(v1.NamespacedResourcesKindListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/resources", Gzip(v1.NamespacedResourcesListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/sources", Gzip(v1.NamespacedSourceListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/status-counts", Gzip(v1.NamespacedResourcesStatusCountsHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/results", Gzip(v1.NamespacedResourcesResultHandler(finder)))

	s.mux.HandleFunc("/v1/cluster-resources/policies", Gzip(v1.ClusterResourcesPolicyListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/rules", Gzip(v1.ClusterResourcesRuleListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/kinds", Gzip(v1.ClusterResourcesKindListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/resources", Gzip(v1.ClusterResourcesListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/sources", Gzip(v1.ClusterResourcesSourceListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/status-counts", Gzip(v1.ClusterResourcesStatusCountHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/results", Gzip(v1.ClusterResourcesResultHandler(finder)))
}

func (s *httpServer) RegisterMetricsHandler() {
	s.mux.Handle("/metrics", promhttp.Handler())
}

func (s *httpServer) RegisterProfilingHandler() {
	s.mux.HandleFunc("/debug/pprof/", pprof.Index)
	s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (s *httpServer) Start() error {
	return s.http.ListenAndServe()
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// NewServer constructor for a new API Server
func NewServer(targets []target.Client, port int, synced func() bool) Server {
	mux := http.NewServeMux()

	s := &httpServer{
		targets: targets,
		synced:  synced,
		mux:     mux,
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	s.RegisterLifecycleHandler()

	return s
}
