package v2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/gin-gonic/gin"
	"github.com/kyverno/policy-reporter/pkg/api"
	v2 "github.com/kyverno/policy-reporter/pkg/api/v2"
	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/namespaces"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
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
}
