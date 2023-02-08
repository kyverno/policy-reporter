package kubernetes_test

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metafake "k8s.io/client-go/metadata/fake"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	v1alpha2client "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func NewFakeMetaClient() (*metafake.FakeMetadataClient, metafake.MetadataClient, metafake.MetadataClient) {
	schema := metafake.NewTestScheme()
	metav1.AddMetaToScheme(schema)

	client := metafake.NewSimpleMetadataClient(schema)
	return client, client.Resource(pr.SchemeGroupVersion.WithResource("policyreports")).Namespace("test").(metafake.MetadataClient), client.Resource(pr.SchemeGroupVersion.WithResource("clusterpolicyreports")).(metafake.MetadataClient)
}

func NewFakeClient() (*fake.Clientset, v1alpha2client.PolicyReportInterface, v1alpha2client.ClusterPolicyReportInterface) {
	client := fake.NewSimpleClientset()

	return client, client.Wgpolicyk8sV1alpha2().PolicyReports("test"), client.Wgpolicyk8sV1alpha2().ClusterPolicyReports()
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
