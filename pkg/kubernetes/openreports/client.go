package orclient

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

	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	OpenreportsReport  = v1alpha1.SchemeGroupVersion.WithResource("reports")
	OpenreportsCReport = v1alpha1.SchemeGroupVersion.WithResource("clusterreports")
)

const (
	openreportsAPIGroup = "openreports.io/v1alpha1"
)

type openreportsClient struct {
	queue        *ORQueue
	metaClient   metadata.Interface
	synced       bool
	mx           *sync.Mutex
	reportFilter *report.MetaFilter
	stopChan     chan struct{}
}

func (k *openreportsClient) HasSynced() bool {
	return k.synced
}

func (k *openreportsClient) Stop() {
	close(k.stopChan)
}

func (k *openreportsClient) Sync(stopper chan struct{}) error {
	factory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)

	var orCInformer cache.SharedIndexInformer

	orInformer := k.configureInformer(factory.ForResource(OpenreportsReport).Informer())

	if !k.reportFilter.DisableClusterReports() {
		orCInformer = k.configureInformer(factory.ForResource(OpenreportsCReport).Informer())
	}

	factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, orInformer.HasSynced) {
		return fmt.Errorf("failed to sync policy reports")
	}
	if !cache.WaitForCacheSync(stopper, orInformer.HasSynced) {
		return fmt.Errorf("failed to sync openreports reports")
	}

	if orCInformer != nil && !cache.WaitForCacheSync(stopper, orCInformer.HasSynced) {
		return fmt.Errorf("failed to sync openreports cluster reports")
	}

	k.synced = true

	zap.L().Info("openreports informer sync completed")

	return nil
}

func (k *openreportsClient) Run(worker int, stopper chan struct{}) error {
	k.stopChan = stopper
	if err := k.Sync(stopper); err != nil {
		return err
	}

	k.queue.Run(worker, stopper)
	return nil
}

func (k *openreportsClient) configureInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
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
func NewOpenreportsClient(metaClient metadata.Interface, reportFilter *report.MetaFilter, queue *ORQueue) report.PolicyReportClient {
	return &openreportsClient{
		metaClient:   metaClient,
		mx:           &sync.Mutex{},
		queue:        queue,
		reportFilter: reportFilter,
	}
}
