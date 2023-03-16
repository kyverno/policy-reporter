package api

import (
	"context"
	"fmt"
	"net/http"
	pprof "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/target"
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
	logger  *zap.Logger
}

func (s *httpServer) RegisterLifecycleHandler() {
	s.mux.HandleFunc("/healthz", HealthzHandler(s.synced, s.logger))
	s.mux.HandleFunc("/ready", ReadyHandler(s.synced, s.logger))
}

func (s *httpServer) RegisterV1Handler(finder v1.PolicyReportFinder) {
	handler := v1.NewHandler(finder, s.logger)

	s.mux.HandleFunc("/v1/targets", Gzip(handler.TargetsHandler(s.targets)))
	s.mux.HandleFunc("/v1/categories", Gzip(handler.CategoryListHandler()))
	s.mux.HandleFunc("/v1/namespaces", Gzip(handler.NamespaceListHandler()))
	s.mux.HandleFunc("/v1/rule-status-count", Gzip(handler.RuleStatusCountHandler()))

	s.mux.HandleFunc("/v1/policy-reports", Gzip(handler.PolicyReportListHandler()))
	s.mux.HandleFunc("/v1/cluster-policy-reports", Gzip(handler.ClusterPolicyReportListHandler()))

	s.mux.HandleFunc("/v1/namespaced-resources/policies", Gzip(handler.NamespacedResourcesPolicyListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/rules", Gzip(handler.NamespacedResourcesRuleListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/kinds", Gzip(handler.NamespacedResourcesKindListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/resources", Gzip(handler.NamespacedResourcesListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/sources", Gzip(handler.NamespacedSourceListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/report-labels", Gzip(handler.NamespacedReportLabelListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/status-counts", Gzip(handler.NamespacedResourcesStatusCountsHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/results", Gzip(handler.NamespacedResourcesResultHandler()))

	s.mux.HandleFunc("/v1/cluster-resources/policies", Gzip(handler.ClusterResourcesPolicyListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/rules", Gzip(handler.ClusterResourcesRuleListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/kinds", Gzip(handler.ClusterResourcesKindListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/resources", Gzip(handler.ClusterResourcesListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/sources", Gzip(handler.ClusterResourcesSourceListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/report-labels", Gzip(handler.ClusterReportLabelListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/status-counts", Gzip(handler.ClusterResourcesStatusCountHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/results", Gzip(handler.ClusterResourcesResultHandler()))
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
func NewServer(targets []target.Client, port int, logger *zap.Logger, synced func() bool) Server {
	mux := http.NewServeMux()

	s := &httpServer{
		targets: targets,
		synced:  synced,
		mux:     mux,
		logger:  logger,
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: NewLoggerMiddleware(logger, mux),
		},
	}

	s.RegisterLifecycleHandler()

	return s
}

func NewLoggerMiddleware(logger *zap.Logger, mux http.Handler) http.Handler {
	if logger == nil {
		return mux
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := []zapcore.Field{
			zap.String("proto", r.Proto),
			zap.String("user-agent", r.Header.Get("User-Agent")),
			zap.String("path", r.URL.Path),
		}

		if query := r.URL.RawQuery; query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if ref := r.Header.Get("Referer"); ref != "" {
			fields = append(fields, zap.String("referer", ref))
		}
		if scheme := r.URL.Scheme; scheme != "" {
			fields = append(fields, zap.String("scheme", scheme))
		}

		logger.Debug("Serve", fields...)

		mux.ServeHTTP(w, r)
	})
}
