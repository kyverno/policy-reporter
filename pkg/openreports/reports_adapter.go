package openreports

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
)

type ReportAdapter struct {
	*v1alpha1.Report
	Results []ResultAdapter
	Source  string
}

func (r *ReportAdapter) GetResults() []ResultAdapter {
	if len(r.Results) > 0 {
		return r.Results
	}
	ors := []ResultAdapter{}
	for _, r := range r.Report.Results {
		ors = append(ors, ResultAdapter{ReportResult: r})
	}
	r.Results = ors
	return ors
}

func (r *ReportAdapter) HasResult(id string) bool {
	for _, r := range r.GetResults() {
		if r.GetID() == id {
			return true
		}
	}

	return false
}

func (r *ReportAdapter) SetResults(results []ResultAdapter) {
	r.Results = results
}

func (r *ReportAdapter) GetSummary() v1alpha1.ReportSummary {
	return r.Summary
}

func (r *ReportAdapter) GetSource() string {
	if r.Report.Source != "" {
		return r.Report.Source
	}

	if len(r.GetResults()) == 0 {
		return ""
	}

	r.Report.Source = r.Report.Results[0].Source
	return r.Report.Source
}

func (r *ReportAdapter) GetKinds() []string {
	if r.GetScope() != nil {
		return []string{r.Scope.Kind}
	}

	list := make([]string, 0)
	for _, k := range r.Report.Results {
		or := &ResultAdapter{ReportResult: k}
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

func (r *ReportAdapter) GetSeverities() []string {
	list := make([]string, 0)
	for _, k := range r.Report.Results {
		if k.Severity == "" || slices.Contains(list, string(k.Severity)) {
			continue
		}

		list = append(list, string(k.Severity))
	}

	return list
}

func (r *ReportAdapter) GetID() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.GetName())
	h1 = fnv1a.AddString64(h1, r.GetNamespace())

	return strconv.FormatUint(h1, 10)
}

func (r *ReportAdapter) GetKey() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.Name)
}

func (r *ReportAdapter) GetScope() *corev1.ObjectReference {
	return r.Scope
}
