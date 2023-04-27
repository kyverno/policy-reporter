package report

import (
	"context"
	"errors"
	"sync"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type PolicyReportStore interface {
	// CreateSchemas for PolicyReports and PolicyReportResults
	CreateSchemas(context.Context) error
	// Get an PolicyReport by Type and ID
	Get(ctx context.Context, id string) (v1alpha2.ReportInterface, error)
	// Add a PolicyReport to the Store
	Add(ctx context.Context, r v1alpha2.ReportInterface) error
	// Update a PolicyReport to the Store
	Update(ctx context.Context, r v1alpha2.ReportInterface) error
	// Remove a PolicyReport with the given Type and ID from the Store
	Remove(ctx context.Context, id string) error
	// CleanUp removes all items in the store
	CleanUp(ctx context.Context) error
}

// PolicyReportStore caches the latest version of an PolicyReport
type policyReportStore struct {
	store map[string]map[string]v1alpha2.ReportInterface
	rwm   *sync.RWMutex
}

func (s *policyReportStore) CreateSchemas(_ context.Context) error {
	return nil
}

func (s *policyReportStore) Get(_ context.Context, id string) (v1alpha2.ReportInterface, error) {
	s.rwm.RLock()
	r, ok := s.store[PolicyReportType][id]
	s.rwm.RUnlock()
	if ok {
		return r, nil
	}

	s.rwm.RLock()
	r, ok = s.store[ClusterPolicyReportType][id]
	s.rwm.RUnlock()
	if ok {
		return r, nil
	}

	return nil, errors.New("report not found")
}

func (s *policyReportStore) Add(_ context.Context, r v1alpha2.ReportInterface) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.store[GetType(r)][r.GetID()] = r

	return nil
}

func (s *policyReportStore) Update(_ context.Context, r v1alpha2.ReportInterface) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.store[GetType(r)][r.GetID()] = r

	return nil
}

func (s *policyReportStore) Remove(ctx context.Context, id string) error {
	if r, err := s.Get(ctx, id); err == nil {
		s.rwm.Lock()
		defer s.rwm.Unlock()
		delete(s.store[GetType(r)], id)
	}

	return nil
}

func (s *policyReportStore) CleanUp(_ context.Context) error {
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
