package pagerduty_test

import (
	"testing"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/pagerduty"
)

type mockPagerDutyClient struct {
	createCalls    int
	resolveCalls   int
	lastIncidentID string
	lastIncident   *pagerduty.CreateIncidentOptions
	lastResolve    *pagerduty.ManageIncidentsOptions
	shouldError    bool
}

func (m *mockPagerDutyClient) CreateIncident(from string, incident *pagerduty.CreateIncidentOptions) (*pagerduty.Incident, error) {
	m.createCalls++
	m.lastIncident = incident
	return &pagerduty.Incident{Id: "test-incident-id"}, nil
}

func (m *mockPagerDutyClient) ManageIncidents(from string, incident *pagerduty.ManageIncidentsOptions) error {
	m.resolveCalls++
	m.lastResolve = incident
	m.lastIncidentID = incident.ID
	return nil
}

func createTestResult(status v1alpha2.PolicyResult) v1alpha2.PolicyReportResult {
	return v1alpha2.PolicyReportResult{
		Policy:   "test-policy",
		Rule:     "test-rule",
		Message:  "test message",
		Result:   status,
		Severity: v1alpha2.SeverityHigh,
		Resources: []*corev1.ObjectReference{
			{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "test-ns",
				UID:      types.UID("test-uid"),
			},
		},
		Properties: map[string]string{
			"test-prop": "test-value",
		},
	}
}

func TestPagerDutyTarget(t *testing.T) {
	t.Run("Create incident for failing policy", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:     "test-token",
			ServiceID:    "test-service",
			CustomFields: map[string]string{"cluster": "test-cluster"},
		})
		// Replace internal PD client with mock
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)

		assert.Equal(t, 1, mockClient.createCalls)
		assert.Equal(t, 0, mockClient.resolveCalls)
		assert.Equal(t, "Policy Violation: test-policy", mockClient.lastIncident.Title)
		assert.Equal(t, "high", mockClient.lastIncident.Urgency)
	})

	t.Run("Do not create duplicate incidents", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		
		// Send same failing result twice
		client.Send(result)
		client.Send(result)

		assert.Equal(t, 1, mockClient.createCalls)
	})

	t.Run("Resolve incident when policy passes", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		// First send failing result
		failResult := createTestResult(v1alpha2.StatusFail)
		client.Send(failResult)

		// Then send passing result for same policy
		passResult := createTestResult(v1alpha2.StatusPass)
		client.Send(passResult)

		assert.Equal(t, 1, mockClient.createCalls)
		assert.Equal(t, 1, mockClient.resolveCalls)
		assert.Equal(t, "test-incident-id", mockClient.lastIncidentID)
	})

	t.Run("Ignore non-fail results", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusWarn)
		client.Send(result)

		assert.Equal(t, 0, mockClient.createCalls)
		assert.Equal(t, 0, mockClient.resolveCalls)
	})

	t.Run("Map severity to urgency", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		
		// Test high severity
		result.Severity = v1alpha2.SeverityHigh
		client.Send(result)
		assert.Equal(t, "high", mockClient.lastIncident.Urgency)

		// Test low severity
		result.Severity = v1alpha2.SeverityLow
		client.Send(result)
		assert.Equal(t, "low", mockClient.lastIncident.Urgency)
	})
} 