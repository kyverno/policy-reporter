package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	polrResource  = pr.SchemeGroupVersion.WithResource("policyreports")
	cpolrResource = pr.SchemeGroupVersion.WithResource("clusterpolicyreports")
)

type k8sPolicyReportClient struct {
	queue        *Queue
	fatcory      metadatainformer.SharedInformerFactory
	polr         informers.GenericInformer
	cpolr        informers.GenericInformer
	metaClient   metadata.Interface
	synced       bool
	mx           *sync.Mutex
	reportFilter *report.Filter
}

func (k *k8sPolicyReportClient) HasSynced() bool {
	return k.synced
}

func (k *k8sPolicyReportClient) Sync(stopper chan struct{}) error {
	var cpolrInformer cache.SharedIndexInformer

	polrInformer := k.configureInformer(k.polr.Informer())

	if !k.reportFilter.DisableClusterReports() {
		cpolrInformer = k.configureInformer(k.cpolr.Informer())
	}

	k.fatcory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, polrInformer.HasSynced) {
		return fmt.Errorf("failed to sync policy reports")
	}

	if cpolrInformer != nil && !cache.WaitForCacheSync(stopper, cpolrInformer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policy reports")
	}

	k.synced = true

	zap.L().Info("informer sync completed")

	return nil
}

func (k *k8sPolicyReportClient) Run(worker int, stopper chan struct{}) error {
	if err := k.Sync(stopper); err != nil {
		return err
	}

	k.queue.Run(worker, stopper)

	return nil
}

func (k *k8sPolicyReportClient) RefreshPolicyReports(ctx context.Context) error {
	g := &errgroup.Group{}
	g.Go(func() error {
		return k.refresh(ctx, polrResource)
	})

	g.Go(func() error {
		return k.refresh(ctx, cpolrResource)
	})

	return g.Wait()
}

func (k *k8sPolicyReportClient) refresh(ctx context.Context, resource schema.GroupVersionResource) error {
	var limit int64 = 25
	var cont string

	for {
		list, err := k.metaClient.Resource(resource).List(ctx, v1.ListOptions{
			Limit:    limit,
			Continue: cont,
		})
		if err != nil {
			return err
		}

		for _, report := range list.Items {
			k.queue.Add(&report)
		}

		if list.Continue == "" {
			return nil
		}

		cont = list.Continue
		time.Sleep(100 * time.Millisecond)
	}
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
func NewPolicyReportClient(metaClient metadata.Interface, reportFilter *report.Filter, queue *Queue) report.PolicyReportClient {
	fatcory := metadatainformer.NewSharedInformerFactory(metaClient, 15*time.Minute)
	polr := fatcory.ForResource(polrResource)
	cpolr := fatcory.ForResource(cpolrResource)

	return &k8sPolicyReportClient{
		metaClient:   metaClient,
		fatcory:      fatcory,
		polr:         polr,
		cpolr:        cpolr,
		mx:           &sync.Mutex{},
		queue:        queue,
		reportFilter: reportFilter,
	}
}
