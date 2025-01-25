package targetconfig

import (
	"fmt"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	tcv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned"
	tcinformer "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/informers/externalversions"
	"github.com/kyverno/policy-reporter/pkg/target"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

type TargetConfigClient struct {
	tcClient      tcv1alpha1.Interface
	targetFactory target.Factory
	targetClients *target.Collection
	logger        *zap.Logger
	informer      cache.SharedIndexInformer
}

func (c *TargetConfigClient) configureInformer() {
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			targetKey := tc.Name + "," + tc.Namespace + "," + tc.Spec.TargetType

			t, err := c.targetFactory.CreateSingleClient(tc)
			if err != nil {
				c.logger.Error("unable to create target from TargetConfig: " + err.Error()) // logger is nil
			}
			c.targetClients.AddCrdTarget(targetKey, t)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// todo
		},
		DeleteFunc: func(obj interface{}) {
			// todo
		},
	})
}

func (c *TargetConfigClient) CreateInformer() {
	tcInformer := tcinformer.NewSharedInformerFactory(c.tcClient, time.Second)
	inf := tcInformer.Wgpolicyk8s().V1alpha1().TargetConfigs().Informer()
	c.informer = inf

	c.configureInformer()
}

func (c *TargetConfigClient) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		fmt.Println("Failed to sync target config cache")
		return
	}
	fmt.Println("Target config cache synced")
}

func NewTargetConfigClient(tcClient tcv1alpha1.Interface, f target.Factory, targets *target.Collection, logger *zap.Logger) *TargetConfigClient {
	return &TargetConfigClient{
		tcClient:      tcClient,
		targetFactory: f,
		targetClients: targets,
		logger:        logger,
	}
}
