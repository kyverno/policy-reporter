package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/api"
)

func TestHealthCheckSuccess(t *testing.T) {
	check := func() error {
		return nil
	}

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithHealthChecks([]api.HealthCheck{check}))

	req, _ := http.NewRequest("GET", "/healthz", http.NoBody)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheckError(t *testing.T) {
	check := func() error {
		return nil
	}

	err := func() error {
		return errors.New("unhealthy")
	}

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithHealthChecks([]api.HealthCheck{check, err}))

	req, _ := http.NewRequest("GET", "/healthz", http.NoBody)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestReadyCheckSuccess(t *testing.T) {
	check := func() error {
		return nil
	}

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithHealthChecks([]api.HealthCheck{check}))

	req, _ := http.NewRequest("GET", "/ready", http.NoBody)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
