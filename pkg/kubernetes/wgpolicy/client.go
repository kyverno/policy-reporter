package wgpolicyclient

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	PolrResource  = pr.SchemeGroupVersion.WithResource("policyreports")
	CpolrResource = pr.SchemeGroupVersion.WithResource("clusterpolicyreports")
)

const (
	wgpolicyAPIGroup = "wgpolicyk8s.io/v1alpha2"
)

type wgpolicyReportClient struct {
	queue        *WGPolicyQueue
	metaClient   metadata.Interface
	synced       bool
	mx           *sync.Mutex
	reportFilter *report.MetaFilter
	stopChan     chan struct{}
}

func (k *wgpolicyReportClient) HasSynced() bool {
	return k.synced
}

func (k *wgpolicyReportClient) Stop() {
	close(k.stopChan)
}

func (k *wgpolicyReportClient) Sync(stopper chan struct{}) error {
	factory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)

	var cpolrInformer cache.SharedIndexInformer

	polrInformer := k.configureInformer(factory.ForResource(PolrResource).Informer())

	if !k.reportFilter.DisableClusterReports() {
		cpolrInformer = k.configureInformer(factory.ForResource(PolrResource).Informer())
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

func (k *wgpolicyReportClient) Run(worker int, stopper chan struct{}) error {
	k.stopChan = stopper
	if err := k.Sync(stopper); err != nil {
		return err
	}

	k.queue.Run(worker, stopper)
	return nil
}

func (k *wgpolicyReportClient) configureInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
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

// NewPolicyReportClient new Client for Policy Report Kubernetes API
func NewPolicyReportClient(metaClient metadata.Interface, reportFilter *report.MetaFilter, queue *WGPolicyQueue) report.PolicyReportClient {
	return &wgpolicyReportClient{
		metaClient:   metaClient,
		mx:           &sync.Mutex{},
		queue:        queue,
		reportFilter: reportFilter,
	}
}
