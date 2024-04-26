package v2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/api"
	v2 "github.com/kyverno/policy-reporter/pkg/api/v2"
	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/namespaces"
)

const (
	nsDefault = "default"
	nsTest    = "test"
)

func newFakeClient() v1.NamespaceInterface {
	return fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nsDefault,
				Labels: map[string]string{
					"team":  "team-a",
					"group": "all",
				},
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nsTest,
				Labels: map[string]string{
					"team":  "team-b",
					"group": "all",
				},
			},
		},
	).CoreV1().Namespaces()
}

func TestV2(t *testing.T) {
	db, err := database.NewSQLiteDB("db_v2.db")
	if err != nil {
		assert.Fail(t, "failed to init SQLite DB")
	}

	store, err := database.NewStore(db, "1.0")
	if err != nil {
		assert.Fail(t, "failed to init Store")
	}

	if err := store.PrepareDatabase(context.Background()); err != nil {
		assert.Fail(t, "failed to prepare Store")
	}

	store.Add(context.Background(), fixtures.DefaultPolicyReport)
	store.Add(context.Background(), fixtures.KyvernoPolicyReport)
	store.Add(context.Background(), fixtures.KyvernoClusterPolicyReport)

	client := namespaces.NewClient(newFakeClient(), cache.New(time.Second, time.Second))

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), v2.WithAPI(store, client, config.Targets{
		Webhook: &config.Target[config.WebhookOptions]{
			Name:            "Webhook",
			MinimumPriority: "warn",
			Config: &config.WebhookOptions{
				Webhook: "http://localhost:8080",
			},
		},
	}))

	t.Run("TargetResponse", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/targets", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ResolveNamespaces", func(t *testing.T) {
		body := new(bytes.Buffer)
		body.Write([]byte(`{"team":"team-a"}`))

		req, _ := http.NewRequest("POST", "/v2/namespaces/resolve-selector", body)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := make([]string, 0, 1)

		json.NewDecoder(w.Body).Decode(&resp)

		assert.Equal(t, 1, len(resp))
	})

	t.Run("ListNamespaces", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/namespaces", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0, 1)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 2, len(resp))
			assert.Contains(t, resp, "test")
			assert.Contains(t, resp, "kyverno")
		}
	})

	t.Run("ListSources", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/sources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0, 1)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 2, len(resp))
			assert.Contains(t, resp, "test")
			assert.Contains(t, resp, "Kyverno")
		}
	})

	t.Run("ListPolicies", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/policies", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v2.Policy, 0, 1)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 4, len(resp))
			assert.Contains(t, resp, v2.Policy{Source: "test", Category: "Other", Name: "priority-test", Severity: "", Results: map[string]int{"fail": 1}})
			assert.Contains(t, resp, v2.Policy{Source: "Kyverno", Category: "test", Name: "cluster-required-quota", Severity: "high", Results: map[string]int{"fail": 1}})
			assert.Contains(t, resp, v2.Policy{Source: "Kyverno", Category: "test", Name: "required-limit", Severity: "high", Results: map[string]int{"pass": 1, "warn": 1}})
			assert.Contains(t, resp, v2.Policy{Source: "test", Category: "test", Name: "required-label", Severity: "high", Results: map[string]int{"fail": 2}})
		}
	})

	t.Run("UseResources", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/sources/Kyverno/use-resources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make(map[string]bool)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.True(t, resp["resources"])
		}
	})

	t.Run("ListSourceWithCategories", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/sources/categories", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v2.SourceDetails, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Contains(t, resp, v2.SourceDetails{Name: "Kyverno", Categories: []*v2.Category{{Name: "test", Pass: 1, Warn: 1, Fail: 1}}})
			assert.Contains(t, resp, v2.SourceDetails{Name: "test", Categories: []*v2.Category{{Name: "Other", Fail: 1}, {Name: "test", Fail: 2}}})
		}
	})

	t.Run("ListResourceCategories", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/resource/17962226559046503697/source-categories", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v2.SourceDetails, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, resp[0], v2.SourceDetails{Name: "test", Categories: []*v2.Category{{Name: "test", Fail: 1}}})
		}
	})

	t.Run("GetResource", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/resource/17962226559046503697", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Resource{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, resp, v2.Resource{ID: "17962226559046503697", UID: "dfd57c50-f30c-4729-b63f-b1954d8988d1", Namespace: "test", Name: "nginx", Kind: "Deployment", APIVersion: "v1"})
		}
	})

	t.Run("GetResourceStatusCounts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/resource/17962226559046503697/status-counts", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v2.ResourceStatusCount, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Contains(t, resp, v2.ResourceStatusCount{Source: "test", Fail: 1})
		}
	})

	t.Run("ListNamespaceResourceResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/namespace-scoped/resource-results?namespaces=kyverno", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Paginated[v2.ResourceResult]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, resp.Count, 2)
			assert.Contains(t, resp.Items, v2.ResourceResult{ID: "6274512523942114905", UID: "dfd57c50-f30c-4729-b63f-b1954d8988d1", Name: "nginx", Kind: "Deployment", APIVersion: "v1", Namespace: "kyverno", Pass: 1})
			assert.Contains(t, resp.Items, v2.ResourceResult{ID: "8277600851619588241", UID: "dfd57c50-f30c-4729-b63f-b1954d8988d2", Name: "nginx2", Kind: "Deployment", APIVersion: "v1", Namespace: "kyverno", Warn: 1})
		}
	})

	t.Run("ListClusterResourceResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/cluster-scoped/resource-results", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Paginated[v2.ResourceResult]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, resp.Count, 1)
			assert.Equal(t, resp.Items[0], v2.ResourceResult{ID: "11786270724827677857", UID: "dfd57c50-f30c-4729-b63f-b1954d8988d1", Name: "kyverno", Kind: "Namespace", APIVersion: "v1", Source: "", Fail: 1})
		}
	})

	t.Run("GetClusterStatusCounts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/cluster-scoped/Kyverno/status-counts", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make(map[string]int, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 5, len(resp))
			assert.Equal(t, 1, resp["fail"])
		}
	})

	t.Run("GetNamespaceStatusCounts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/namespace-scoped/Kyverno/status-counts", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make(map[string]map[string]int, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.Equal(t, 5, len(resp["kyverno"]))
			assert.Equal(t, 1, resp["kyverno"]["pass"])
			assert.Equal(t, 1, resp["kyverno"]["warn"])
			assert.Equal(t, 0, resp["kyverno"]["fail"])
			assert.Equal(t, 0, resp["kyverno"]["error"])
			assert.Equal(t, 0, resp["kyverno"]["skip"])
		}
	})

	t.Run("ListClusterKinds", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/cluster-scoped/kinds", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.Equal(t, "Namespace", resp[0])
		}
	})

	t.Run("ListNamespaceKinds", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/namespace-scoped/kinds", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.Equal(t, "Deployment", resp[0])
		}
	})

	t.Run("ListResourceResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/resource/6274512523942114905/resource-results", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v2.ResourceResult, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.Equal(t, resp[0], v2.ResourceResult{ID: "6274512523942114905", UID: "dfd57c50-f30c-4729-b63f-b1954d8988d1", Name: "nginx", Kind: "Deployment", APIVersion: "v1", Namespace: "kyverno", Source: "Kyverno", Pass: 1})
		}
	})

	t.Run("ListResourcePolilcyResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/resource/6274512523942114905/results", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Paginated[v2.PolicyResult]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, resp.Count)
			assert.Equal(t, resp.Items[0], v2.PolicyResult{ID: "14158407137220160684", ResourceID: "6274512523942114905", Severity: "high", Name: "nginx", Kind: "Deployment", APIVersion: "v1", Namespace: "kyverno", Message: "message", Category: "test", Policy: "required-limit", Rule: "resource-limit-required", Status: "pass", Timestamp: 1614093003})
		}
	})

	t.Run("ListPolicyResults Namespaced", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/namespace-scoped/results?namespaces=kyverno", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Paginated[v2.PolicyResult]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 2, resp.Count)
			assert.Equal(t, resp.Items[0], v2.PolicyResult{ID: "14158407137220160684", ResourceID: "6274512523942114905", Severity: "high", Name: "nginx", Kind: "Deployment", APIVersion: "v1", Namespace: "kyverno", Message: "message", Category: "test", Policy: "required-limit", Rule: "resource-limit-required", Status: "pass", Timestamp: 1614093003})
			assert.Equal(t, resp.Items[1], v2.PolicyResult{ID: "2079631062832497014", ResourceID: "8277600851619588241", Severity: "high", Name: "nginx2", Kind: "Deployment", APIVersion: "v1", Namespace: "kyverno", Message: "message", Category: "test", Policy: "required-limit", Rule: "resource-limit-required", Status: "warn", Timestamp: 1614093003})
		}
	})

	t.Run("ListPolicyResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/cluster-scoped/results", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Paginated[v2.PolicyResult]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, resp.Count)
			assert.Equal(t, resp.Items[0], v2.PolicyResult{ID: "16800058481201255747", ResourceID: "11786270724827677857", Severity: "high", Name: "kyverno", Kind: "Namespace", APIVersion: "v1", Namespace: "", Message: "message", Category: "test", Policy: "cluster-required-quota", Rule: "ns-quota-required", Status: "fail", Timestamp: 1614093000})
		}
	})

	t.Run("ListResultsWithoutResource", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/results-without-resources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := v2.Paginated[v2.PolicyResult]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, resp.Count)
			assert.Equal(t, resp.Items[0], v2.PolicyResult{ID: "8115731892871392633", ResourceID: "18007334074686647077", Severity: "", Name: "", Kind: "", APIVersion: "", Namespace: "test", Message: "message 2", Category: "", Policy: "priority-test", Rule: "", Status: "fail", Timestamp: 1614093000})
		}
	})

	t.Run("UseResources", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v2/sources/Kyverno/use-resources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make(map[string]bool, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.True(t, resp["resources"])
		}
	})
}
