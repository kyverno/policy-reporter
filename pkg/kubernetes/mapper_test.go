package kubernetes_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var mapper = kubernetes.NewMapper(priorityMap)

func Test_MapPolicyReport(t *testing.T) {
	preport := mapper.MapPolicyReport(policyReportCRD)

	if preport.Name != "policy-report" {
		t.Errorf("Expected Name 'policy-report' (acutal %s)", preport.Name)
	}
	if preport.Namespace != "test" {
		t.Errorf("Expected Namespace 'test' (acutal %s)", preport.Namespace)
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

	result1 := preport.GetResult(result1ID)
	if result1.ID == "" {
		t.Error("Expected result not found")
	}

	if result1.Message != "message" {
		t.Errorf("Expected Message 'message' (acutal %s)", result1.Message)
	}
	if result1.Status != report.Fail {
		t.Errorf("Expected Message '%s' (acutal %s)", report.Fail, result1.Status)
	}
	if result1.Priority != report.CriticalPriority {
		t.Errorf("Expected Priority '%d' (acutal %d)", report.CriticalPriority, result1.Priority)
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
	if result1.Source != "test" {
		t.Errorf("Expected Source 'test' (acutal %s)", result1.Source)
	}
	if result1.Severity != report.High {
		t.Errorf("Expected Severity '%s' (acutal %s)", report.High, result1.Severity)
	}
	if result1.Timestamp.Format("2006-01-02T15:04:05Z") != "2021-02-23T15:10:00Z" {
		t.Errorf("Expected Timestamp '2021-02-23T15:10:00Z' (acutal %s)", result1.Timestamp.Format("2006-01-02T15:04:05Z"))
	}
	if result1.Properties["version"] != "1.2.0" {
		t.Errorf("Expected Property '1.2.0' (acutal %s)", result1.Properties["version"])
	}

	resource := result1.Resource
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

	result2 := preport.GetResult(result2ID)
	if result1.ID == "" {
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

func Test_MapClusterPolicyReport(t *testing.T) {
	preport := mapper.MapClusterPolicyReport(clusterPolicyReportCRD)

	if preport.Name != "cluster-policy-report" {
		t.Errorf("Expected Name 'cluster-policy-report' (acutal %s)", preport.Name)
	}
	if preport.Summary.Skip != 0 {
		t.Errorf("Unexpected Summary.Skip value %d (expected 2)", preport.Summary.Skip)
	}
	if preport.Summary.Warn != 0 {
		t.Errorf("Unexpected Summary.Warn value %d (expected 3)", preport.Summary.Warn)
	}
	if preport.Summary.Fail != 4 {
		t.Errorf("Unexpected Summary.Fail value %d (expected 4)", preport.Summary.Fail)
	}
	if preport.Summary.Error != 0 {
		t.Errorf("Unexpected Summary.Error value %d (expected 5)", preport.Summary.Error)
	}

	result1 := preport.GetResult(cresult1ID)
	if result1.ID == "" {
		t.Error("Expected result not found")
	}

	if result1.Message != "message" {
		t.Errorf("Expected Message 'message' (acutal %s)", result1.Message)
	}
	if result1.Status != report.Fail {
		t.Errorf("Expected Message '%s' (acutal %s)", report.Fail, result1.Status)
	}
	if result1.Priority != report.CriticalPriority {
		t.Errorf("Expected Priority '%d' (acutal %d)", report.CriticalPriority, result1.Priority)
	}
	if !result1.Scored {
		t.Errorf("Expected Scored to be true")
	}
	if result1.Policy != "cluster-required-label" {
		t.Errorf("Expected Policy 'required-label' (acutal %s)", result1.Policy)
	}
	if result1.Rule != "ns-label-required" {
		t.Errorf("Expected Rule 'app-label-required' (acutal %s)", result1.Rule)
	}
	if result1.Category != "test" {
		t.Errorf("Expected Category 'test' (acutal %s)", result1.Category)
	}
	if result1.Source != "test" {
		t.Errorf("Expected Source 'test' (acutal %s)", result1.Source)
	}
	if result1.Severity != report.High {
		t.Errorf("Expected Severity '%s' (acutal %s)", report.High, result1.Severity)
	}
	if result1.Timestamp.Format("2006-01-02T15:04:05Z") != "2021-02-23T15:10:00Z" {
		t.Errorf("Expected Timestamp '2021-02-23T15:10:00Z' (acutal %s)", result1.Timestamp.Format("2006-01-02T15:04:05Z"))
	}

	resource := result1.Resource
	if resource.APIVersion != "v1" {
		t.Errorf("Expected Resource.APIVersion 'v1' (acutal %s)", resource.APIVersion)
	}
	if resource.Kind != "Namespace" {
		t.Errorf("Expected Resource.Kind 'Deployment' (acutal %s)", resource.Kind)
	}
	if resource.Name != "policy-reporter" {
		t.Errorf("Expected Resource.Name 'nginx' (acutal %s)", resource.Name)
	}
	if resource.UID != "dfd57c50-f30c-4729-b63f-b1954d8988d1" {
		t.Errorf("Expected Resource.Namespace 'dfd57c50-f30c-4729-b63f-b1954d8988d1' (acutal %s)", resource.UID)
	}
}

func Test_ResultIDPropertyMapping(t *testing.T) {
	mapper := kubernetes.NewMapper(map[string]string{})

	preport := mapper.MapPolicyReport(enforceReportCRD)

	result := preport.GetResult(result3ID)

	if result.ID == "" {
		t.Errorf("Expected ResultID was mapped from property Key %s", kubernetes.ResultIDKey)
	}
}

func Test_PriorityMap(t *testing.T) {
	t.Run("Test exact match, without default", func(t *testing.T) {
		mapper := kubernetes.NewMapper(map[string]string{"required-label": "debug"})

		preport := mapper.MapPolicyReport(policyReportCRD)

		result := preport.GetResult(result1ID)

		if result.Priority != report.DebugPriority {
			t.Errorf("Expected Policy '%d' (acutal %d)", report.DebugPriority, result.Priority)
		}
	})

	t.Run("Test exact match handled over default", func(t *testing.T) {
		mapper := kubernetes.NewMapper(map[string]string{"required-label": "debug", "default": "warning"})

		preport := mapper.MapPolicyReport(policyReportCRD)

		result := preport.GetResult(result1ID)
		if result.Priority != report.DebugPriority {
			t.Errorf("Expected Priority '%d' (acutal %d)", report.DebugPriority, result.Priority)
		}
	})

	t.Run("Test default expressions", func(t *testing.T) {
		mapper := kubernetes.NewMapper(map[string]string{"default": "warning"})

		preport := mapper.MapPolicyReport(policyReportCRD)

		result := preport.GetResult(result2ID)

		if result.Priority != report.WarningPriority {
			t.Errorf("Expected Policy '%d' (acutal %d)", report.WarningPriority, result.Priority)
		}
	})
}

func Test_MapCustomResultID(t *testing.T) {
	mapper := kubernetes.NewMapper(make(map[string]string))

	r := mapper.MapPolicyReport(enforceReportCRD)

	result := r.GetResult("123456")
	if result.ID == "" {
		t.Errorf("Expected resultID used as result.ID")
	}
}
