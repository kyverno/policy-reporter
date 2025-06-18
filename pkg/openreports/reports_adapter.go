package openreports

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
)

type ORReportAdapter struct {
	*v1alpha1.Report
	Results []ORResultAdapter
}

func (r *ORReportAdapter) GetResults() []ORResultAdapter {
	if len(r.Results) > 0 {
		return r.Results
	}
	ors := []ORResultAdapter{}
	for _, r := range r.Report.Results {
		ors = append(ors, ORResultAdapter{ReportResult: r})
	}
	r.Results = ors
	return ors
}

func (r *ORReportAdapter) HasResult(id string) bool {
	for _, r := range r.Report.Results {
		or := &ORResultAdapter{ReportResult: r}
		if or.GetID() == id {
			return true
		}
	}

	return false
}

func (r *ORReportAdapter) SetResults(results []ORResultAdapter) {
	r.Results = results
}

func (r *ORReportAdapter) GetSummary() v1alpha1.ReportSummary {
	return r.Summary
}

func (r *ORReportAdapter) GetSource() string {
	if len(r.Report.Results) == 0 {
		return ""
	}

	return r.Report.Results[0].Source
}

func (r *ORReportAdapter) GetKinds() []string {
	if r.GetScope() != nil {
		return []string{r.Scope.Kind}
	}

	list := make([]string, 0)
	for _, k := range r.Report.Results {
		or := &ORResultAdapter{ReportResult: k}
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

func (r *ORReportAdapter) GetSeverities() []string {
	list := make([]string, 0)
	for _, k := range r.Report.Results {
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
