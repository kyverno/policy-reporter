package report

import "sync"

type PolicyReportStore struct {
	store map[string]PolicyReport
	rwm   *sync.RWMutex
}

func (s *PolicyReportStore) Get(id string) (PolicyReport, bool) {
	s.rwm.RLock()
	r, ok := s.store[id]
	s.rwm.RUnlock()

	return r, ok
}

func (s *PolicyReportStore) List() []PolicyReport {
	s.rwm.RLock()
	list := make([]PolicyReport, 0, len(s.store))

	for _, r := range s.store {
		list = append(list, r)
	}
	s.rwm.RUnlock()

	return list
}

func (s *PolicyReportStore) Add(r PolicyReport) {
	s.rwm.Lock()
	s.store[r.GetIdentifier()] = r
	s.rwm.Unlock()
}

func (s *PolicyReportStore) Remove(id string) {
	s.rwm.Lock()
	delete(s.store, id)
	s.rwm.Unlock()
}

func NewPolicyReportStore() *PolicyReportStore {
	return &PolicyReportStore{
		store: map[string]PolicyReport{},
		rwm:   new(sync.RWMutex),
	}
}

type ClusterPolicyReportStore struct {
	store map[string]ClusterPolicyReport
	rwm   *sync.RWMutex
}

func (s *ClusterPolicyReportStore) Get(id string) (ClusterPolicyReport, bool) {
	s.rwm.RLock()
	r, ok := s.store[id]
	s.rwm.RUnlock()

	return r, ok
}

func (s *ClusterPolicyReportStore) List() []ClusterPolicyReport {
	s.rwm.RLock()
	list := make([]ClusterPolicyReport, 0, len(s.store))

	for _, r := range s.store {
		list = append(list, r)
	}
	s.rwm.RUnlock()

	return list
}

func (s *ClusterPolicyReportStore) Add(r ClusterPolicyReport) {
	s.rwm.Lock()
	s.store[r.GetIdentifier()] = r
	s.rwm.Unlock()
}

func (s *ClusterPolicyReportStore) Remove(id string) {
	s.rwm.Lock()
	delete(s.store, id)
	s.rwm.Unlock()
}

func NewClusterPolicyReportStore() *ClusterPolicyReportStore {
	return &ClusterPolicyReportStore{
		store: map[string]ClusterPolicyReport{},
		rwm:   new(sync.RWMutex),
	}
}
