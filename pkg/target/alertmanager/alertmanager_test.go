package alertmanager

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

func Test_AlertManagerClient_Send(t *testing.T) {
	t.Run("Send Single Alert", func(t *testing.T) {
		receivedAlerts := make([]Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			assert.Equal(t, "/api/v2/alerts", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))

			var alerts []Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts

			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := NewClient(Options{
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
			Category:    "test-category",
			Severity:    "high",
			Properties: map[string]string{
				"property1": "value1",
			},
		}

		client.Send(openreports.ORResultAdapter{ReportResult: result})

		require.Len(t, receivedAlerts, 1)
		alert := receivedAlerts[0]

		assert.Equal(t, map[string]string{
			"alertname": "PolicyReporterViolation",
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

	t.Run("Send Batch Alerts", func(t *testing.T) {
		receivedAlerts := make([]Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			assert.Equal(t, "/api/v2/alerts", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))

			var alerts []Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts

			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := NewClient(Options{
			ClientOptions: target.ClientOptions{
				Name: "test",
			},
			Host:       server.URL,
			HTTPClient: http.NewClient("", false),
		}).(*client)

		results := []v1alpha1.ReportResult{
			{
				Description: "test message 1",
				Policy:      "test-policy-1",
				Rule:        "test-rule-1",
				Result:      "fail",
				Source:      "test",
				Severity:    "high",
			},
			{
				Description: "test message 2",
				Policy:      "test-policy-2",
				Rule:        "test-rule-2",
				Result:      "fail",
				Source:      "test",
				Severity:    "high",
			},
		}

		// Create alerts directly instead of using BatchSend
		alerts := make([]Alert, 0, len(results))
		for _, result := range results {
			alerts = append(alerts, client.createAlert(openreports.ORResultAdapter{ReportResult: result}))
		}
		client.sendAlerts(alerts)

		require.Len(t, receivedAlerts, 2)
		assert.Equal(t, "test-policy-1", receivedAlerts[0].Labels["policy"])
		assert.Equal(t, "test-policy-2", receivedAlerts[1].Labels["policy"])
	})

	t.Run("Handle HTTP Error", func(t *testing.T) {
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			w.WriteHeader(stdhttp.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(Options{
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
		client.Send(openreports.ORResultAdapter{ReportResult: result})
	})

	t.Run("With Custom Headers", func(t *testing.T) {
		receivedHeaders := make(stdhttp.Header)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := NewClient(Options{
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

		client.Send(openreports.ORResultAdapter{ReportResult: result})

		assert.Equal(t, "Bearer test-token", receivedHeaders.Get("Authorization"))
		assert.Equal(t, "custom-value", receivedHeaders.Get("X-Custom"))
	})

	t.Run("With Custom Fields", func(t *testing.T) {
		receivedAlerts := make([]Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			var alerts []Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts
			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := NewClient(Options{
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

		client.Send(openreports.ORResultAdapter{ReportResult: result})

		require.Len(t, receivedAlerts, 1)
		assert.Equal(t, "production", receivedAlerts[0].Annotations["environment"])
		assert.Equal(t, "security", receivedAlerts[0].Annotations["team"])
	})

	t.Run("With Resource Information", func(t *testing.T) {
		receivedAlerts := make([]Alert, 0)
		server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			var alerts []Alert
			err := json.NewDecoder(r.Body).Decode(&alerts)
			require.NoError(t, err)
			receivedAlerts = alerts
			w.WriteHeader(stdhttp.StatusOK)
		}))
		defer server.Close()

		client := NewClient(Options{
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

		client.Send(openreports.ORResultAdapter{ReportResult: result})

		require.Len(t, receivedAlerts, 1)
		assert.Equal(t, "test-namespace/pod/test-pod", receivedAlerts[0].Annotations["resource"])
		assert.Equal(t, "Pod", receivedAlerts[0].Annotations["resource_kind"])
		assert.Equal(t, "test-pod", receivedAlerts[0].Annotations["resource_name"])
		assert.Equal(t, "test-namespace", receivedAlerts[0].Annotations["resource_namespace"])
		assert.Equal(t, "v1", receivedAlerts[0].Annotations["resource_apiversion"])
	})
}

func Test_AlertManagerClient_Type(t *testing.T) {
	client := NewClient(Options{
		ClientOptions: target.ClientOptions{
			Name: "test",
		},
		HTTPClient: http.NewClient("", false),
	})

	assert.Equal(t, target.BatchSend, client.Type())
}

// mockReportInterface is a simple implementation of ReportInterface for testing
type mockReportInterface struct {
	name      string
	namespace string
	scope     *corev1.ObjectReference
	results   []v1alpha1.ReportResult
}

func (m *mockReportInterface) GetName() string {
	return m.name
}

func (m *mockReportInterface) GetNamespace() string {
	return m.namespace
}

func (m *mockReportInterface) GetResults() []v1alpha1.ReportResult {
	return m.results
}

func (m *mockReportInterface) GetScope() *corev1.ObjectReference {
	return m.scope
}

func (m *mockReportInterface) SetResults(results []v1alpha1.ReportResult) {
	m.results = results
}

func (m *mockReportInterface) GetSummary() v1alpha1.ReportSummary {
	return v1alpha1.ReportSummary{}
}

func (m *mockReportInterface) GetSource() string {
	if len(m.results) == 0 {
		return ""
	}
	return m.results[0].Source
}

func (m *mockReportInterface) GetKinds() []string {
	return []string{}
}

func (m *mockReportInterface) GetSeverities() []string {
	return []string{}
}

func (m *mockReportInterface) HasResult(string) bool {
	return false
}

func (m *mockReportInterface) GetAnnotations() map[string]string {
	return map[string]string{}
}

func (m *mockReportInterface) GetCreationTimestamp() metav1.Time {
	return metav1.Now()
}
