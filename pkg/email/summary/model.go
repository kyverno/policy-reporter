package summary

import (
	"sync"

	reportsv1alpha1 "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
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

func (s *Source) AddClusterSummary(rep openreports.ReportInterface) {
	s.ClusterScopeSummary.Skip += rep.GetSummary().Skip
	s.ClusterScopeSummary.Pass += rep.GetSummary().Pass
	s.ClusterScopeSummary.Warn += rep.GetSummary().Warn
	s.ClusterScopeSummary.Fail += rep.GetSummary().Fail
	s.ClusterScopeSummary.Error += rep.GetSummary().Error
}

func (s *Source) AddNamespacedSummary(ns string, sum reportsv1alpha1.ReportSummary) {
	s.mx.Lock()
	defer s.mx.Unlock()
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
