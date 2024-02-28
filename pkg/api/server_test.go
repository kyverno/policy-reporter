package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/stretchr/testify/assert"
)

var check = func() error {
	return nil
}

func TestWithoutGZIP(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	server := api.NewServer(engine, api.WithHealthChecks([]api.HealthCheck{check}))

	req, _ := http.NewRequest("GET", "/healthz", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("", w.Header().Get("Content-Encoding"))
}

func TestWithGZIP(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithGZIP(), api.WithHealthChecks([]api.HealthCheck{check}))

	req, _ := http.NewRequest("GET", "/healthz", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("gzip", w.Header().Get("Content-Encoding"))
}

func TestWithProfiling(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithProfiling())

	req, _ := http.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
}
