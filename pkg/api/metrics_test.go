package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/api"
)

func TestMetrics(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithMetrics())

	req, _ := http.NewRequest("GET", "/metrics", http.NoBody)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricsWithBasicAuthError(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithBasicAuth(api.BasicAuth{
		Username: "user",
		Password: "password",
	}), api.WithMetrics())

	req, _ := http.NewRequest("GET", "/metrics", http.NoBody)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMetricsWithBasicAuthSuccess(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithBasicAuth(api.BasicAuth{
		Username: "user",
		Password: "password",
	}), api.WithMetrics())

	req, _ := http.NewRequest("GET", "/metrics", http.NoBody)
	req.SetBasicAuth("user", "password")
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
