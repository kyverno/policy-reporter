package v1alpha2_test

import (
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

func TestCommon(t *testing.T) {
	t.Run("Priority.String", func(t *testing.T) {
		if v1alpha2.DefaultPriority.String() != "" {
			t.Error("unexpected default priority mapping")
		}

		if v1alpha2.DebugPriority.String() != "debug" {
			t.Error("unexpected debug priority mapping")
		}

		if v1alpha2.InfoPriority.String() != "info" {
			t.Error("unexpected info mapping")
		}

		if v1alpha2.WarningPriority.String() != "warning" {
			t.Error("unexpected warning mapping")
		}

		if v1alpha2.ErrorPriority.String() != "error" {
			t.Error("unexpected error mapping")
		}

		if v1alpha2.CriticalPriority.String() != "critical" {
			t.Error("unexpected critical mapping")
		}
	})

	t.Run("Priority.MarshalJSON", func(t *testing.T) {
		v, err := json.Marshal(v1alpha2.WarningPriority)
		if err != nil {
			t.Fatalf("unexpected marshal error: %s", err.Error())
		}

		if string(v) != `"warning"` {
			t.Fatalf("unexpected marshal value: %s", v)
		}
	})

	t.Run("NewPriority", func(t *testing.T) {
		if v1alpha2.NewPriority("") != v1alpha2.DefaultPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.NewPriority("debug") != v1alpha2.DebugPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.NewPriority("info") != v1alpha2.InfoPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.NewPriority("warning") != v1alpha2.WarningPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.NewPriority("error") != v1alpha2.ErrorPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.NewPriority("critical") != v1alpha2.CriticalPriority {
			t.Error("unexpected prioriry created")
		}
	})

	t.Run("PriorityFromSeverity", func(t *testing.T) {
		if v1alpha2.PriorityFromSeverity(v1alpha2.SeverityCritical) != v1alpha2.CriticalPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.PriorityFromSeverity(v1alpha2.SeverityHigh) != v1alpha2.ErrorPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.PriorityFromSeverity(v1alpha2.SeverityMedium) != v1alpha2.WarningPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.PriorityFromSeverity(v1alpha2.SeverityInfo) != v1alpha2.InfoPriority {
			t.Error("unexpected prioriry created")
		}

		if v1alpha2.PriorityFromSeverity(v1alpha2.SeverityLow) != v1alpha2.InfoPriority {
			t.Error("unexpected prioriry created")
		}
		if v1alpha2.PriorityFromSeverity("") != v1alpha2.DebugPriority {
			t.Error("unexpected prioriry created")
		}
	})
}

func TestPolicyReportResul(t *testing.T) {
	t.Run("GetResource Without Resources", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{}

		if r.GetResource() != nil {
			t.Error("expected nil resource for empty result")
		}
	})
	t.Run("GetResource With Resources", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test"}}}

		if r.GetResource().Name != "test" {
			t.Error("expected result resource returned")
		}
	})
	t.Run("GetKind Without Resource", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{}

		if r.GetKind() != "" {
			t.Error("expected result kind to be empty string")
		}
	})
	t.Run("GetKind", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test", Kind: "Pod"}}}

		if r.GetKind() != "Pod" {
			t.Error("expected result kind to be Pod")
		}
	})
	t.Run("GetID from Result With Resource", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test", Kind: "Pod"}}}

		if r.GetID() != "18007334074686647077" {
			t.Errorf("expected result kind to be '18007334074686647077', got :%s", r.GetID())
		}
	})
	t.Run("GetID from Result With ID Property", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test", Kind: "Pod"}}, Properties: map[string]string{"resultID": "result-id"}}

		if r.GetID() != "result-id" {
			t.Errorf("expected result kind to be 'result-id', got :%s", r.GetID())
		}
	})
	t.Run("GetID cached", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test", Kind: "Pod"}}, Properties: map[string]string{"resultID": "result-id"}}

		if r.GetID() != "result-id" {
			t.Errorf("expected result kind to be 'result-id', got :%s", r.GetID())
		}

		r.Properties["resultID"] = "test"

		if r.GetID() != "result-id" {
			t.Errorf("expected result ID doesn't change, got :%s", r.GetID())
		}
	})
	t.Run("ToResourceString with Namespace and Kind", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test", Namespace: "default", Kind: "Pod"}}}

		if r.ResourceString() != "default/pod/test" {
			t.Errorf("expected result resource string 'default/pod/name', got: %s", r.ResourceString())
		}
	})
	t.Run("ToResourceString with Kind", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test", Kind: "Namespace"}}}

		if r.ResourceString() != "namespace/test" {
			t.Errorf("expected result resource string 'namespace/test', got: %s", r.ResourceString())
		}
	})
	t.Run("ToResourceString with Name", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{{Name: "test"}}}

		if r.ResourceString() != "test" {
			t.Errorf("expected result resource string 'test', got :%s", r.ResourceString())
		}
	})
	t.Run("ToResourceString Without Resource", func(t *testing.T) {
		r := &v1alpha2.PolicyReportResult{}

		if r.ResourceString() != "" {
			t.Errorf("expected result resource string to be empty, got :%s", r.ResourceString())
		}
	})
}
