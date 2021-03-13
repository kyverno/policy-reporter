package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/api"
	"github.com/fjogeleit/policy-reporter/pkg/report"
)

func Test_TargetsAPI(t *testing.T) {
	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.TargetsHandler(make([]api.Target, 0)))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.TargetsHandler([]api.Target{
			{Name: "Loki", MinimumPriority: "debug", SkipExistingOnStartup: true},
		}))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"name":"Loki","minimumPriority":"debug","skipExistingOnStartup":true}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_PolicyReportAPI(t *testing.T) {
	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/policy-reports", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.PolicyReportHandler(report.NewPolicyReportStore()))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/policy-reports", nil)
		if err != nil {
			t.Fatal(err)
		}

		preport := report.PolicyReport{
			Name:              "polr-test",
			Namespace:         "test",
			Results:           make(map[string]report.Result, 0),
			Summary:           report.Summary{},
			CreationTimestamp: time.Now(),
		}

		store := report.NewPolicyReportStore()
		store.Add(preport)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.PolicyReportHandler(store))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"name":"polr-test","namespace":"test","results":[],"summary":{"pass":0,"skip":0,"warn":0,"error":0,"fail":0}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_ClusterPolicyReportAPI(t *testing.T) {
	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/cluster-policy-reports", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.ClusterPolicyReportHandler(report.NewClusterPolicyReportStore()))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/cluster-policy-reports", nil)
		if err != nil {
			t.Fatal(err)
		}

		creport := report.ClusterPolicyReport{
			Name:              "cpolr-test",
			Results:           make(map[string]report.Result, 0),
			Summary:           report.Summary{},
			CreationTimestamp: time.Now(),
		}

		store := report.NewClusterPolicyReportStore()
		store.Add(creport)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.ClusterPolicyReportHandler(store))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"name":"cpolr-test","results":[],"summary":{"pass":0,"skip":0,"warn":0,"error":0,"fail":0}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}
