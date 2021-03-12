package api

import (
	"fmt"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/report"
)

type Server interface {
	Start() error
}

type httpServer struct {
	port   int
	mux    *http.ServeMux
	pStore *report.PolicyReportStore
	cStore *report.ClusterPolicyReportStore
}

func (s *httpServer) registerHandler() {
	s.mux.HandleFunc("/policy-reports", policyReportHandler(s.pStore))
	s.mux.HandleFunc("/cluster-policy-reports", clusterPolicyReportHandler(s.cStore))
}

func (s *httpServer) Start() error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}

	return server.ListenAndServe()
}

func NewServer(pStore *report.PolicyReportStore, cStore *report.ClusterPolicyReportStore, port int) Server {
	s := &httpServer{
		port:   port,
		cStore: cStore,
		pStore: pStore,
		mux:    http.NewServeMux(),
	}

	s.registerHandler()

	return s
}
