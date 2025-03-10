package report_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

var controlled = true

type podClient struct {
	pod *corev1.Pod
	err error
}

func (c *podClient) Get(res *corev1.ObjectReference) (*corev1.Pod, error) {
	return c.pod, c.err
}

type jobClient struct {
	job *batchv1.Job
	err error
}

func (c jobClient) Get(res *corev1.ObjectReference) (*batchv1.Job, error) {
	return c.job, c.err
}

func TestSourceFilter(t *testing.T) {
	t.Run("include by namespace succeed", func(t *testing.T) {
		filter := report.NewSourceFilter(nil, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				Namespaces: validate.RuleSets{
					Include: []string{"test"},
				},
			},
		})

		result := filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		})

		assert.True(t, result)
	})

	t.Run("include by namespace fails", func(t *testing.T) {
		filter := report.NewSourceFilter(nil, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				Namespaces: validate.RuleSets{
					Include: []string{"default"},
				},
			},
		})

		result := filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		})

		assert.False(t, result)
	})

	t.Run("include by kind succeed", func(t *testing.T) {
		filter := report.NewSourceFilter(nil, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				Kinds: validate.RuleSets{
					Include: []string{"Pod"},
				},
			},
		})

		result := filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		})

		assert.True(t, result)
	})

	t.Run("include by kind fails", func(t *testing.T) {
		filter := report.NewSourceFilter(nil, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				Kinds: validate.RuleSets{
					Include: []string{"Job"},
				},
			},
		})

		result := filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		})

		assert.False(t, result)
	})

	t.Run("disable cluster reports", func(t *testing.T) {
		filter := report.NewSourceFilter(nil, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				DisableClusterReports: true,
			},
		})

		assert.False(t, filter.Validate(&v1alpha2.ClusterPolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: ""},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailNamespaceResult},
		}))

		assert.True(t, filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailNamespaceResult},
		}))
	})

	t.Run("include by kind succeed", func(t *testing.T) {
		filter := report.NewSourceFilter(nil, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				Kinds: validate.RuleSets{
					Include: []string{"Pod"},
				},
			},
		})

		result := filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		})

		assert.True(t, result)
	})

	t.Run("filter controlled pod", func(t *testing.T) {
		c := podClient{
			pod: &corev1.Pod{ObjectMeta: v1.ObjectMeta{Name: "nginx", Namespace: "test", OwnerReferences: []v1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "nginx-rs", Controller: &controlled},
			}}},
		}

		filter := report.NewSourceFilter(&c, nil, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				UncontrolledOnly: true,
			},
		})

		assert.False(t, filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		}))

		c.pod = &corev1.Pod{ObjectMeta: v1.ObjectMeta{Name: "nginx", Namespace: "test"}}

		assert.True(t, filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Pod", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		}))
	})

	t.Run("filter controlled job", func(t *testing.T) {
		c := jobClient{
			job: &batchv1.Job{ObjectMeta: v1.ObjectMeta{Name: "nginx", Namespace: "test", OwnerReferences: []v1.OwnerReference{
				{APIVersion: "batch/v1", Kind: "CronJob", Name: "nginx-rs", Controller: &controlled},
			}}},
		}

		filter := report.NewSourceFilter(nil, &c, []report.SourceValidation{
			{
				Selector: report.ReportSelector{
					Source: "kyverno",
				},
				UncontrolledOnly: true,
			},
		})

		assert.False(t, filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Job", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		}))

		c.job = &batchv1.Job{ObjectMeta: v1.ObjectMeta{Name: "nginx", Namespace: "test"}}

		assert.True(t, filter.Validate(&v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "polr", Namespace: "test"},
			Scope:      &corev1.ObjectReference{APIVersion: "v1", Kind: "Job", Name: "nginx", Namespace: "test"},
			Results:    []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
		}))
	})
}
