package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/api"
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

type testHandler struct{}

func (h *testHandler) Register(group *gin.RouterGroup) error {
	group.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	return nil
}

func TestWithCustomHandler(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithProfiling())
	server.Register("/test", &testHandler{})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
}

func TestWithRecover(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	server := api.NewServer(engine, api.WithRecovery())

	engine.GET("/recover", func(ctx *gin.Context) {
		panic("recover")
	})

	req, _ := http.NewRequest("GET", "/recover", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusInternalServerError, w.Code)
}

func TestWithZapLoggingRecover(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	server := api.NewServer(engine, api.WithLogging(zap.L()))

	engine.GET("/recover", func(ctx *gin.Context) {
		panic("recover")
	})

	req, _ := http.NewRequest("GET", "/recover", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusInternalServerError, w.Code)
}

func TestWithPort(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), api.WithProfiling(), api.WithPort(8082))
	req, _ := http.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
}
