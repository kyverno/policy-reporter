package kubernetes

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"
	"openreports.io/apis/openreports.io/v1alpha1"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	polrResource       = pr.SchemeGroupVersion.WithResource("policyreports")
	cpolrResource      = pr.SchemeGroupVersion.WithResource("clusterpolicyreports")
	openreportsReport  = v1alpha1.SchemeGroupVersion.WithResource("reports")
	openreportsCReport = v1alpha1.SchemeGroupVersion.WithResource("clusterreports")
)

const (
	wgpolicyAPIGroup    = "wgpolicyk8s.io/v1alpha2"
	openreportsAPIGroup = "openreports.io/v1alpha1"
)

type k8sPolicyReportClient struct {
	queue        *Queue
	metaClient   metadata.Interface
	synced       bool
	mx           *sync.Mutex
	reportFilter *report.MetaFilter
	stopChan     chan struct{}
}

func (k *k8sPolicyReportClient) HasSynced() bool {
	return k.synced
}

func (k *k8sPolicyReportClient) Stop() {
	close(k.stopChan)
}

func (k *k8sPolicyReportClient) Sync(stopper chan struct{}) error {
	polrFactory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)
	orFactory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)

	var (
		cpolrInformer cache.SharedIndexInformer
		orCInformer   cache.SharedIndexInformer
	)

	orInformer := k.configureORInformer(orFactory.ForResource(openreportsReport).Informer())
	polrInformer := k.configurePolrInformer(polrFactory.ForResource(polrResource).Informer())

	if !k.reportFilter.DisableClusterReports() {
		orCInformer = k.configureORInformer(orFactory.ForResource(openreportsCReport).Informer())
		cpolrInformer = k.configurePolrInformer(polrFactory.ForResource(cpolrResource).Informer())
	}

	polrFactory.Start(stopper)
	orFactory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, polrInformer.HasSynced) {
		return fmt.Errorf("failed to sync policy reports")
	}
	if !cache.WaitForCacheSync(stopper, orInformer.HasSynced) {
		return fmt.Errorf("failed to openreports reports")
	}

	if cpolrInformer != nil && !cache.WaitForCacheSync(stopper, cpolrInformer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policy reports")
	}
	if orCInformer != nil && !cache.WaitForCacheSync(stopper, orCInformer.HasSynced) {
		return fmt.Errorf("failed to sync openreports cluster reports")
	}

	k.synced = true

	zap.L().Info("policy report informer sync completed")

	return nil
}

func (k *k8sPolicyReportClient) Run(worker int, stopper chan struct{}) error {
	k.stopChan = stopper
	if err := k.Sync(stopper); err != nil {
		return err
	}

	k.queue.Run(worker, stopper)
	return nil
}

func (k *k8sPolicyReportClient) configurePolrInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					item.APIVersion = wgpolicyAPIGroup
					k.queue.Add(item)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					item.APIVersion = wgpolicyAPIGroup
					k.queue.Add(item)
				}
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			if item, ok := newObj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					item.APIVersion = wgpolicyAPIGroup
					k.queue.Add(item)
				}
			}
		},
	})

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return informer
}

func (k *k8sPolicyReportClient) configureORInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					item.APIVersion = openreportsAPIGroup
					k.queue.Add(item)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					item.APIVersion = openreportsAPIGroup
					k.queue.Add(item)
				}
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			if item, ok := newObj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					item.APIVersion = openreportsAPIGroup
					k.queue.Add(item)
				}
			}
		},
	})

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return informer
}

// NewPolicyReportClient new Client for Policy Report Kubernetes API
func NewPolicyReportClient(metaClient metadata.Interface, reportFilter *report.MetaFilter, queue *Queue) report.PolicyReportClient {
	return &k8sPolicyReportClient{
		metaClient:   metaClient,
		mx:           &sync.Mutex{},
		queue:        queue,
		reportFilter: reportFilter,
	}
}
