package pagerduty_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/pagerduty"
)

// mockPagerDutyClient implements a test double for the PagerDuty API client
type mockPagerDutyClient struct {
	createCalls    int                              // Number of incident creation attempts
	resolveCalls   int                             // Number of incident resolution attempts
	lastIncidentID string                          // ID of most recently managed incident
	lastIncident   *pagerduty.CreateIncidentOptions // Most recent incident creation request
	lastResolve    *pagerduty.ManageIncidentsOptions // Most recent incident resolution request
	shouldError    bool                            // Whether operations should fail
	failureCount   int                             // Number of times operations should fail
	currentTry     int                             // Current retry attempt count
	incidents      []*pagerduty.Incident           // List of existing incidents
}

func (m *mockPagerDutyClient) CreateIncident(from string, incident *pagerduty.CreateIncidentOptions) (*pagerduty.Incident, error) {
	m.createCalls++
	m.lastIncident = incident
	
	m.currentTry++
	if m.currentTry <= m.failureCount {
		return nil, fmt.Errorf("API error (attempt %d)", m.currentTry)
	}
	
	return &pagerduty.Incident{Id: "test-incident-id"}, nil
}

func (m *mockPagerDutyClient) ManageIncidents(from string, incident *pagerduty.ManageIncidentsOptions) error {
	m.resolveCalls++
	m.lastResolve = incident
	m.lastIncidentID = incident.ID

	m.currentTry++
	if m.currentTry <= m.failureCount {
		return fmt.Errorf("API error (attempt %d)", m.currentTry)
	}

	return nil
}

func (m *mockPagerDutyClient) ListIncidents(opts pagerduty.ListIncidentsOptions) ([]pagerduty.Incident, error) {
	return m.incidents, nil
}

// createTestResult creates a PolicyReportResult for testing
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

// TestPagerDutyTarget contains integration tests for the PagerDuty client
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

	t.Run("Retry failed create incident", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{failureCount: 2}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
			Retry: pagerduty.RetryConfig{
				MaxRetries:  3,
				InitialWait: 10 * time.Millisecond,
				MaxWait:     50 * time.Millisecond,
			},
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)

		assert.Equal(t, 3, mockClient.createCalls) // Initial try + 2 retries
		assert.NotEmpty(t, mockClient.lastIncident)
	})

	t.Run("Retry failed resolve incident", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{failureCount: 1}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
			Retry: pagerduty.RetryConfig{
				MaxRetries:  2,
				InitialWait: 10 * time.Millisecond,
				MaxWait:     50 * time.Millisecond,
			},
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		// First create incident
		failResult := createTestResult(v1alpha2.StatusFail)
		client.Send(failResult)

		// Reset failure count and current try for resolve
		mockClient.failureCount = 1
		mockClient.currentTry = 0

		// Then resolve it
		passResult := createTestResult(v1alpha2.StatusPass)
		client.Send(passResult)

		assert.Equal(t, 2, mockClient.resolveCalls) // Initial try + 1 retry
		assert.Equal(t, "test-incident-id", mockClient.lastIncidentID)
	})

	t.Run("Give up after max retries", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{failureCount: 5} // More failures than max retries
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
			Retry: pagerduty.RetryConfig{
				MaxRetries:  3,
				InitialWait: 10 * time.Millisecond,
				MaxWait:     50 * time.Millisecond,
			},
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)

		assert.Equal(t, 4, mockClient.createCalls) // Initial try + 3 retries
	})

	t.Run("Create separate incidents for different rules in same policy", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
			ClusterName: "test-cluster",
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		// Create two failures for same policy but different rules
		result1 := createTestResult(v1alpha2.StatusFail)
		result1.Rule = "rule-1"
		client.Send(result1)

		result2 := createTestResult(v1alpha2.StatusFail)
		result2.Rule = "rule-2"
		client.Send(result2)

		assert.Equal(t, 2, mockClient.createCalls, "Should create separate incidents for different rules")
		
		// Verify incident titles include both policy and rule
		assert.Contains(t, mockClient.lastIncident.Title, "Policy Violation: test-policy - Rule: rule-2")

		// Resolve one rule
		result1.Result = v1alpha2.StatusPass
		client.Send(result1)

		assert.Equal(t, 1, mockClient.resolveCalls, "Should only resolve the passing rule's incident")
	})

	t.Run("Handle deleted policy", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:     "test-token",
			ServiceID:    "test-service",
			CustomFields: map[string]string{"cluster": "test-cluster"},
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		// First create an incident
		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)

		// Simulate policy deletion by sending empty result
		result.Result = ""
		client.Send(result)

		assert.Equal(t, 1, mockClient.resolveCalls)
		assert.Contains(t, mockClient.lastResolve.Incidents[0].Resolution, 
			"Policy or resource has been deleted")
	})

	t.Run("Include cluster name in incidents", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "prod-cluster",
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)

		assert.Equal(t, 1, mockClient.createCalls)
		assert.Contains(t, mockClient.lastIncident.Title, "[prod-cluster]")
		assert.Equal(t, "prod-cluster", 
			mockClient.lastIncident.Body.Details.(map[string]interface{})["cluster"])
	})

	t.Run("Use default cluster name if not specified", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
			// ClusterName not set
		})
		client.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)

		assert.Equal(t, 1, mockClient.createCalls)
		assert.Contains(t, mockClient.lastIncident.Title, "[default]")
	})

	// Add more tests for HA scenarios...
}

