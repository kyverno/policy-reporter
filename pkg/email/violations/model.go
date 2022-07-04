package violations

import (
	"sync"

	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
)

type Result struct {
	Policy string
	Rule   string
	Kind   string
	Name   string
	Status string
}

func mapResult(res v1alpha2.PolicyReportResult) []Result {
	count := len(res.Resources)
	rule := res.Rule
	if rule == "" {
		rule = res.Message
	}

	if count == 0 {
		return []Result{{
			Policy: res.Policy,
			Rule:   rule,
			Status: string(res.Result),
		}}
	}

	list := make([]Result, 0, count)
	for _, re := range res.Resources {
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
	s.ClusterResults[result[0].Status] = append(s.ClusterResults[result[0].Status], result...)
	s.crMX.Unlock()
}

func (s *Source) AddClusterPassed(results int) {
	s.ClusterPassed += results
}

func (s *Source) AddNamespacedPassed(ns string, results int) {
	s.passMX.Lock()
	s.NamespacePassed[ns] += results
	s.passMX.Unlock()
}

func (s *Source) AddNamespacedResults(ns string, result []Result) {
	s.nrMX.Lock()
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
	s.nrMX.Unlock()
}

func (s Source) InitResults(ns string) {
	s.nrMX.Lock()
	if _, ok := s.NamespaceResults[ns]; !ok {
		s.NamespaceResults[ns] = map[string][]Result{
			v1alpha2.StatusWarn:  make([]Result, 0),
			v1alpha2.StatusFail:  make([]Result, 0),
			v1alpha2.StatusError: make([]Result, 0),
		}
	}
	s.nrMX.Unlock()
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
