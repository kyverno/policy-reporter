package v1_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/sqlite3"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
)

var seconds = time.Date(2022, 9, 6, 0, 0, 0, 0, time.UTC).Unix()

var result1 = v1alpha2.PolicyReportResult{
	ID:        "123",
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "autogen-check-for-requests-and-limits",
	Priority:  v1alpha2.ErrorPriority,
	Result:    v1alpha2.StatusFail,
	Category:  "Best Practices",
	Severity:  v1alpha2.SeverityHigh,
	Scored:    true,
	Source:    "Kyverno",
	Timestamp: metav1.Timestamp{Seconds: seconds},
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
}

var result2 = v1alpha2.PolicyReportResult{
	ID:        "124",
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "autogen-check-for-requests-and-limits",
	Priority:  v1alpha2.WarningPriority,
	Result:    v1alpha2.StatusPass,
	Category:  "Best Practices",
	Scored:    true,
	Source:    "Kyverno",
	Timestamp: metav1.Timestamp{Seconds: seconds},
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
}

var cresult1 = v1alpha2.PolicyReportResult{
	ID:        "125",
	Message:   "validation error: The label `test` is required. Rule check-for-labels-on-namespace",
	Policy:    "require-ns-labels",
	Rule:      "check-for-labels-on-namespace",
	Priority:  v1alpha2.ErrorPriority,
	Result:    v1alpha2.StatusPass,
	Category:  "Convention",
	Severity:  v1alpha2.SeverityMedium,
	Scored:    true,
	Source:    "Kyverno",
	Timestamp: metav1.Timestamp{Seconds: seconds},
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188411",
	}},
}

var cresult2 = v1alpha2.PolicyReportResult{
	ID:       "126",
	Message:  "validation error: The label `test` is required. Rule check-for-labels-on-namespace",
	Policy:   "require-ns-labels",
	Rule:     "check-for-labels-on-namespace",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Category: "Convention",
	Severity: v1alpha2.SeverityHigh,
	Scored:   true,
	Source:   "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "dev",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188412",
	}},
}

var preport = &v1alpha2.PolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Labels:            map[string]string{"app": "policy-reporter", "scope": "namespace"},
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{result1, result2},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1},
}

var creport = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Labels:            map[string]string{"app": "policy-reporter", "scope": "cluster"},
		Name:              "cpolr",
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{cresult1, cresult2},
	Summary: v1alpha2.PolicyReportSummary{},
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

	handl := v1.NewHandler(store)

	t.Run("ClusterPolicyListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-policies", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.ClusterResourcesPolicyListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["require-ns-labels"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterRuleListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-rules", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.ClusterResourcesRuleListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["check-for-labels-on-namespace"]`
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
		handler := handl.NamespacedResourcesPolicyListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["require-requests-and-limits-required"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedRuleListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-rules", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.NamespacedResourcesRuleListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["autogen-check-for-requests-and-limits"]`
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
		handler := handl.CategoryListHandler()
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
		handler := handl.ClusterResourcesKindListHandler()
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
		handler := handl.NamespacedResourcesKindListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["Deployment","Pod"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterResourcesListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-resources/resources", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.ClusterResourcesListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"name":"dev","kind":"Namespace"},{"name":"test","kind":"Namespace"}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("NamespacedResourcesListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-resources/resources", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.NamespacedResourcesListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"name":"nginx","kind":"Deployment"},{"name":"nginx","kind":"Pod"}]`
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
		handler := handl.ClusterResourcesSourceListHandler()
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
		handler := handl.NamespacedSourceListHandler()
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
		handler := handl.ClusterResourcesStatusCountHandler()
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
		handler := handl.NamespacedResourcesStatusCountsHandler()
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
		handler := handl.RuleStatusCountHandler()
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
		req, err := http.NewRequest("GET", "/v1/namespaced-results?direction=desc", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.NamespacedResourcesResultHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"id":"123","namespace":"test","kind":"Deployment","apiVersion":"v1","name":"nginx","message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","category":"Best Practices","policy":"require-requests-and-limits-required","rule":"autogen-check-for-requests-and-limits","status":"fail","severity":"high","timestamp":1662422400},{"id":"124","namespace":"test","kind":"Pod","apiVersion":"v1","name":"nginx","message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","category":"Best Practices","policy":"require-requests-and-limits-required","rule":"autogen-check-for-requests-and-limits","status":"pass","timestamp":1662422400}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterResultHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-results?direction=desc", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.ClusterResourcesResultHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := "{\"id\":\"125\",\"kind\":\"Namespace\",\"apiVersion\":\"v1\",\"name\":\"test\",\"message\":\"validation error: The label `test` is required. Rule check-for-labels-on-namespace\",\"category\":\"Convention\",\"policy\":\"require-ns-labels\",\"rule\":\"check-for-labels-on-namespace\",\"status\":\"pass\",\"severity\":\"medium\",\"timestamp\":1662422400}"
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
		handler := handl.NamespaceListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `["test"]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterReportLabelListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/cluster-resources/report-labels?sources=kyverno", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.ClusterReportLabelListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{"app":["policy-reporter"],"scope":["cluster"]}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterReportLabelListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/namespaced-resources/report-labels?sources=kyverno", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.NamespacedReportLabelListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{"app":["policy-reporter"],"scope":["namespace"]}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("PolicyReportListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/policy-reports?namespaces=test&labels=app:policy-reporter", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.PolicyReportListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{"items":[{"id":"7605991845421273693","name":"polr-test","namespace":"test","source":"Kyverno","labels":{"app":"policy-reporter","scope":"namespace"},"pass":0,"skip":0,"warn":0,"error":0,"fail":1}],"count":1}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("ClusterPolicyReportListHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/policy-reports?labels=app:policy-reporter", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.ClusterPolicyReportListHandler()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{"items":[{"id":"7174304213499286261","name":"cpolr","source":"Kyverno","labels":{"app":"policy-reporter","scope":"cluster"},"pass":0,"skip":0,"warn":0,"error":0,"fail":0}],"count":1}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_TargetsAPI(t *testing.T) {
	handl := v1.NewHandler(nil)

	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := handl.TargetsHandler(make([]target.Client, 0))

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
		handler := handl.TargetsHandler([]target.Client{
			loki.NewClient(loki.Options{
				ClientOptions: target.ClientOptions{
					Name:                  "Loki",
					SkipExistingOnStartup: true,
				},
				HTTPClient: &http.Client{},
			}),
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
