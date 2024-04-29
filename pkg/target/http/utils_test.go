package http_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

func TestResultMapping(t *testing.T) {
	result := http.NewJSONResult(fixtures.CompleteTargetSendResult)

	assert.Equal(t, result.Message, fixtures.CompleteTargetSendResult.Message)
	assert.Equal(t, result.Policy, fixtures.CompleteTargetSendResult.Policy)
	assert.Equal(t, result.Rule, fixtures.CompleteTargetSendResult.Rule)
	assert.Equal(t, result.Resource.Name, fixtures.CompleteTargetSendResult.Resources[0].Name)
}

func TestCreateJSONRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req, err := http.CreateJSONRequest("Test", "GET", "http://localhost:8080", []string{"test"})

		assert.Nil(t, err)

		list := make([]string, 0)

		json.NewDecoder(req.Body).Decode(&list)

		assert.Equal(t, []string{"test"}, list)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "application/json; charset=utf-8", req.Header.Get("Content-Type"))
		assert.Equal(t, "Policy-Reporter", req.Header.Get("User-Agent"))
	})

	t.Run("error", func(t *testing.T) {
		_, err := http.CreateJSONRequest("Test", "GET", "\test", []string{"test"})

		assert.NotNil(t, err)
	})
}

func TestClient(t *testing.T) {
	assert.NotNil(t, http.NewClient("", true))
}

func TestProcessHTTPResponse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		obs, logs := observer.New(zap.InfoLevel)

		zap.ReplaceGlobals(zap.New(obs))

		w := httptest.NewRecorder()
		w.Write([]byte(`["test"]`))

		http.ProcessHTTPResponse("Test", w.Result(), nil)

		assert.Equal(t, 1, logs.Len())
		assert.Equal(t, 1, logs.FilterLevelExact(zap.InfoLevel).Len())
	})
	t.Run("error", func(t *testing.T) {
		obs, logs := observer.New(zap.InfoLevel)

		zap.ReplaceGlobals(zap.New(obs))

		w := httptest.NewRecorder()
		w.Write([]byte(`["test"]`))

		http.ProcessHTTPResponse("Test", w.Result(), errors.New("error"))

		assert.Equal(t, 1, logs.Len())
		assert.Equal(t, 1, logs.FilterMessage("Test: PUSH FAILED").Len())
	})
	t.Run("error status code", func(t *testing.T) {
		obs, logs := observer.New(zap.InfoLevel)

		zap.ReplaceGlobals(zap.New(obs))

		w := httptest.NewRecorder()
		resp := w.Result()
		resp.StatusCode = 404

		http.ProcessHTTPResponse("Test", w.Result(), nil)

		assert.Equal(t, 1, logs.Len())
		assert.Equal(t, 1, logs.FilterMessage("Test: PUSH FAILED").Len())
	})
}
