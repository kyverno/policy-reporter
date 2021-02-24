package kubernetes_test

import (
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/report"
)

var policyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "policy-report",
		"namespace":         "test",
		"creationTimestamp": "2021-02-23T15:00:00Z",
	},
	"summary": map[string]interface{}{
		"pass":  int64(1),
		"skip":  int64(2),
		"warn":  int64(3),
		"fail":  int64(4),
		"error": int64(5),
	},
	"results": []interface{}{
		map[string]interface{}{
			"message":  "message",
			"status":   "fail",
			"scored":   true,
			"policy":   "required-label",
			"rule":     "app-label-required",
			"category": "test",
			"severity": "low",
			"resources": []interface{}{
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Deployment",
					"name":       "nginx",
					"namespace":  "test",
					"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
		map[string]interface{}{
			"message":   "message 2",
			"status":    "fail",
			"scored":    true,
			"policy":    "priority-test",
			"resources": []interface{}{},
		},
	},
}

var minPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":      "policy-report",
		"namespace": "test",
	},
	"results": []interface{}{},
}

var clusterPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "clusterpolicy-report",
		"creationTimestamp": "2021-02-23T15:00:00Z",
	},
	"summary": map[string]interface{}{
		"pass":  int64(1),
		"skip":  int64(2),
		"warn":  int64(3),
		"fail":  int64(4),
		"error": int64(5),
	},
	"results": []interface{}{
		map[string]interface{}{
			"message":  "message",
			"status":   "fail",
			"scored":   true,
			"policy":   "required-label",
			"rule":     "app-label-required",
			"category": "test",
			"severity": "low",
			"resources": []interface{}{
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"name":       "policy-reporter",
					"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
	},
}

var minClusterPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name": "clusterpolicy-report",
	},
	"results": []interface{}{},
}

var priorityMap = map[string]string{
	"priority-test": "warning",
}

var mapper = kubernetes.NewMapper(priorityMap)

func Test_MapPolicyReport(t *testing.T) {
	preport := mapper.MapPolicyReport(policyMap)

	if preport.Name != "policy-report" {
		t.Errorf("Expected Name 'policy-report' (acutal %s)", preport.Name)
	}
	if preport.Namespace != "test" {
		t.Errorf("Expected Name 'test' (acutal %s)", preport.Namespace)
	}
	if preport.Summary.Pass != 1 {
		t.Errorf("Unexpected Summary.Pass value %d (expected 1)", preport.Summary.Pass)
	}
	if preport.Summary.Skip != 2 {
		t.Errorf("Unexpected Summary.Skip value %d (expected 2)", preport.Summary.Skip)
	}
	if preport.Summary.Warn != 3 {
		t.Errorf("Unexpected Summary.Warn value %d (expected 3)", preport.Summary.Warn)
	}
	if preport.Summary.Fail != 4 {
		t.Errorf("Unexpected Summary.Fail value %d (expected 4)", preport.Summary.Fail)
	}
	if preport.Summary.Error != 5 {
		t.Errorf("Unexpected Summary.Error value %d (expected 5)", preport.Summary.Error)
	}

	result1, ok := preport.Results["required-label__app-label-required__fail__dfd57c50-f30c-4729-b63f-b1954d8988d1"]
	if !ok {
		t.Error("Expected result not found")
	}

	if result1.Message != "message" {
		t.Errorf("Expected Message 'message' (acutal %s)", result1.Message)
	}
	if result1.Status != report.Fail {
		t.Errorf("Expected Message '%s' (acutal %s)", report.Fail, result1.Status)
	}
	if result1.Priority != report.ErrorPriority {
		t.Errorf("Expected Priority '%d' (acutal %d)", report.ErrorPriority, result1.Priority)
	}
	if !result1.Scored {
		t.Errorf("Expected Scored to be true")
	}
	if result1.Policy != "required-label" {
		t.Errorf("Expected Policy 'required-label' (acutal %s)", result1.Policy)
	}
	if result1.Rule != "app-label-required" {
		t.Errorf("Expected Rule 'app-label-required' (acutal %s)", result1.Rule)
	}
	if result1.Category != "test" {
		t.Errorf("Expected Category 'test' (acutal %s)", result1.Category)
	}
	if result1.Severity != report.Low {
		t.Errorf("Expected Severity '%s' (acutal %s)", report.Low, result1.Severity)
	}

	resource := result1.Resources[0]
	if resource.APIVersion != "v1" {
		t.Errorf("Expected Resource.APIVersion 'v1' (acutal %s)", resource.APIVersion)
	}
	if resource.Kind != "Deployment" {
		t.Errorf("Expected Resource.Kind 'Deployment' (acutal %s)", resource.Kind)
	}
	if resource.Name != "nginx" {
		t.Errorf("Expected Resource.Name 'nginx' (acutal %s)", resource.Name)
	}
	if resource.Namespace != "test" {
		t.Errorf("Expected Resource.Namespace 'test' (acutal %s)", resource.Namespace)
	}
	if resource.UID != "dfd57c50-f30c-4729-b63f-b1954d8988d1" {
		t.Errorf("Expected Resource.Namespace 'dfd57c50-f30c-4729-b63f-b1954d8988d1' (acutal %s)", resource.UID)
	}

	result2, ok := preport.Results["priority-test____fail"]
	if !ok {
		t.Error("Expected result not found")
	}

	if result2.Message != "message 2" {
		t.Errorf("Expected Message 'message' (acutal %s)", result1.Message)
	}
	if result2.Status != report.Fail {
		t.Errorf("Expected Message '%s' (acutal %s)", report.Fail, result2.Status)
	}
	if result2.Priority != report.WarningPriority {
		t.Errorf("Expected Priority '%d' (acutal %s)", report.WarningPriority, result2.Priority)
	}
	if !result2.Scored {
		t.Errorf("Expected Scored to be true")
	}
	if result2.Policy != "priority-test" {
		t.Errorf("Expected Policy 'priority-test' (acutal %s)", result2.Policy)
	}
	if result2.Rule != "" {
		t.Errorf("Expected Rule to be empty (acutal %s)", result2.Rule)
	}
	if result2.Category != "" {
		t.Errorf("Expected Category to be empty (acutal %s)", result2.Category)
	}
	if result2.Severity != "" {
		t.Errorf("Expected Severity to be empty (acutal %s)", report.Low)
	}
}

