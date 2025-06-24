package violations

import (
	"sync"

	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type Result struct {
	Policy string
	Rule   string
	Kind   string
	Name   string
	Status string
}

func mapResult(polr openreports.ReportInterface, res v1alpha1.ReportResult) []Result {
	count := len(res.Subjects)
	rule := res.Rule
	if rule == "" {
		rule = res.Description
	}

	if count == 0 && polr.GetScope() == nil {
		return []Result{{
			Policy: res.Policy,
			Rule:   rule,
			Status: string(res.Result),
		}}
	} else if count == 0 && polr.GetScope() != nil {
		return []Result{{
			Policy: res.Policy,
			Rule:   rule,
			Name:   polr.GetScope().Name,
			Kind:   polr.GetScope().Kind,
			Status: string(res.Result),
		}}
	}

	list := make([]Result, 0, count)
	for _, re := range res.Subjects {
		list = append(list, Result{
			Policy: res.Policy,
			Rule:   rule,
			Name:   re.Name,
			Kind:   re.Kind,
			Status: string(res.Result),
		})
	}

	return list
}

type Source struct {
	Name             string
	ClusterPassed    int
	NamespacePassed  map[string]int
	ClusterResults   map[string][]Result
	NamespaceResults map[string]map[string][]Result
	ClusterReports   bool

	passMX *sync.Mutex
	crMX   *sync.Mutex
	nrMX   *sync.Mutex
}

func (s *Source) AddClusterResults(result []Result) {
	s.crMX.Lock()
	defer s.crMX.Unlock()
	s.ClusterResults[result[0].Status] = append(s.ClusterResults[result[0].Status], result...)
}

func (s *Source) AddClusterPassed(results int) {
	s.ClusterPassed += results
}

func (s *Source) AddNamespacedPassed(ns string, results int) {
	s.passMX.Lock()
	defer s.passMX.Unlock()
	s.NamespacePassed[ns] += results
}

func (s *Source) AddNamespacedResults(ns string, result []Result) {
	s.nrMX.Lock()
	defer s.nrMX.Unlock()
	if nr, ok := s.NamespaceResults[ns]; ok {
		s.NamespaceResults[ns][result[0].Status] = append(nr[result[0].Status], result...)
	} else {
		s.NamespaceResults[ns] = map[string][]Result{
			v1alpha2.StatusWarn:  make([]Result, 0),
			v1alpha2.StatusFail:  make([]Result, 0),
			v1alpha2.StatusError: make([]Result, 0),
		}

		s.NamespaceResults[ns][result[0].Status] = result
	}
}

func (s Source) InitResults(ns string) {
	s.nrMX.Lock()
	defer s.nrMX.Unlock()
	if _, ok := s.NamespaceResults[ns]; !ok {
		s.NamespaceResults[ns] = map[string][]Result{
			v1alpha2.StatusWarn:  make([]Result, 0),
			v1alpha2.StatusFail:  make([]Result, 0),
			v1alpha2.StatusError: make([]Result, 0),
		}
	}
}

func NewSource(name string, clusterReports bool) *Source {
	return &Source{
		Name:           name,
		ClusterReports: clusterReports,
		ClusterResults: map[string][]Result{
			v1alpha2.StatusWarn:  make([]Result, 0),
			v1alpha2.StatusFail:  make([]Result, 0),
			v1alpha2.StatusError: make([]Result, 0),
		},
		NamespaceResults: map[string]map[string][]Result{},
		NamespacePassed:  map[string]int{},
		passMX:           new(sync.Mutex),
		crMX:             new(sync.Mutex),
		nrMX:             new(sync.Mutex),
	}
}
