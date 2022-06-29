package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/api"
)

func Test_HealthzAPI(t *testing.T) {
	t.Run("Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/healthz", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.HealthzHandler(func() bool { return true })

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})
	t.Run("Unavailable Response", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/healthz", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.HealthzHandler(func() bool { return false })

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusServiceUnavailable {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
		}
	})
}

func Test_ReadyAPI(t *testing.T) {
	t.Run("Success Response", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/ready", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.ReadyHandler(func() bool { return true })

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})
	t.Run("Unavailable Response", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/ready", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.ReadyHandler(func() bool { return false })

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusServiceUnavailable {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
		}
	})
}
