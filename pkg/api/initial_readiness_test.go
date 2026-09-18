package api_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kyverno/policy-reporter/pkg/api"
)

type initializationHandler struct{}

func (initializationHandler) Register(group *gin.RouterGroup) error {
	group.GET("data", func(c *gin.Context) { c.Status(200) })
	return nil
}

func TestInitialReadinessOnlyGatesReadyAndREST(t *testing.T) {
	ready := false
	check := func() error {
		if !ready {
			return errors.New("initial reports pending")
		}
		return nil
	}
	server := api.NewServer(gin.New(), api.WithHealthChecks(nil, check), api.WithRESTReadiness(check), api.WithMetrics(), api.WithProfiling())
	for _, version := range []string{"v1", "v2"} {
		if err := server.Register(version, initializationHandler{}); err != nil {
			t.Fatal(err)
		}
	}
	for _, loaded := range []bool{false, true} {
		ready = loaded
		for _, path := range []string{"/healthz", "/metrics", "/debug/pprof/", "/ready", "/v1/data", "/v2/data"} {
			expected := 200
			if !loaded && (path == "/ready" || path == "/v1/data" || path == "/v2/data") {
				expected = 503
			}
			w := httptest.NewRecorder()
			server.Serve(w, httptest.NewRequest("GET", path, nil))
			if w.Code != expected {
				t.Errorf("loaded=%v %s: got %d want %d", loaded, path, w.Code, expected)
			}
		}
	}
}
