package report

import "sync"

type PolicyReportStore interface {
	// CreateSchemas for PolicyReports and PolicyReportResults
	CreateSchemas() error
	// Get an PolicyReport by Type and ID
	Get(id string) (*PolicyReport, bool)
	// Add a PolicyReport to the Store
	Add(r *PolicyReport) error
	// Add a PolicyReport to the Store
	Update(r *PolicyReport) error
	// Remove a PolicyReport with the given Type and ID from the Store
	Remove(id string) error
	// CleanUp removes all items in the store
	CleanUp() error
}

// PolicyReportStore caches the latest version of an PolicyReport
type policyReportStore struct {
	store map[string]map[string]*PolicyReport
	rwm   *sync.RWMutex
}

func (s *policyReportStore) CreateSchemas() error {
	return nil
}

func (s *policyReportStore) Get(id string) (*PolicyReport, bool) {
	s.rwm.RLock()
	r, ok := s.store[PolicyReportType][id]
	s.rwm.RUnlock()
	if ok {
		return r, ok
	}

	s.rwm.RLock()
	r, ok = s.store[ClusterPolicyReportType][id]
	s.rwm.RUnlock()

	return r, ok
}

func (s *policyReportStore) Add(r *PolicyReport) error {
	s.rwm.Lock()
	s.store[r.GetType()][r.GetIdentifier()] = r
	s.rwm.Unlock()

	return nil
}

func (s *policyReportStore) Update(r *PolicyReport) error {
	s.rwm.Lock()
	s.store[r.GetType()][r.GetIdentifier()] = r
	s.rwm.Unlock()

	return nil
}

func (s *policyReportStore) Remove(id string) error {
	if r, ok := s.Get(id); ok {
		s.rwm.Lock()
		delete(s.store[r.GetType()], id)
		s.rwm.Unlock()
	}

	return nil
}

func (s *policyReportStore) CleanUp() error {
	s.rwm.Lock()
	s.store = map[ResourceType]map[string]*PolicyReport{
		PolicyReportType:        {},
		ClusterPolicyReportType: {},
	}
	s.rwm.Unlock()

	return nil
}

// NewPolicyReportStore construct a PolicyReportStore
func NewPolicyReportStore() PolicyReportStore {
	return &policyReportStore{
		store: map[ResourceType]map[string]*PolicyReport{
			PolicyReportType:        {},
			ClusterPolicyReportType: {},
		},
		rwm: new(sync.RWMutex),
	}
}
