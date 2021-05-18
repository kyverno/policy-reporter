package api

import (
	"fmt"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
)

// Server for the optional HTTP REST API
type Server interface {
	// Start the HTTP REST API
	Start() error
}

type httpServer struct {
	port    int
	mux     *http.ServeMux
	store   *report.PolicyReportStore
	targets []Target
}

func (s *httpServer) registerHandler() {
	s.mux.HandleFunc("/policy-reports", Gzip(PolicyReportHandler(s.store)))
	s.mux.HandleFunc("/cluster-policy-reports", Gzip(ClusterPolicyReportHandler(s.store)))
	s.mux.HandleFunc("/targets", Gzip(TargetsHandler(s.targets)))
}

func (s *httpServer) Start() error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}

	return server.ListenAndServe()
}

// NewServer constructor for a new API Server
func NewServer(store *report.PolicyReportStore, targets []target.Client, port int) Server {
	apiTargets := make([]Target, 0, len(targets))
	for _, t := range targets {
		apiTargets = append(apiTargets, mapTarget(t))
	}

	s := &httpServer{
		port:    port,
		targets: apiTargets,
		store:   store,
		mux:     http.NewServeMux(),
	}

	s.registerHandler()

	return s
}
