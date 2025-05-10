package targetconfig

import (
	"context"

	"github.com/kyverno/policy-reporter/pkg/crd/client/policyreport/clientset/versioned/typed/policyreport/v1alpha2"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"

	report "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	tcv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned"
	tcinformer "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/informers/externalversions"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/target"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
	targetFactory target.Factory
	collection    *target.Collection
	informer      cache.SharedIndexInformer
	polrClient    v1alpha2.Wgpolicyk8sV1alpha2Interface
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
				reports := []report.ReportInterface{}
				existingPolrs, err := c.polrClient.PolicyReports("").List(context.Background(), metav1.ListOptions{})
				if err != nil {
					zap.L().Error("Failed to sync existing policy reports for client", zap.String("name", tc.Name), zap.Error(err))
				}
				existingcPolrs, err := c.polrClient.ClusterPolicyReports().List(context.Background(), metav1.ListOptions{})
				if err != nil {
					zap.L().Error("Failed to sync existing policy reports for client", zap.String("name", tc.Name), zap.Error(err))
				}

				for _, p := range existingPolrs.Items {
					reports = append(reports, &p)
				}

				for _, cp := range existingcPolrs.Items {
					reports = append(reports, &cp)
				}

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

func (c *Client) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		zap.L().Error("Failed to sync target config cache")
		return
	}

	zap.L().Info("target config cache synced")
}

func NewClient(tcClient tcv1alpha1.Interface, f target.Factory, targets *target.Collection, polrClient v1alpha2.Wgpolicyk8sV1alpha2Interface) *Client {
	tcInformer := tcinformer.NewSharedInformerFactory(tcClient, 0)

	return &Client{
		informer:      tcInformer.Policyreporter().V1alpha1().TargetConfigs().Informer(),
		targetFactory: f,
		collection:    targets,
		polrClient:    polrClient,
	}
}
