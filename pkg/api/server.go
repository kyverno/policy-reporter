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
	auth    *BasicAuth
}

func (s *httpServer) middleware(handler http.HandlerFunc) http.HandlerFunc {
	handler = Gzip(handler)

	if s.auth != nil {
		handler = HTTPBasic(s.auth, handler)
	}

	return handler
}

func (s *httpServer) RegisterLifecycleHandler() {
	s.mux.HandleFunc("/healthz", HealthzHandler(s.synced))
	s.mux.HandleFunc("/ready", ReadyHandler(s.synced))
}

func (s *httpServer) RegisterV1Handler(finder v1.PolicyReportFinder) {
	handler := v1.NewHandler(finder)

	s.mux.HandleFunc("/v1/targets", s.middleware(handler.TargetsHandler(s.targets)))
	s.mux.HandleFunc("/v1/namespaces", s.middleware(handler.NamespaceListHandler()))
	s.mux.HandleFunc("/v1/rule-status-count", s.middleware(handler.RuleStatusCountHandler()))

	s.mux.HandleFunc("/v1/policy-reports", s.middleware(handler.PolicyReportListHandler()))
	s.mux.HandleFunc("/v1/cluster-policy-reports", s.middleware(handler.ClusterPolicyReportListHandler()))

	s.mux.HandleFunc("/v1/namespaced-resources/categories", s.middleware(handler.NamespacedCategoryListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/policies", s.middleware(handler.NamespacedResourcesPolicyListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/rules", s.middleware(handler.NamespacedResourcesRuleListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/kinds", s.middleware(handler.NamespacedResourcesKindListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/resources", s.middleware(handler.NamespacedResourcesListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/sources", s.middleware(handler.NamespacedSourceListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/report-labels", s.middleware(handler.NamespacedReportLabelListHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/status-counts", s.middleware(handler.NamespacedResourcesStatusCountsHandler()))
	s.mux.HandleFunc("/v1/namespaced-resources/results", s.middleware(handler.NamespacedResourcesResultHandler()))

	s.mux.HandleFunc("/v1/cluster-resources/policies", s.middleware(handler.ClusterResourcesPolicyListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/rules", s.middleware(handler.ClusterResourcesRuleListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/kinds", s.middleware(handler.ClusterResourcesKindListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/resources", s.middleware(handler.ClusterResourcesListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/sources", Gzip(handler.ClusterResourcesSourceListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/report-labels", s.middleware(handler.ClusterReportLabelListHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/status-counts", s.middleware(handler.ClusterResourcesStatusCountHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/results", s.middleware(handler.ClusterResourcesResultHandler()))
	s.mux.HandleFunc("/v1/cluster-resources/categories", s.middleware(handler.ClusterCategoryListHandler()))
}

func (s *httpServer) RegisterMetricsHandler() {
	handler := promhttp.Handler()

	if s.auth != nil {
		s.mux.HandleFunc("/metrics", HTTPBasic(s.auth, handler.ServeHTTP))
		return
	}

	s.mux.Handle("/metrics", handler)
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
func NewServer(targets []target.Client, port int, logger *zap.Logger, auth *BasicAuth, synced func() bool) Server {
	mux := http.NewServeMux()

	s := &httpServer{
		targets: targets,
		synced:  synced,
		mux:     mux,
		logger:  logger,
		auth:    auth,
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
