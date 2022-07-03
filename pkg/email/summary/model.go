package summary

import (
	"sync"

	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
)

type Summary struct {
	Skip  int
	Pass  int
	Warn  int
	Fail  int
	Error int
}

type Source struct {
	Name                  string
	ClusterScopeSummary   *Summary
	NamespaceScopeSummary map[string]*Summary
	ClusterReports        bool

	mx *sync.Mutex
}

func (s *Source) AddClusterSummary(sum v1alpha2.PolicyReportSummary) {
	s.ClusterScopeSummary.Skip += sum.Skip
	s.ClusterScopeSummary.Pass += sum.Pass
	s.ClusterScopeSummary.Warn += sum.Warn
	s.ClusterScopeSummary.Fail += sum.Fail
	s.ClusterScopeSummary.Error += sum.Error
}

func (s *Source) AddNamespacedSummary(ns string, sum v1alpha2.PolicyReportSummary) {
	s.mx.Lock()
	if d, ok := s.NamespaceScopeSummary[ns]; ok {
		d.Skip += sum.Skip
		d.Pass += sum.Pass
		d.Warn += sum.Warn
		d.Fail += sum.Fail
		d.Error += sum.Error
	} else {
		s.NamespaceScopeSummary[ns] = &Summary{
			Skip:  sum.Skip,
			Pass:  sum.Pass,
			Fail:  sum.Fail,
			Warn:  sum.Warn,
			Error: sum.Error,
		}
	}
	s.mx.Unlock()
}

func NewSource(name string, clusterReports bool) *Source {
	return &Source{
		Name:                  name,
		ClusterScopeSummary:   &Summary{},
		NamespaceScopeSummary: map[string]*Summary{},
		ClusterReports:        clusterReports,
		mx:                    new(sync.Mutex),
	}
}
