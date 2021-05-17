package report

import "sync"

type PolicyReportStore struct {
	store map[string]map[string]PolicyReport
	rwm   *sync.RWMutex
}

func (s *PolicyReportStore) Get(rType ReportType, id string) (PolicyReport, bool) {
	s.rwm.RLock()
	r, ok := s.store[rType][id]
	s.rwm.RUnlock()

	return r, ok
}

func (s *PolicyReportStore) List(rType ReportType) []PolicyReport {
	s.rwm.RLock()
	list := make([]PolicyReport, 0, len(s.store))

	for _, r := range s.store[rType] {
		list = append(list, r)
	}
	s.rwm.RUnlock()

	return list
}

func (s *PolicyReportStore) Add(r PolicyReport) {
	s.rwm.Lock()
	s.store[r.GetType()][r.GetIdentifier()] = r
	s.rwm.Unlock()
}

func (s *PolicyReportStore) Remove(rType ReportType, id string) {
	s.rwm.Lock()
	delete(s.store[rType], id)
	s.rwm.Unlock()
}

func NewPolicyReportStore() *PolicyReportStore {
	return &PolicyReportStore{
		store: map[ReportType]map[string]PolicyReport{
			PolicyReportType:        {},
			ClusterPolicyReportType: {},
		},
		rwm: new(sync.RWMutex),
	}
}
