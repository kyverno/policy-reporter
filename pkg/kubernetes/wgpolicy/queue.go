package wgpolicyclient

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	reportsv1alpha1 "openreports.io/apis/openreports.io/v1alpha1"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/policyreport/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

type WGPolicyQueue struct {
	queue         workqueue.TypedRateLimitingInterface[string]
	client        v1alpha2.Wgpolicyk8sV1alpha2Interface
	reconditioner *result.Reconditioner
	debouncer     kubernetes.Debouncer
	lock          *sync.Mutex
	cache         sets.Set[string]
	filter        *report.SourceFilter
}

func (q *WGPolicyQueue) Add(obj *v1.PartialObjectMetadata) error {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}

	q.queue.Add(key)
	return nil
}

func (q *WGPolicyQueue) Run(workers int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	for i := 0; i < workers; i++ {
		go wait.Until(q.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (q *WGPolicyQueue) runWorker() {
	for q.processNextItem() {
	}
}

func (q *WGPolicyQueue) processNextItem() bool {
	key, quit := q.queue.Get()
	if quit {
		return false
	}
	defer q.queue.Done(key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		q.queue.Forget(key)
		return true
	}

	var (
		rep   openreports.ReportInterface
		polr  *pr.PolicyReport
		cpolr *pr.ClusterPolicyReport
	)

	if namespace != "" {
		polr, err = q.client.PolicyReports(namespace).Get(context.Background(), name, v1.GetOptions{})
		rep = polr.ToOpenReports()
	} else {
		cpolr, err = q.client.ClusterPolicyReports().Get(context.Background(), name, v1.GetOptions{})
		rep = cpolr.ToOpenReports()
	}
	if errors.IsNotFound(err) {
		q.handleNotFoundReport(key)
		return true
	}

	if ok := q.filter.Validate(rep); !ok {
		return true
	}

	event := func() report.Event {
		q.lock.Lock()
		defer q.lock.Unlock()
		event := report.Added
		if q.cache.Has(key) {
			event = report.Updated
		} else {
			q.cache.Insert(key)
		}
		return event
	}()

	q.handleErr(err, key)

	q.debouncer.Add(report.LifecycleEvent{Type: event, PolicyReport: q.reconditioner.Prepare(rep)})

	return true
}

func (q *WGPolicyQueue) handleErr(err error, key string) {
	if err == nil {
		q.queue.Forget(key)
		return
	}

	if q.queue.NumRequeues(key) < 5 {
		zap.L().Error("process error", zap.Any("key", key), zap.Error(err))

		q.queue.AddRateLimited(key)
		return
	}

	q.queue.Forget(key)

	runtime.HandleError(err)
	zap.L().Warn("dropping report out of queue", zap.Any("key", key), zap.Error(err))
}

func (q *WGPolicyQueue) handleNotFoundReport(key string) {
	var rep openreports.ReportInterface
	namespace, name, _ := cache.SplitMetaNamespaceKey(key) // todo: check this

	if namespace == "" {
		rep = &reportsv1alpha1.ClusterReport{
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
		}
	} else {
		rep = &reportsv1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	func() {
		q.lock.Lock()
		defer q.lock.Unlock()
		q.cache.Delete(key)
	}()
	q.debouncer.Add(report.LifecycleEvent{Type: report.Deleted, PolicyReport: q.reconditioner.Prepare(rep)})
}

func NewWGPolicyQueue(
	debouncer kubernetes.Debouncer,
	queue workqueue.TypedRateLimitingInterface[string],
	client v1alpha2.Wgpolicyk8sV1alpha2Interface,
	filter *report.SourceFilter,
	reconditioner *result.Reconditioner,
) *WGPolicyQueue {
	return &WGPolicyQueue{
		debouncer:     debouncer,
		queue:         queue,
		client:        client,
		cache:         sets.New[string](),
		lock:          &sync.Mutex{},
		filter:        filter,
		reconditioner: reconditioner,
	}
}
