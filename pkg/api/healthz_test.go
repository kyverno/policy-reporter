package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckSuccess(t *testing.T) {
	check := func() error {
		return nil
	}

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithHealthChecks([]api.HealthCheck{check}))

	req, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
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

	req, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusServiceUnavailable, w.Code)
}

func TestReadyCheckSuccess(t *testing.T) {
	check := func() error {
		return nil
	}

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithHealthChecks([]api.HealthCheck{check}))

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
}
