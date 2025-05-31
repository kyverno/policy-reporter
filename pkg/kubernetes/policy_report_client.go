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

	"github.com/kyverno/policy-reporter/pkg/report"
	pr "openreports.io/apis/openreports.io/v1alpha1"
)

var (
	polrResource  = pr.SchemeGroupVersion.WithResource("policyreports")
	cpolrResource = pr.SchemeGroupVersion.WithResource("clusterpolicyreports")
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
	factory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)

	var cpolrInformer cache.SharedIndexInformer

	polrInformer := k.configureInformer(factory.ForResource(polrResource).Informer())

	if !k.reportFilter.DisableClusterReports() {
		cpolrInformer = k.configureInformer(factory.ForResource(cpolrResource).Informer())
	}

	factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, polrInformer.HasSynced) {
		return fmt.Errorf("failed to sync policy reports")
	}

	if cpolrInformer != nil && !cache.WaitForCacheSync(stopper, cpolrInformer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policy reports")
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

func (k *k8sPolicyReportClient) configureInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					k.queue.Add(item)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					k.queue.Add(item)
				}
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			if item, ok := newObj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
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
