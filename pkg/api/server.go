package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServerOption func(s *Server) error

type Handler interface {
	Register(group *gin.RouterGroup) error
}

type BasicAuth struct {
	Username string
	Password string
}

type Server struct {
	middleware     []gin.HandlerFunc
	restMiddleware []gin.HandlerFunc
	engine         *gin.Engine
	port           int
}

func (s *Server) Start() error {
	return s.engine.Run(fmt.Sprintf(":%d", s.port))
}

func (s *Server) Serve(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

func (s *Server) Register(path string, handler Handler) error {
	middleware := append([]gin.HandlerFunc{}, s.middleware...)
	middleware = append(middleware, s.restMiddleware...)
	return handler.Register(s.engine.Group(path, middleware...))
}

func NewServer(engine *gin.Engine, options ...ServerOption) *Server {
	server := &Server{
		engine: engine,
		port:   8080,
	}

	for _, opt := range options {
		if err := opt(server); err != nil {
			zap.L().Error("failed to apply server function", zap.Error(err))
		}
	}

	return server
}

func WithBasicAuth(auth BasicAuth) ServerOption {
	return func(s *Server) error {
		s.middleware = append(s.middleware, gin.BasicAuth(gin.Accounts{
			auth.Username: auth.Password,
		}))

		return nil
	}
}

func WithHealthChecks(checks []HealthCheck, readinessChecks ...HealthCheck) ServerOption {
	return func(s *Server) error {
		s.engine.GET("healthz", HealthzHandler(checks))
		s.engine.GET("ready", HealthzHandler(append(append([]HealthCheck{}, checks...), readinessChecks...)))

		return nil
	}
}

func WithLogging(logger *zap.Logger) ServerOption {
	return func(s *Server) error {
		s.engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
		s.engine.Use(ginzap.RecoveryWithZap(logger, true))

		return nil
	}
}

func WithGZIP() ServerOption {
	return func(s *Server) error {
		s.engine.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/metrics"})))

		return nil
	}
}

func WithRecovery() ServerOption {
	return func(s *Server) error {
		s.engine.Use(gin.Recovery())

		return nil
	}
}

func WithPort(port int) ServerOption {
	return func(s *Server) error {
		s.port = port

		return nil
	}
}

func WithProfiling() ServerOption {
	return func(s *Server) error {
		pprof.Register(s.engine)

		return nil
	}
}

func WithMetrics() ServerOption {
	return func(s *Server) error {
		s.engine.GET("metrics", append(s.middleware, MetricsHandler())...)

		return nil
	}
}

// WithRESTReadiness protects REST groups without gating metrics or profiling.
func WithRESTReadiness(check HealthCheck) ServerOption {
	return func(s *Server) error {
		s.restMiddleware = append(s.restMiddleware, func(ctx *gin.Context) {
			if err := check(); err != nil {
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}
			ctx.Next()
		})
		return nil
	}
}