func Test_MapMinPolicyReport(t *testing.T) {
	report := mapper.MapPolicyReport(minPolicyMap)

	if report.Name != "policy-report" {
		t.Errorf("Expected Name 'policy-report' (acutal %s)", report.Name)
	}
	if report.Namespace != "test" {
		t.Errorf("Expected Name 'test' (acutal %s)", report.Namespace)
	}
	if report.Summary.Pass != 0 {
		t.Errorf("Unexpected Summary.Pass value %d (expected 0)", report.Summary.Pass)
	}
	if report.Summary.Skip != 0 {
		t.Errorf("Unexpected Summary.Skip value %d (expected 0)", report.Summary.Skip)
	}
	if report.Summary.Warn != 0 {
		t.Errorf("Unexpected Summary.Warn value %d (expected 0)", report.Summary.Warn)
	}
	if report.Summary.Fail != 0 {
		t.Errorf("Unexpected Summary.Fail value %d (expected 0)", report.Summary.Fail)
	}
	if report.Summary.Error != 0 {
		t.Errorf("Unexpected Summary.Error value %d (expected 0)", report.Summary.Error)
	}
}

func Test_MapClusterPolicyReport(t *testing.T) {
	report := mapper.MapClusterPolicyReport(clusterPolicyMap)

	if report.Name != "clusterpolicy-report" {
		t.Errorf("Expected Name 'clusterpolicy-report' (acutal %s)", report.Name)
	}
	if report.Summary.Pass != 1 {
		t.Errorf("Unexpected Summary.Pass value %d (expected 1)", report.Summary.Pass)
	}
	if report.Summary.Skip != 2 {
		t.Errorf("Unexpected Summary.Skip value %d (expected 2)", report.Summary.Skip)
	}
	if report.Summary.Warn != 3 {
		t.Errorf("Unexpected Summary.Warn value %d (expected 3)", report.Summary.Warn)
	}
	if report.Summary.Fail != 4 {
		t.Errorf("Unexpected Summary.Fail value %d (expected 4)", report.Summary.Fail)
	}
	if report.Summary.Error != 5 {
		t.Errorf("Unexpected Summary.Error value %d (expected 5)", report.Summary.Error)
	}
}

func Test_MapMinClusterPolicyReport(t *testing.T) {
	report := mapper.MapClusterPolicyReport(minClusterPolicyMap)

	if report.Name != "clusterpolicy-report" {
		t.Errorf("Expected Name 'clusterpolicy-report' (acutal %s)", report.Name)
	}
	if report.Summary.Pass != 0 {
		t.Errorf("Unexpected Summary.Pass value %d (expected 0)", report.Summary.Pass)
	}
	if report.Summary.Skip != 0 {
		t.Errorf("Unexpected Summary.Skip value %d (expected 0)", report.Summary.Skip)
	}
	if report.Summary.Warn != 0 {
		t.Errorf("Unexpected Summary.Warn value %d (expected 0)", report.Summary.Warn)
	}
	if report.Summary.Fail != 0 {
		t.Errorf("Unexpected Summary.Fail value %d (expected 0)", report.Summary.Fail)
	}
	if report.Summary.Error != 0 {
		t.Errorf("Unexpected Summary.Error value %d (expected 0)", report.Summary.Error)
	}
}

func Test_MapperSetPriorityMap(t *testing.T) {
	mapper := kubernetes.NewMapper(make(map[string]string))
	mapper.SetPriorityMap(map[string]string{"required-label": "debug"})

	preport := mapper.MapPolicyReport(policyMap)

	result1 := preport.Results["required-label__app-label-required__fail__dfd57c50-f30c-4729-b63f-b1954d8988d1"]

	if result1.Priority != report.DebugPriority {
		t.Errorf("Expected Policy '%d' (acutal %d)", report.DebugPriority, result1.Priority)
	}
}
