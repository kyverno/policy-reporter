package openreports

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
)

func (r *ORClusterReportAdapter) GetResults() []v1alpha1.ReportResult {
	return r.Results
}

func (r *ORClusterReportAdapter) HasResult(id string) bool {
	for _, r := range r.Results {
		if r.GetID() == id {
			return true
		}
	}

	return false
}

func (r *ORClusterReportAdapter) SetResults(results []v1alpha1.ReportResult) {
	r.Results = results
}

func (r *ORClusterReportAdapter) GetSummary() v1alpha1.ReportSummary {
	return r.Summary
}

func (r *ORClusterReportAdapter) GetSource() string {
	if len(r.Results) == 0 {
		return ""
	}

	return r.Results[0].Source
}

func (r *ORClusterReportAdapter) GetKinds() []string {
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

func (r *ORClusterReportAdapter) GetSeverities() []string {
	list := make([]string, 0)
	for _, k := range r.Results {

		if k.Severity == "" || slices.Contains(list, string(k.Severity)) {
			continue
		}

		list = append(list, string(k.Severity))
	}

	return list
}

func (r *ORClusterReportAdapter) GetID() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.GetName())

	return strconv.FormatUint(h1, 10)
}

func (r *ORClusterReportAdapter) GetKey() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.Name)
}

func (r *ORClusterReportAdapter) GetScope() *corev1.ObjectReference {
	return r.Scope
}
