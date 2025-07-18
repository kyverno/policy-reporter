package orclient

import (
	"sync"

	pr "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/openreports/reports-api/pkg/client/clientset/versioned/fake"
	"github.com/openreports/reports-api/pkg/client/clientset/versioned/typed/openreports.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metafake "k8s.io/client-go/metadata/fake"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func NewFakeMetaClient() (*metafake.FakeMetadataClient, metafake.MetadataClient, metafake.MetadataClient) {
	schema := metafake.NewTestScheme()
	metav1.AddMetaToScheme(schema)

	client := metafake.NewSimpleMetadataClient(schema)
	return client, client.Resource(pr.SchemeGroupVersion.WithResource("reports")).Namespace("test").(metafake.MetadataClient), client.Resource(pr.SchemeGroupVersion.WithResource("clusterreports")).(metafake.MetadataClient)
}

func NewFakeClient() (*fake.Clientset, v1alpha1.ReportInterface, v1alpha1.ClusterReportInterface) {
	client := fake.NewSimpleClientset()

	return client, client.OpenreportsV1alpha1().Reports("test"), client.OpenreportsV1alpha1().ClusterReports()
}

type store struct {
	store []report.LifecycleEvent
	rwm   *sync.RWMutex
}

func (s *store) Add(r report.LifecycleEvent) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.store = append(s.store, r)
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
