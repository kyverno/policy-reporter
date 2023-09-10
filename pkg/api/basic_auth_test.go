package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/api"
)

func Test_HTTPBasicSkipped(t *testing.T) {
	handler := api.HTTPBasic(nil, api.HealthzHandler(func() bool { return true }))

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func Test_HTTPBasicUnauthorized(t *testing.T) {
	handler := api.HTTPBasic(&api.BasicAuth{Username: "user", Password: "password"}, api.HealthzHandler(func() bool { return true }))

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	if rr.Header().Get("WWW-Authenticate") == "" {
		t.Errorf("Missing Header: WWW-Authenticate")
	}
}

func Test_HTTPBasicAuthorized(t *testing.T) {
	handler := api.HTTPBasic(&api.BasicAuth{Username: "user", Password: "password"}, api.HealthzHandler(func() bool { return true }))

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.SetBasicAuth("user", "password")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
