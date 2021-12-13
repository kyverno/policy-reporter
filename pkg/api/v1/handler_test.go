package v1_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/sqlite3"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
)

var result1 = &report.Result{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "Best Practices",
	Severity: report.High,
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

var result2 = &report.Result{
	ID:       "124",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Pass,
	Category: "Best Practices",
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	},
}

var cresult1 = &report.Result{
	ID:       "125",
	Message:  "validation error: The label `test` is required. Rule check-for-labels-on-namespace",
	Policy:   "require-ns-labels",
	Rule:     "check-for-labels-on-namespace",
	Priority: report.ErrorPriority,
	Status:   report.Pass,
	Category: "Convention",
	Severity: report.Medium,
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188411",
	},
}

var cresult2 = &report.Result{
	ID:       "126",
	Message:  "validation error: The label `test` is required. Rule check-for-labels-on-namespace",
	Policy:   "require-ns-labels",
	Rule:     "check-for-labels-on-namespace",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Category: "Convention",
	Severity: report.High,
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "dev",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188412",
	},
}

var preport = &report.PolicyReport{
	ID:        report.GeneratePolicyReportID("polr-test", "test"),
	Name:      "polr-test",
	Namespace: "test",
	Results: map[string]*report.Result{
		result1.GetIdentifier(): result1,
		result2.GetIdentifier(): result2,
	},
	Summary:           &report.Summary{Fail: 1},
	CreationTimestamp: time.Now(),
}

var creport = &report.PolicyReport{
	ID:   report.GeneratePolicyReportID("cpolr", ""),
	Name: "cpolr",
	Results: map[string]*report.Result{
		cresult1.GetIdentifier(): cresult1,
		cresult2.GetIdentifier(): cresult2,
	},
	Summary:           &report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_V1_API(t *testing.T) {
	db, err := sqlite3.NewDatabase("test.db")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	if err != nil {
		t.Fatal(err)
	}
	store, err := sqlite3.NewPolicyReportStore(db)
	if err != nil {
		t.Fatal(err)
	}
	defer store.CleanUp()

	store.Add(preport)
	store.Add(creport)

	t.Run("ClusterPolicyListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-policies", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.ClusterResourcesPolicyListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["require-ns-labels"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedPolicyListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-policies", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.NamespacedResourcesPolicyListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["require-requests-and-limits-required"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("CategoryListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/categories", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.CategoryListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["Best Practices","Convention"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterKindListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-kinds", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.ClusterResourcesKindListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["Namespace"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedKindListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-kinds", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.NamespacedResourcesKindListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["Deployment","Pod"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterSourceListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-sources", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.ClusterResourcesSourceListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["Kyverno"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedSourceListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namspaced-sources", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.NamespacedSourceListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["Kyverno"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterStatusCountHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-status-counts?status=pass", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.ClusterResourcesStatusCountHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"status":"pass","count":1}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedStatusCountHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-status-counts?status=pass", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.NamespacedResourcesStatusCountsHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"status":"pass","items":[{"namespace":"test","count":1}]}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("RuleStatusCountHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/rule-status-count?policy=require-requests-and-limits-required&rule=autogen-check-for-requests-and-limits", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.RuleStatusCountHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{"status":"fail","count":1}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}

		expected = `{"status":"pass","count":1}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}

		expected = `{"status":"warn","count":0}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedResultHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-results", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.NamespacedResourcesResultHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"id":"123","namespace":"test","kind":"Deployment","name":"nginx","message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","policy":"require-requests-and-limits-required","rule":"autogen-check-for-requests-and-limits","status":"fail","severity":"high"},{"id":"124","namespace":"test","kind":"Pod","name":"nginx","message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","policy":"require-requests-and-limits-required","rule":"autogen-check-for-requests-and-limits","status":"pass"}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterResultHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-results", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.ClusterResourcesResultHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := "{\"id\":\"125\",\"kind\":\"Namespace\",\"name\":\"test\",\"message\":\"validation error: The label `test` is required. Rule check-for-labels-on-namespace\",\"policy\":\"require-ns-labels\",\"rule\":\"check-for-labels-on-namespace\",\"status\":\"pass\",\"severity\":\"medium\"}"
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespaceListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaces", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.NamespaceListHandler(store)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["test"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_TargetsAPI(t *testing.T) {
	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.TargetsHandler(make([]target.Client, 0))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := "[]"

		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := v1.TargetsHandler([]target.Client{
			loki.NewClient("", "", []string{}, true, &http.Client{}),
		})

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
