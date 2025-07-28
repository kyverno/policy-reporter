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

	prcache "github.com/kyverno/policy-reporter/pkg/cache"
	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	PolrResource  = pr.SchemeGroupVersion.WithResource("policyreports")
	CpolrResource = pr.SchemeGroupVersion.WithResource("clusterpolicyreports")
)

type wgpolicyReportClient struct {
	queue        *WGPolicyQueue
	metaClient   metadata.Interface
	synced       bool
	mx           *sync.Mutex
	reportFilter *report.MetaFilter
	stopChan     chan struct{}
	periodicSync bool
	syncInterval time.Duration
	cache        prcache.Cache
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
		cpolrInformer = k.configureInformer(factory.ForResource(CpolrResource).Informer())
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

	// Periodic sync if enabled - just stop the informer to trigger restart
	if k.periodicSync {
		zap.L().Info("policy report periodic sync enabled",
			zap.String("interval", k.syncInterval.String()))
		ticker := time.NewTicker(k.syncInterval)
		go func() {
			for {
				select {
				case <-ticker.C:
					zap.L().Info("triggering policy report sync - clearing cache and stopping informer")
					if k.cache != nil {
						k.cache.Clear()
						zap.L().Info("result cache cleared for periodic sync")
					}
					k.Stop()
					return
				case <-stopper:
					ticker.Stop()
					return
				}
			}
		}()
	} else {
		zap.L().Info("policy report periodic sync disabled")
	}

	k.queue.Run(worker, stopper)
	return nil
}

func (k *wgpolicyReportClient) configureInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
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
func NewPolicyReportClient(metaClient metadata.Interface, reportFilter *report.MetaFilter, queue *WGPolicyQueue, periodicSync bool, syncInterval time.Duration, cache prcache.Cache) report.PolicyReportClient {
	return &wgpolicyReportClient{
		metaClient:   metaClient,
		mx:           &sync.Mutex{},
		queue:        queue,
		reportFilter: reportFilter,
		periodicSync: periodicSync,
		syncInterval: syncInterval,
		cache:        cache,
	}
}
