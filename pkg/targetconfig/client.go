package targetconfig

import (
	"context"

	reports "github.com/openreports/reports-api/pkg/client/clientset/versioned/typed/openreports.io/v1alpha1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	crds "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	informer "github.com/kyverno/policy-reporter/pkg/crd/client/informers/externalversions"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type Client struct {
	targetFactory  target.Factory
	collection     *target.Collection
	informer       cache.SharedIndexInformer
	orClient       reports.OpenreportsV1alpha1Interface
	wgpolicyClient v1alpha2.Wgpolicyk8sV1alpha2Interface
}

func (c *Client) ConfigureInformer() {
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			zap.L().Info("new target", zap.String("name", tc.Name))

			t, err := c.targetFactory.CreateSingleClient(tc)
			if err != nil {
				zap.L().Error("unable to create target from TargetConfig", zap.String("name", tc.Name), zap.Error(err))
				return
			} else if t == nil {
				zap.L().Error("provided TargetConfig is invalid", zap.String("name", tc.Name))
				return
			}

			c.collection.AddTarget(tc.Name, t)

			if !tc.Spec.SkipExisting {
				reports := []openreports.ReportInterface{}

				polrReports, err := c.WGPolicyReports(context.TODO())
				if err != nil {
					zap.L().Error("unable to get WGPolicy reports", zap.String("name", tc.Name), zap.Error(err))
				}

				reports = append(reports, polrReports...)

				openReports, err := c.OpenReports(context.TODO())
				if err != nil {
					zap.L().Error("unable to get OpenReports", zap.String("name", tc.Name), zap.Error(err))
				}

				reports = append(reports, openReports...)

				switch t.Client.Type() {
				case target.SingleSend:
					listener := listener.NewSendResultListener(target.NewCollection(t))
					for _, polr := range reports {
						for _, res := range polr.GetResults() {
							listener(polr, res, false)
						}
					}

				case target.BatchSend:
					listener := listener.NewSendScopeResultsListener(target.NewCollection(t))
					for _, polr := range reports {
						listener(polr, polr.GetResults(), false)
					}

				case target.SyncSend:
					listener := listener.NewSendSyncResultsListener(target.NewCollection(t))
					for _, polr := range reports {
						listener(polr)
					}
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			tc := newObj.(*v1alpha1.TargetConfig)
			zap.L().Info("update target", zap.String("name", tc.Name))

			t, err := c.targetFactory.CreateSingleClient(tc)
			if err != nil {
				zap.L().Error("unable to create target from TargetConfig", zap.String("name", tc.Name), zap.Error(err))
				return
			} else if t == nil {
				zap.L().Error("provided TargetConfig is invalid", zap.String("name", tc.Name))
				return
			}

			c.collection.AddTarget(tc.Name, t)
		},
		DeleteFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			zap.L().Info("delete target", zap.String("name", tc.Name))

			c.collection.RemoveTarget(tc.Name)
		},
	})
}

func (c *Client) OpenReports(ctx context.Context) ([]openreports.ReportInterface, error) {
	reports := make([]openreports.ReportInterface, 0)
	if c.orClient == nil {
		return reports, nil
	}

	existingReports, err := c.orClient.Reports("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	existingCReports, err := c.orClient.ClusterReports().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, p := range existingReports.Items {
		reports = append(reports, &openreports.ReportAdapter{Report: &p})
	}

	for _, cp := range existingCReports.Items {
		reports = append(reports, &openreports.ClusterReportAdapter{ClusterReport: &cp})
	}

	return reports, nil
}

func (c *Client) WGPolicyReports(ctx context.Context) ([]openreports.ReportInterface, error) {
	reports := make([]openreports.ReportInterface, 0)
	if c.wgpolicyClient == nil {
		return reports, nil
	}

	existingPolrs, err := c.wgpolicyClient.PolicyReports("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	existingcPolrs, err := c.wgpolicyClient.ClusterPolicyReports().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, p := range existingPolrs.Items {
		reports = append(reports, &openreports.ReportAdapter{Report: p.ToOpenReports()})
	}

	for _, cp := range existingcPolrs.Items {
		reports = append(reports, &openreports.ClusterReportAdapter{ClusterReport: cp.ToOpenReports()})
	}

	return reports, nil
}

func (c *Client) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		zap.L().Error("Failed to sync target config cache")
		return
	}

	zap.L().Info("target config cache synced")
}

func NewClient(tcClient crds.Interface, f target.Factory, targets *target.Collection,
	orClient reports.OpenreportsV1alpha1Interface, wgpolicyClient v1alpha2.Wgpolicyk8sV1alpha2Interface,
) *Client {
	tcInformer := informer.NewSharedInformerFactory(tcClient, 0)
	return &Client{
		informer:       tcInformer.Policyreporter().V1alpha1().TargetConfigs().Informer(),
		targetFactory:  f,
		collection:     targets,
		orClient:       orClient,
		wgpolicyClient: wgpolicyClient,
	}
}
