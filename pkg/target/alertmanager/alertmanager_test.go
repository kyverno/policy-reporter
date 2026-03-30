package alertmanager_test

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/alertmanager"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

func Test_AlertManagerClient_Send(t *testing.T) {
	t.Parallel()
	t.Run("Send Single Alert", func(t *testing.T) {
		t.Parallel()
		receivedAlerts := make([]alertmanager.Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			assert.Equal(t, "/api/v2/alerts", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))

			var alerts []alertmanager.Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts

			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := alertmanager.NewClient(alertmanager.Options{
			ClientOptions: target.ClientOptions{
				Name: "test",
			},
			Host:       server.URL,
			HTTPClient: http.NewClient("", false),
		})

		report := v1alpha1.Report{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-report",
				Namespace: "test-namespace",
			},
		}

		result := v1alpha1.ReportResult{
			Description: "test message",
			Policy:      "test-policy",
			Rule:        "test-rule",
			Result:      "fail",
			Source:      "test",
			Category:    "test-category",
			Severity:    "high",
			Properties: map[string]string{
				"property1": "value1",
			},
		}

		client.Send(&openreports.ReportAdapter{Report: &report}, openreports.ResultAdapter{ReportResult: result})

		require.Len(t, receivedAlerts, 1)
		alert := receivedAlerts[0]

		assert.Equal(t, map[string]string{
			"alertname": "PolicyReporterViolation",
			"name":      "test-report",
			"namespace": "test-namespace",
			"severity":  "high",
			"status":    "fail",
			"source":    "test",
			"policy":    "test-policy",
			"rule":      "test-rule",
		}, alert.Labels)

		assert.Equal(t, map[string]string{
			"message":   "test message",
			"category":  "test-category",
			"property1": "value1",
		}, alert.Annotations)

		assert.True(t, time.Since(alert.StartsAt) < time.Second)
		assert.True(t, alert.EndsAt.Sub(alert.StartsAt) == 24*time.Hour)
	})

	t.Run("Handle HTTP Error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			w.WriteHeader(stdhttp.StatusInternalServerError)
		}))
		defer server.Close()

		client := alertmanager.NewClient(alertmanager.Options{
			ClientOptions: target.ClientOptions{
				Name: "test",
			},
			Host:       server.URL,
			HTTPClient: http.NewClient("", false),
		})

		result := v1alpha1.ReportResult{
			Description: "test message",
			Policy:      "test-policy",
			Rule:        "test-rule",
			Result:      "fail",
			Source:      "test",
			Severity:    "high",
		}

		// Should not panic
		client.Send(&openreports.ReportAdapter{Report: &v1alpha1.Report{}}, openreports.ResultAdapter{ReportResult: result})
	})

	t.Run("With Custom Headers", func(t *testing.T) {
		t.Parallel()
		receivedHeaders := make(stdhttp.Header)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := alertmanager.NewClient(alertmanager.Options{
			ClientOptions: target.ClientOptions{
				Name: "test",
			},
			Host: server.URL,
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
				"X-Custom":      "custom-value",
			},
			HTTPClient: http.NewClient("", false),
		})

		result := v1alpha1.ReportResult{
			Description: "test message",
			Policy:      "test-policy",
			Rule:        "test-rule",
			Result:      "fail",
			Source:      "test",
			Severity:    "high",
		}

		client.Send(&openreports.ReportAdapter{Report: &v1alpha1.Report{}}, openreports.ResultAdapter{ReportResult: result})

		assert.Equal(t, "Bearer test-token", receivedHeaders.Get("Authorization"))
		assert.Equal(t, "custom-value", receivedHeaders.Get("X-Custom"))
	})

	t.Run("With Custom Fields", func(t *testing.T) {
		t.Parallel()
		receivedAlerts := make([]alertmanager.Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			var alerts []alertmanager.Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts
			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := alertmanager.NewClient(alertmanager.Options{
			ClientOptions: target.ClientOptions{
				Name: "test",
			},
			Host: server.URL,
			CustomFields: map[string]string{
				"environment": "production",
				"team":        "security",
			},
			HTTPClient: http.NewClient("", false),
		})

		result := v1alpha1.ReportResult{
			Description: "test message",
			Policy:      "test-policy",
			Rule:        "test-rule",
			Result:      "fail",
			Source:      "test",
			Severity:    "high",
		}

		client.Send(&openreports.ReportAdapter{Report: &v1alpha1.Report{}}, openreports.ResultAdapter{ReportResult: result})

		require.Len(t, receivedAlerts, 1)
		assert.Equal(t, "production", receivedAlerts[0].Annotations["environment"])
		assert.Equal(t, "security", receivedAlerts[0].Annotations["team"])
	})

	t.Run("With Resource Information", func(t *testing.T) {
		t.Parallel()
		receivedAlerts := make([]alertmanager.Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			var alerts []alertmanager.Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts
			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := alertmanager.NewClient(alertmanager.Options{
			ClientOptions: target.ClientOptions{
				Name: "test",
			},
			Host:       server.URL,
			HTTPClient: http.NewClient("", false),
		})

		result := v1alpha1.ReportResult{
			Description: "test message",
			Policy:      "test-policy",
			Rule:        "test-rule",
			Result:      "fail",
			Source:      "test",
			Severity:    "high",
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "test-pod",
					Namespace:  "test-namespace",
					UID:        "test-uid",
				},
			},
		}

		client.Send(&openreports.ReportAdapter{Report: &v1alpha1.Report{}}, openreports.ResultAdapter{ReportResult: result})

		require.Len(t, receivedAlerts, 1)
		assert.Equal(t, "test-namespace/pod/test-pod", receivedAlerts[0].Annotations["resource"])
		assert.Equal(t, "Pod", receivedAlerts[0].Annotations["resource_kind"])
		assert.Equal(t, "test-pod", receivedAlerts[0].Annotations["resource_name"])
		assert.Equal(t, "test-namespace", receivedAlerts[0].Annotations["resource_namespace"])
		assert.Equal(t, "v1", receivedAlerts[0].Annotations["resource_apiversion"])
	})
}

func Test_AlertManagerClient_Type(t *testing.T) {
	t.Parallel()
	client := alertmanager.NewClient(alertmanager.Options{
		ClientOptions: target.ClientOptions{
			Name: "test",
		},
		HTTPClient: http.NewClient("", false),
	})

	assert.Equal(t, target.BatchSend, client.Type())
}
