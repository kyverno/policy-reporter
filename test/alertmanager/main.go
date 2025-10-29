package main

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/alertmanager"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

func main() {

	// Create a partial test report
	report := v1alpha1.Report{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-report",
			Namespace: "test-namespace",
		},
	}

	// Create a test result
	result := v1alpha1.ReportResult{
		Description: "Test policy violation from manual test",
		Policy:      "test-policy",
		Rule:        "test-rule",
		Result:      "fail",
		Source:      "policy-reporter",
		Category:    "test-category",
		Severity:    "warning",
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
	client.Send(&openreports.ReportAdapter{Report: &report}, openreports.ResultAdapter{ReportResult: result})

	// Also test batch send
	results := []openreports.ResultAdapter{
		{
			ReportResult: v1alpha1.ReportResult{
				Description: "Batch test alert 1",
				Policy:      "batch-policy-1",
				Rule:        "batch-rule-1",
				Result:      "fail",
				Source:      "policy-reporter",
				Severity:    "warning",
			},
		},
		{
			ReportResult: v1alpha1.ReportResult{
				Description: "Batch test alert 2",
				Policy:      "batch-policy-2",
				Rule:        "batch-rule-2",
				Result:      "fail",
				Source:      "policy-reporter",
				Severity:    "warning",
			},
		},
	}

	fmt.Println("Sending batch test alerts")
	client.BatchSend(&openreports.ReportAdapter{Report: &report}, results)

	// Wait a moment to ensure alerts are sent before the program exits
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Test completed. Check AlertManager UI to verify alerts were received.")
}
