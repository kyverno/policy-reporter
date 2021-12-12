package report_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

var result1 = &report.Result{
	ID:       "e0659854c6ee5c1a4df9242b2eb8b40919967842",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "resources",
	Severity: report.High,
	Scored:   true,
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

var result2 = &report.Result{
	ID:       "2",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "resources",
	Scored:   true,
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	},
}

var preport = &report.PolicyReport{
	ID:                "24cfa233af033d104cd6ce0ff9a5a875c71a5844",
	Name:              "polr-test",
	Namespace:         "test",
	Results:           make(map[string]*report.Result),
	Summary:           &report.Summary{},
	CreationTimestamp: time.Now(),
}

var creport = &report.PolicyReport{
	ID:                "57e1551475e17740bacc3640d2412b1a6aad6a93",
	Name:              "cpolr-test",
	Results:           make(map[string]*report.Result),
	Summary:           &report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_PolicyReport(t *testing.T) {
	t.Run("Check PolicyReport.GetIdentifier", func(t *testing.T) {
		expected := report.GeneratePolicyReportID(preport.Name, preport.Namespace)

		if preport.GetIdentifier() != expected {
			t.Errorf("Expected PolicyReport.GetIdentifier() to be %s (actual: %s)", expected, preport.GetIdentifier())
		}
	})

	t.Run("Check PolicyReport.GetNewResults", func(t *testing.T) {
		preport1 := &report.PolicyReport{
			ID:                "24cfa233af033d104cd6ce0ff9a5a875c71a5844",
			Name:              "polr-test",
			Namespace:         "test",
			Summary:           &report.Summary{},
			CreationTimestamp: time.Now(),
			Results:           map[string]*report.Result{result1.GetIdentifier(): result1},
		}
		preport2 := &report.PolicyReport{
			ID:                "24cfa233af033d104cd6ce0ff9a5a875c71a5844",
			Name:              "polr-test",
			Namespace:         "test",
			Summary:           &report.Summary{},
			CreationTimestamp: time.Now(),
			Results:           map[string]*report.Result{result1.GetIdentifier(): result1, result2.GetIdentifier(): result2},
		}

		diff := preport2.GetNewResults(preport1)
		if len(diff) != 1 {
			t.Error("Expected 1 new result in diff")
		}
	})
}

func Test_ClusterPolicyReport(t *testing.T) {
	t.Run("Check ClusterPolicyReport.GetIdentifier", func(t *testing.T) {
		expected := report.GeneratePolicyReportID(creport.Name, creport.Namespace)

		if creport.GetIdentifier() != expected {
			t.Errorf("Expected ClusterPolicyReport.GetIdentifier() to be %s (actual: %s)", expected, creport.GetIdentifier())
		}
	})
	t.Run("Check ClusterPolicyReport.GetType", func(t *testing.T) {
		if creport.GetType() != report.ClusterPolicyReportType {
			t.Errorf("Expected ClusterPolicyReport.GetType() to be %s (actual: %s)", report.ClusterPolicyReportType, creport.GetType())
		}
	})

	t.Run("Check ClusterPolicyReport.GetNewResults", func(t *testing.T) {
		creport1 := &report.PolicyReport{
			ID:                "57e1551475e17740bacc3640d2412b1a6aad6a93",
			Name:              "cpolr-test",
			Summary:           &report.Summary{},
			CreationTimestamp: time.Now(),
			Results:           map[string]*report.Result{result1.GetIdentifier(): result1},
		}

		creport2 := &report.PolicyReport{
			ID:                "57e1551475e17740bacc3640d2412b1a6aad6a93",
			Name:              "cpolr-test",
			Summary:           &report.Summary{},
			CreationTimestamp: time.Now(),
			Results:           map[string]*report.Result{result1.GetIdentifier(): result1, result2.GetIdentifier(): result2},
		}

		diff := creport2.GetNewResults(creport1)
		if len(diff) != 1 {
			t.Error("Expected 1 new result in diff")
		}
	})
}

func Test_Result(t *testing.T) {
	t.Run("Check Result.GetIdentifier", func(t *testing.T) {
		expected := report.GeneratePolicyReportResultID(result1.Resource.UID, result1.Policy, result1.Rule, result1.Status, "")

		if result1.GetIdentifier() != expected {
			t.Errorf("Expected ClusterPolicyReport.GetIdentifier() to be %s (actual: %s)", expected, creport.GetIdentifier())
		}
	})
	t.Run("Check Result.HasResource with Resource", func(t *testing.T) {
		if result1.HasResource() == false {
			t.Errorf("Expected result1.HasResource() to be true (actual: %v)", result1.HasResource())
		}
	})
	t.Run("Check Result.HasResource without Resource", func(t *testing.T) {
		result := report.Result{}

		if result.HasResource() == true {
			t.Errorf("Expected result.HasResource() to be false without a Resource (actual: %v)", result1.HasResource())
		}
	})

}

func Test_MarshalPriority(t *testing.T) {
	priority := report.NewPriority("error")
	if result, _ := priority.MarshalJSON(); string(result) != `"error"` {
		t.Errorf("Unexpected Marshel Result: %s", result)
	}
}

func Test_Priorities(t *testing.T) {
	t.Run("Priority.String", func(t *testing.T) {
		if prio := report.Priority(report.DefaultPriority).String(); prio != "" {
			t.Errorf("Expected Priority to be '' (actual %s)", prio)
		}
		if prio := report.Priority(report.DebugPriority).String(); prio != "debug" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(report.InfoPriority).String(); prio != "info" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(report.WarningPriority).String(); prio != "warning" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(report.ErrorPriority).String(); prio != "error" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(report.CriticalPriority).String(); prio != "critical" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
	})
	t.Run("PriorityFromStatus", func(t *testing.T) {
		if prio := report.PriorityFromStatus(report.Fail); prio != report.CriticalPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.CriticalPriority, prio)
		}
		if prio := report.PriorityFromStatus(report.Error); prio != report.ErrorPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.ErrorPriority, prio)
		}
		if prio := report.PriorityFromStatus(report.Pass); prio != report.InfoPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.InfoPriority, prio)
		}
		if prio := report.PriorityFromStatus(report.Skip); prio != report.DefaultPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.DefaultPriority, prio)
		}
		if prio := report.PriorityFromStatus(report.Warn); prio != report.WarningPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.WarningPriority, prio)
		}
	})
	t.Run("NewPriority", func(t *testing.T) {
		if prio := report.NewPriority(""); prio != report.DefaultPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.DefaultPriority, prio)
		}
		if prio := report.NewPriority("error"); prio != report.ErrorPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.ErrorPriority, prio)
		}
		if prio := report.NewPriority("critical"); prio != report.CriticalPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.ErrorPriority, prio)
		}
		if prio := report.NewPriority("warning"); prio != report.WarningPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.WarningPriority, prio)
		}
		if prio := report.NewPriority("info"); prio != report.InfoPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.InfoPriority, prio)
		}
		if prio := report.NewPriority("debug"); prio != report.DebugPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.DebugPriority, prio)
		}
	})
	t.Run("PriorityFromSeverity", func(t *testing.T) {
		if prio := report.PriorityFromSeverity(report.High); prio != report.CriticalPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.CriticalPriority, prio)
		}
		if prio := report.PriorityFromSeverity(report.Medium); prio != report.WarningPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.WarningPriority, prio)
		}
		if prio := report.PriorityFromSeverity(report.Low); prio != report.InfoPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.InfoPriority, prio)
		}
	})
}
