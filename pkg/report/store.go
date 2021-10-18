package report

import "sync"

// PolicyReportStore caches the latest version of an PolicyReport
type PolicyReportStore struct {
	store map[string]map[string]PolicyReport
	rwm   *sync.RWMutex
}

// Get an PolicyReport by Type and ID
func (s *PolicyReportStore) Get(rType Type, id string) (PolicyReport, bool) {
	s.rwm.RLock()
	r, ok := s.store[rType][id]
	s.rwm.RUnlock()

	return r, ok
}

// List all PolicyReports of the given Type
func (s *PolicyReportStore) List(rType Type) []PolicyReport {
	s.rwm.RLock()
	list := make([]PolicyReport, 0, len(s.store))

	for _, r := range s.store[rType] {
		list = append(list, r)
	}
	s.rwm.RUnlock()

	return list
}

// Add a PolicyReport to the Store
func (s *PolicyReportStore) Add(r PolicyReport) {
	s.rwm.Lock()
	s.store[r.GetType()][r.GetIdentifier()] = r
	s.rwm.Unlock()
}

// Remove a PolicyReport with the given Type and ID from the Store
func (s *PolicyReportStore) Remove(rType Type, id string) {
	s.rwm.Lock()
	delete(s.store[rType], id)
	s.rwm.Unlock()
}

// NewPolicyReportStore construct a PolicyReportStore
func NewPolicyReportStore() *PolicyReportStore {
	return &PolicyReportStore{
		store: map[Type]map[string]PolicyReport{
			PolicyReportType:        {},
			ClusterPolicyReportType: {},
		},
		rwm: new(sync.RWMutex),
	}
}