// TestIncidentTracking verifies incident tracking and cleanup behavior
func TestIncidentTracking(t *testing.T) {
	t.Run("Cleanup old incidents when limit exceeded", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
		})
		pdClient := client.(*pagerduty.Client)
		pdClient.SetClient(mockClient)
		
		// Override tracking config for testing
		pdClient.SetTrackingConfig(pagerduty.IncidentTrackingConfig{
			MaxIncidents: 2,
			CleanupInterval: time.Hour,
		})

		// Create 3 incidents (exceeding limit of 2)
		for i := 1; i <= 3; i++ {
			result := createTestResult(v1alpha2.StatusFail)
			result.Policy = fmt.Sprintf("policy-%d", i)
			client.Send(result)
		}

		// Wait briefly for cleanup
		time.Sleep(100 * time.Millisecond)

		// Verify we only track the latest 2 incidents
		count := 0
		pdClient.IncidentMap().Range(func(key, value interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 2, count)
	})

	t.Run("Cleanup resolved incidents", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:  "test-token",
			ServiceID: "test-service",
		})
		pdClient := client.(*pagerduty.Client)
		pdClient.SetClient(mockClient)

		// Create and resolve an incident
		result := createTestResult(v1alpha2.StatusFail)
		client.Send(result)
		
		result.Result = v1alpha2.StatusPass
		client.Send(result)

		// Verify incident was removed from tracking
		count := 0
		pdClient.IncidentMap().Range(func(key, value interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count)
	})
}

// TestHAScenarios verifies behavior in high-availability deployments
func TestHAScenarios(t *testing.T) {
	t.Run("Deduplicate incidents across replicas", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client1 := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty-1",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "test-cluster",
		})
		client1.(*pagerduty.Client).SetClient(mockClient)

		client2 := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty-2",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "test-cluster",
		})
		client2.(*pagerduty.Client).SetClient(mockClient)

		// Simulate existing incident
		existingIncident := pagerduty.Incident{
			Id: "existing-incident",
			Body: &pagerduty.APIDetails{
				Details: map[string]interface{}{
					"cluster":            "test-cluster",
					"policy":             "test-policy",
					"rule":              "test-rule",
					"resource_kind":      "Pod",
					"resource_namespace": "test-ns",
					"resource_name":      "test-pod",
				},
			},
		}
		mockClient.incidents = []pagerduty.Incident{existingIncident}

		// Send same violation to both clients
		result := createTestResult(v1alpha2.StatusFail)
		client1.Send(result)
		client2.Send(result)

		assert.Equal(t, 0, mockClient.createCalls, 
			"Should not create duplicate incidents when violation reported to multiple replicas")
	})

	t.Run("Handle concurrent resolution and recreation", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		client1 := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty-1",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "test-cluster",
		})
		client1.(*pagerduty.Client).SetClient(mockClient)

		client2 := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty-2",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "test-cluster",
		})
		client2.(*pagerduty.Client).SetClient(mockClient)

		// Create initial incident
		result := createTestResult(v1alpha2.StatusFail)
		client1.Send(result)
		assert.Equal(t, 1, mockClient.createCalls)

		// Simulate existing incident
		existingIncident := pagerduty.Incident{
			Id: "test-incident-id",
			Body: &pagerduty.APIDetails{
				Details: map[string]interface{}{
					"cluster":            "test-cluster",
					"policy":             result.Policy,
					"rule":              result.Rule,
					"resource_kind":      "Pod",
					"resource_namespace": "test-ns",
					"resource_name":      "test-pod",
				},
			},
		}
		mockClient.incidents = []pagerduty.Incident{existingIncident}

		// Simulate concurrent resolution and new violation
		resultPass := createTestResult(v1alpha2.StatusPass)
		resultFail := createTestResult(v1alpha2.StatusFail)

		// Send pass to client1 and fail to client2
		client1.Send(resultPass)
		client2.Send(resultFail)

		assert.Equal(t, 1, mockClient.resolveCalls, 
			"Should resolve incident when violation is fixed")
		assert.Equal(t, 2, mockClient.createCalls, 
			"Should create new incident after resolution")
	})

	t.Run("Handle pod restarts", func(t *testing.T) {
		mockClient := &mockPagerDutyClient{}
		
		// Create initial client and incident
		client1 := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "test-cluster",
		})
		client1.(*pagerduty.Client).SetClient(mockClient)

		result := createTestResult(v1alpha2.StatusFail)
		client1.Send(result)
		assert.Equal(t, 1, mockClient.createCalls)

		// Simulate existing incident
		existingIncident := pagerduty.Incident{
			Id: "test-incident-id",
			Body: &pagerduty.APIDetails{
				Details: map[string]interface{}{
					"cluster":            "test-cluster",
					"policy":             result.Policy,
					"rule":              result.Rule,
					"resource_kind":      "Pod",
					"resource_namespace": "test-ns",
					"resource_name":      "test-pod",
				},
			},
		}
		mockClient.incidents = []pagerduty.Incident{existingIncident}

		// Simulate pod restart by creating new client
		client2 := pagerduty.NewClient(pagerduty.Options{
			ClientOptions: target.ClientOptions{
				Name: "test-pagerduty",
			},
			APIToken:    "test-token",
			ServiceID:   "test-service",
			ClusterName: "test-cluster",
		})
		client2.(*pagerduty.Client).SetClient(mockClient)

		// Send same violation to new client
		client2.Send(result)
		assert.Equal(t, 1, mockClient.createCalls, 
			"Should not create duplicate incident after pod restart")

		// Resolve violation on new client
		resultPass := createTestResult(v1alpha2.StatusPass)
		client2.Send(resultPass)
		assert.Equal(t, 1, mockClient.resolveCalls, 
			"Should resolve incident after pod restart")
	})
} 