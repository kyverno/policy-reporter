package openreports

import (
	"fmt"
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
	"slices"
)

type ORReportAdapter struct {
	v1alpha1.Report
}

type ORClusterReportAdapter struct {
	v1alpha1.ClusterReport
}

func (r *ORReportAdapter) GetResults() []v1alpha1.ReportResult {
	return r.Results
}

func (r *ORReportAdapter) HasResult(id string) bool {
	for _, r := range r.Results {
		if r.GetID() == id {
			return true
		}
	}

	return false
}

func (r *ORReportAdapter) SetResults(results []v1alpha1.ReportResult) {
	r.Results = results
}

func (r *ORReportAdapter) GetSummary() v1alpha1.ReportSummary {
	return r.Summary
}

func (r *ORReportAdapter) GetSource() string {
	if len(r.Results) == 0 {
		return ""
	}

	return r.Results[0].Source
}

func (r *ORReportAdapter) GetKinds() []string {
	if r.GetScope() != nil {
		return []string{r.Scope.Kind}
	}

	list := make([]string, 0)
	for _, k := range r.Results {
		if !k.HasResource() {
			continue
		}

		kind := k.GetResource().Kind

		if kind == "" || slices.Contains(list, kind) {
			continue
		}

		list = append(list, kind)
	}

	return list
}

func (r *ORReportAdapter) GetSeverities() []string {
	list := make([]string, 0)
	for _, k := range r.Results {

		if k.Severity == "" || slices.Contains(list, string(k.Severity)) {
			continue
		}

		list = append(list, string(k.Severity))
	}

	return list
}

func (r *ORReportAdapter) GetID() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.GetName())
	h1 = fnv1a.AddString64(h1, r.GetNamespace())

	return strconv.FormatUint(h1, 10)
}

func (r *ORReportAdapter) GetKey() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.Name)
}

func (r *ORReportAdapter) GetScope() *corev1.ObjectReference {
	return r.Scope
}

type ReportInterface interface {
	metav1.Object
	GetID() string
	GetKey() string
	GetScope() *corev1.ObjectReference
	GetResults() []v1alpha1.ReportResult
	HasResult(id string) bool
	GetSummary() v1alpha1.ReportSummary
	GetSource() string
	GetKinds() []string
	GetSeverities() []string
}
