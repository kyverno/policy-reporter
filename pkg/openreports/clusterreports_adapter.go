package openreports

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
)

type ClusterReportAdapter struct {
	*v1alpha1.ClusterReport
	Results []ResultAdapter
}

func (r *ClusterReportAdapter) GetResults() []ResultAdapter {
	if len(r.Results) > 0 {
		return r.Results
	}
	ors := []ResultAdapter{}
	for _, r := range r.ClusterReport.Results {
		ors = append(ors, ResultAdapter{ReportResult: r})
	}
	r.Results = ors
	return ors
}

func (r *ClusterReportAdapter) HasResult(id string) bool {
	for _, r := range r.ClusterReport.Results {
		or := &ResultAdapter{ReportResult: r}
		if or.GetID() == id {
			return true
		}
	}

	return false
}

func (r *ClusterReportAdapter) SetResults(results []ResultAdapter) {
	r.Results = results
}

func (r *ClusterReportAdapter) GetSummary() v1alpha1.ReportSummary {
	return r.Summary
}

func (r *ClusterReportAdapter) GetSource() string {
	if r.ClusterReport.Source == "" && len(r.GetResults()) > 0 {
		r.ClusterReport.Source = r.ClusterReport.Results[0].Source
	}

	return r.ClusterReport.Source
}

func (r *ClusterReportAdapter) GetKinds() []string {
	if r.GetScope() != nil {
		return []string{r.Scope.Kind}
	}

	list := make([]string, 0)
	for _, or := range r.GetResults() {
		if !or.HasResource() {
			continue
		}

		kind := or.GetResource().Kind

		if kind == "" || slices.Contains(list, kind) {
			continue
		}

		list = append(list, kind)
	}

	return list
}

func (r *ClusterReportAdapter) GetSeverities() []string {
	list := make([]string, 0)
	for _, k := range r.GetResults() {

		if k.Severity == "" || slices.Contains(list, string(k.Severity)) {
			continue
		}

		list = append(list, string(k.Severity))
	}

	return list
}

func (r *ClusterReportAdapter) GetID() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.GetName())

	return strconv.FormatUint(h1, 10)
}

func (r *ClusterReportAdapter) GetKey() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.Name)
}

func (r *ClusterReportAdapter) GetScope() *corev1.ObjectReference {
	return r.Scope
}
