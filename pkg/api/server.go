package api

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	RegisterV1Handler(finder v1.PolicyReportFinder)
}

type httpServer struct {
	http           http.Server
	mux            *http.ServeMux
	targets        []target.Client
	foundResources map[string]string
}

func (s *httpServer) RegisterLifecycleHandler() {
	s.mux.HandleFunc("/healthz", HealthzHandler(s.foundResources))
	s.mux.HandleFunc("/ready", ReadyHandler())
}

func (s *httpServer) RegisterV1Handler(finder v1.PolicyReportFinder) {
	s.mux.HandleFunc("/v1/targets", Gzip(v1.TargetsHandler(s.targets)))
	s.mux.HandleFunc("/v1/categories", Gzip(v1.CategoryListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaces", Gzip(v1.NamespaceListHandler(finder)))
	s.mux.HandleFunc("/v1/rule-status-count", Gzip(v1.RuleStatusCountHandler(finder)))

	s.mux.HandleFunc("/v1/namespaced-resources/policies", Gzip(v1.NamespacedResourcesPolicyListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/kinds", Gzip(v1.NamespacedResourcesKindListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/sources", Gzip(v1.NamespacedSourceListHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/status-counts", Gzip(v1.NamespacedResourcesStatusCountsHandler(finder)))
	s.mux.HandleFunc("/v1/namespaced-resources/results", Gzip(v1.NamespacedResourcesResultHandler(finder)))

	s.mux.HandleFunc("/v1/cluster-resources/policies", Gzip(v1.ClusterResourcesPolicyListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/kinds", Gzip(v1.ClusterResourcesKindListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/sources", Gzip(v1.ClusterResourcesSourceListHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/status-counts", Gzip(v1.ClusterResourcesStatusCountHandler(finder)))
	s.mux.HandleFunc("/v1/cluster-resources/results", Gzip(v1.ClusterResourcesResultHandler(finder)))
}

func (s *httpServer) RegisterMetricsHandler() {
	s.mux.Handle("/metrics", promhttp.Handler())
}

func (s *httpServer) Start() error {
	return s.http.ListenAndServe()
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// NewServer constructor for a new API Server
func NewServer(targets []target.Client, port int, foundResources map[string]string) Server {
	s := &httpServer{
		targets:        targets,
		mux:            http.DefaultServeMux,
		foundResources: foundResources,
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: http.DefaultServeMux,
		},
	}

	s.RegisterLifecycleHandler()

	return s
}
