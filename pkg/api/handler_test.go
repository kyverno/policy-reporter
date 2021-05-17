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

		result := report.Result{
			Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
			Policy:   "require-requests-and-limits-required",
			Rule:     "autogen-check-for-requests-and-limits",
			Priority: report.ErrorPriority,
			Status:   report.Fail,
			Category: "resources",
			Scored:   true,
			Resource: report.Resource{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "nginx",
				Namespace:  "test",
				UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
			},
		}

		preport := report.PolicyReport{
			Name:              "polr-test",
			Namespace:         "test",
			Results:           map[string]report.Result{"": result},
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

		expected := `[{"name":"polr-test","namespace":"test","results":[{"message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","policy":"require-requests-and-limits-required","rule":"autogen-check-for-requests-and-limits","priority":"error","status":"fail","category":"resources","scored":true,"resource":{"apiVersion":"v1","kind":"Deployment","name":"nginx","namespace":"test","uid":"536ab69f-1b3c-4bd9-9ba4-274a56188409"}}],"summary":{"pass":0,"skip":0,"warn":0,"error":0,"fail":0}`
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
		handler := http.HandlerFunc(api.ClusterPolicyReportHandler(report.NewPolicyReportStore()))

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

		result := report.Result{
			Message:  "validation error: Namespace label missing",
			Policy:   "ns-label-env-required",
			Rule:     "ns-label-required",
			Priority: report.ErrorPriority,
			Status:   report.Fail,
			Category: "resources",
			Scored:   true,
			Resource: report.Resource{
				APIVersion: "v1",
				Kind:       "Namespace",
				Name:       "dev",
				UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
			},
		}

		creport := report.PolicyReport{
			Name:              "cpolr-test",
			Summary:           report.Summary{},
			CreationTimestamp: time.Now(),
			Results:           map[string]report.Result{"": result},
		}

		store := report.NewPolicyReportStore()
		store.Add(creport)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.ClusterPolicyReportHandler(store))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"name":"cpolr-test","results":[{"message":"validation error: Namespace label missing","policy":"ns-label-env-required","rule":"ns-label-required","priority":"error","status":"fail","category":"resources","scored":true,"resource":{"apiVersion":"v1","kind":"Namespace","name":"dev","uid":"536ab69f-1b3c-4bd9-9ba4-274a56188409"}}],"summary":{"pass":0,"skip":0,"warn":0,"error":0,"fail":0}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}
