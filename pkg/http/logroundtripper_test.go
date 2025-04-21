package http_test

import (
	net "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/kyverno/policy-reporter/pkg/http"
)

type mock struct{}

func (rt mock) RoundTrip(req *net.Request) (*net.Response, error) {
	return httptest.NewRecorder().Result(), nil
}

func TestDebug(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)

	zap.ReplaceGlobals(zap.New(obs))

	r := http.NewLoggingRoundTripper(mock{})

	_, err := r.RoundTrip(httptest.NewRequest("GET", "http://localhost:8080/healthz", nil))

	assert.Nil(t, err)

	assert.Equal(t, 2, logs.FilterLevelExact(zap.DebugLevel).Len())
}
