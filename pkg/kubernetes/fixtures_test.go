package kubernetes_test

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/report"

	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	v1alpha2client "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
)

func NewFakeCilent() (*fake.Clientset, v1alpha2client.PolicyReportInterface, v1alpha2client.ClusterPolicyReportInterface) {
	client := fake.NewSimpleClientset()

	return client, client.Wgpolicyk8sV1alpha2().PolicyReports("test"), client.Wgpolicyk8sV1alpha2().ClusterPolicyReports()
}

type store struct {
	store []report.LifecycleEvent
	rwm   *sync.RWMutex
}

func (s *store) Add(r report.LifecycleEvent) {
	s.rwm.Lock()
	s.store = append(s.store, r)
	s.rwm.Unlock()
}

func (s *store) Get(index int) report.LifecycleEvent {
	return s.store[index]
}

func (s *store) List() []report.LifecycleEvent {
	return s.store
}

func newStore(size int) *store {
	return &store{
		store: make([]report.LifecycleEvent, 0, size),
		rwm:   &sync.RWMutex{},
	}
}
