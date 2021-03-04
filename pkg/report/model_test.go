package report_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
)

var result1 = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "resources",
	Scored:   true,
	Resources: []report.Resource{
		{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "test",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
		},
	},
}

var result2 = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "resources",
	Scored:   true,
	Resources: []report.Resource{
		{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "test",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
		},
	},
}

var preport = report.PolicyReport{
	Name:              "polr-test",
	Namespace:         "test",
	Results:           make(map[string]report.Result, 0),
	Summary:           report.Summary{},
	CreationTimestamp: time.Now(),
}

var creport = report.ClusterPolicyReport{
	Name:              "cpolr-test",
	Results:           make(map[string]report.Result, 0),
	Summary:           report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_PolicyReport(t *testing.T) {
	t.Run("Check PolicyReport.GetIdentifier", func(t *testing.T) {
		expected := fmt.Sprintf("%s__%s", preport.Namespace, preport.Name)

		if preport.GetIdentifier() != expected {
			t.Errorf("Expected PolicyReport.GetIdentifier() to be %s (actual: %s)", expected, preport.GetIdentifier())
		}
	})

	t.Run("Check PolicyReport.GetNewResults", func(t *testing.T) {
		preport1 := preport
		preport2 := preport

		preport1.Results = map[string]report.Result{result1.GetIdentifier(): result1}
		preport2.Results = map[string]report.Result{result1.GetIdentifier(): result1, result2.GetIdentifier(): result2}

		diff := preport2.GetNewResults(preport1)
		if len(diff) != 1 {
			t.Error("Expected 1 new result in diff")
		}
	})
}

func Test_ClusterPolicyReport(t *testing.T) {
	t.Run("Check ClusterPolicyReport.GetIdentifier", func(t *testing.T) {
		if creport.GetIdentifier() != creport.Name {
			t.Errorf("Expected ClusterPolicyReport.GetIdentifier() to be %s (actual: %s)", creport.Name, creport.GetIdentifier())
		}
	})

	t.Run("Check ClusterPolicyReport.GetNewResults", func(t *testing.T) {
		creport1 := creport
		creport2 := creport

		creport1.Results = map[string]report.Result{result1.GetIdentifier(): result1}
		creport2.Results = map[string]report.Result{result1.GetIdentifier(): result1, result2.GetIdentifier(): result2}

		diff := creport2.GetNewResults(creport1)
		if len(diff) != 1 {
			t.Error("Expected 1 new result in diff")
		}
	})
}

func Test_Result(t *testing.T) {
	t.Run("Check Result.GetIdentifier", func(t *testing.T) {
		expected := fmt.Sprintf("%s__%s__%s__%s", result1.Policy, result1.Rule, result1.Status, result1.Resources[0].UID)

		if result1.GetIdentifier() != expected {
			t.Errorf("Expected ClusterPolicyReport.GetIdentifier() to be %s (actual: %s)", expected, creport.GetIdentifier())
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
		if prio := report.Priority(0).String(); prio != "" {
			t.Errorf("Expected Priority to be '' (actual %s)", prio)
		}
		if prio := report.Priority(1).String(); prio != "debug" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(2).String(); prio != "info" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(3).String(); prio != "warning" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
		if prio := report.Priority(4).String(); prio != "error" {
			t.Errorf("Expected Priority to be debug (actual %s)", prio)
		}
	})
	t.Run("PriorityFromStatus", func(t *testing.T) {
		if prio := report.PriorityFromStatus(report.Fail); prio != report.ErrorPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.ErrorPriority, prio)
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
	t.Run("PriorityFromStatus", func(t *testing.T) {
		if prio := report.NewPriority(""); prio != report.DefaultPriority {
			t.Errorf("Expected Priority to be %d (actual %d)", report.DefaultPriority, prio)
		}
		if prio := report.NewPriority("error"); prio != report.ErrorPriority {
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
}
