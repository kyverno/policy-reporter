package report

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type PolicyReportStore interface {
	// CreateSchemas for PolicyReports and PolicyReportResults
	CreateSchemas() error
	// Get an PolicyReport by Type and ID
	Get(id string) (v1alpha2.ReportInterface, bool)
	// Add a PolicyReport to the Store
	Add(r v1alpha2.ReportInterface) error
	// Update a PolicyReport to the Store
	Update(r v1alpha2.ReportInterface) error
	// Remove a PolicyReport with the given Type and ID from the Store
	Remove(id string) error
	// CleanUp removes all items in the store
	CleanUp() error
}

// PolicyReportStore caches the latest version of an PolicyReport
type policyReportStore struct {
	store map[string]map[string]v1alpha2.ReportInterface
	rwm   *sync.RWMutex
}

func (s *policyReportStore) CreateSchemas() error {
	return nil
}

func (s *policyReportStore) Get(id string) (v1alpha2.ReportInterface, bool) {
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

func (s *policyReportStore) Add(r v1alpha2.ReportInterface) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.store[GetType(r)][r.GetID()] = r

	return nil
}

func (s *policyReportStore) Update(r v1alpha2.ReportInterface) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.store[GetType(r)][r.GetID()] = r

	return nil
}

func (s *policyReportStore) Remove(id string) error {
	if r, ok := s.Get(id); ok {
		s.rwm.Lock()
		defer s.rwm.Unlock()
		delete(s.store[GetType(r)], id)
	}

	return nil
}

func (s *policyReportStore) CleanUp() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.store = map[ResourceType]map[string]v1alpha2.ReportInterface{
		PolicyReportType:        {},
		ClusterPolicyReportType: {},
	}

	return nil
}

// NewPolicyReportStore construct a PolicyReportStore
func NewPolicyReportStore() PolicyReportStore {
	return &policyReportStore{
		store: map[ResourceType]map[string]v1alpha2.ReportInterface{
			PolicyReportType:        {},
			ClusterPolicyReportType: {},
		},
		rwm: new(sync.RWMutex),
	}
}
