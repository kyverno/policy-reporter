package result_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func TestDefaultGenerator(t *testing.T) {
	generator := result.NewIDGenerator(nil)

	t.Run("ID From Property", func(t *testing.T) {
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Properties: map[string]string{"resultID": "12345"}})

		if id != "12345" {
			t.Errorf("expected result id to be '12345', got :%s", id)
		}
	})

	t.Run("ID From Resource", func(t *testing.T) {
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Subjects: []corev1.ObjectReference{{Name: "test", Kind: "Pod"}}})

		if id != "18007334074686647077" {
			t.Errorf("expected result id to be '18007334074686647077', got :%s", id)
		}
	})

	t.Run("ID From Scope", func(t *testing.T) {
		id := generator.Generate(&v1alpha1.Report{Scope: &corev1.ObjectReference{Name: "test", Kind: "Pod"}}, v1alpha1.ReportResult{})

		if id != "18007334074686647077" {
			t.Errorf("expected result id to be '18007334074686647077', got :%s", id)
		}
	})
}

func TestCustomGenerator(t *testing.T) {
	t.Run("ID From Property", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"resource"})

		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Properties: map[string]string{"resultID": "12345"}})

		if id != "12345" {
			t.Errorf("expected result id to be '12345', got :%s", id)
		}
	})

	t.Run("ID From Resource", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"resource"})

		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Subjects: []corev1.ObjectReference{{Name: "test", Kind: "Pod"}}})

		if id != "18007334074686647077" {
			t.Errorf("expected result id to be '18007334074686647077', got :%s", id)
		}
	})

	t.Run("ID From Scope", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"resource"})

		id := generator.Generate(&v1alpha1.Report{Scope: &corev1.ObjectReference{Name: "test", Kind: "Pod"}}, v1alpha1.ReportResult{})

		if id != "18007334074686647077" {
			t.Errorf("expected result id to be '18007334074686647077', got :%s", id)
		}
	})

	t.Run("ID From Namespace", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"namespace"})

		empty := generator.Generate(&v1alpha1.Report{ObjectMeta: v1.ObjectMeta{Namespace: ""}}, v1alpha1.ReportResult{Description: ""})
		id := generator.Generate(&v1alpha1.Report{ObjectMeta: v1.ObjectMeta{Namespace: "test"}}, v1alpha1.ReportResult{Description: ""})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Policy", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"policy"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Policy: ""})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Policy: "test"})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Rule", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"rule"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Rule: ""})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Rule: "test"})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Result", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"result"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Result: ""})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Result: "fail"})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Category", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"category"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Category: ""})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Category: "test"})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Message", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"message"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Description: ""})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Description: "test"})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Created", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"created"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Timestamp: v1.Timestamp{Seconds: 0}})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Timestamp: v1.Timestamp{Seconds: 1714641964}})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Property", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"property:id"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{})
		id := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{Properties: map[string]string{"id": "1234"}})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Label", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"label:id"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{})
		id := generator.Generate(&v1alpha1.Report{ObjectMeta: v1.ObjectMeta{Labels: map[string]string{"id": "1234"}}}, v1alpha1.ReportResult{})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})

	t.Run("ID From Annotation", func(t *testing.T) {
		generator := result.NewIDGenerator([]string{"annotation:id"})

		empty := generator.Generate(&v1alpha1.Report{}, v1alpha1.ReportResult{})
		id := generator.Generate(&v1alpha1.Report{ObjectMeta: v1.ObjectMeta{Annotations: map[string]string{"id": "1234"}}}, v1alpha1.ReportResult{})

		if id == empty {
			t.Errorf("expected result id different from empty %s, got :%s", empty, id)
		}
	})
}
