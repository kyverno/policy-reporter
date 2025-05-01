package main

import (
	"fmt"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/alertmanager"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

func main() {
	// Create a test result
	result := v1alpha2.PolicyReportResult{
		Message:  "Test policy violation from manual test",
		Policy:   "test-policy",
		Rule:     "test-rule",
		Result:   "fail",
		Source:   "policy-reporter",
		Category: "test-category",
		Severity: "warning",
		Properties: map[string]string{
			"test_property": "test_value",
		},
	}

	// Create the AlertManager client
	// Replace with your AlertManager URL
	alertManagerURL := "http://localhost:9093"

	client := alertmanager.NewClient(alertmanager.Options{
		ClientOptions: target.ClientOptions{
			Name: "test-client",
		},
		Host: alertManagerURL,
		Headers: map[string]string{
			"X-Test-Header": "test-value",
		},
		CustomFields: map[string]string{
			"environment": "test",
			"test":        "true",
		},
		HTTPClient: http.NewClient("", false),
	})

	// Send the test alert
	fmt.Println("Sending test alert to", alertManagerURL)
	client.Send(result)

	// Also test batch send
	results := []v1alpha2.PolicyReportResult{
		{
			Message:  "Batch test alert 1",
			Policy:   "batch-policy-1",
			Rule:     "batch-rule-1",
			Result:   "fail",
			Source:   "policy-reporter",
			Severity: "warning",
		},
		{
			Message:  "Batch test alert 2",
			Policy:   "batch-policy-2",
			Rule:     "batch-rule-2",
			Result:   "fail",
			Source:   "policy-reporter",
			Severity: "warning",
		},
	}

	fmt.Println("Sending batch test alerts")
	client.BatchSend(nil, results)

	// Wait a moment to ensure alerts are sent before the program exits
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Test completed. Check AlertManager UI to verify alerts were received.")
}
