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

func (c *TargetConfigClient) configureInformer(targetChan chan *target.Collection) {
	f := func(tc *v1alpha1.TargetConfig) (*target.Target, error) {
		t, err := c.targetFactory.CreateSingleClient(tc)
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			targetKey := tc.Name + "," + tc.Namespace + "," + tc.Spec.TargetType
			c.logger.Info(fmt.Sprintf("new target: %s, namespace: %s, type: %s", tc.Name, tc.Namespace, tc.Spec.TargetType))

			target, err := f(tc)
			if err != nil {
				c.logger.Error("unable to create target from TargetConfig: " + err.Error())
			}

			c.targetClients.AddCrdTarget(targetKey, target)
			targetChan <- c.targetClients
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			targetKey := tc.Name + "," + tc.Namespace + "," + tc.Spec.TargetType
			c.logger.Info(fmt.Sprintf("deleting target: %s, namespace: %s, type: %s", tc.Name, tc.Namespace, tc.Spec.TargetType))

			// todo: dont restart informer sync on delete
			c.targetClients.RemoveCrdTarget(targetKey)
			targetChan <- c.targetClients
		},
	})
}

func (c *TargetConfigClient) CreateInformer(targetChan chan *target.Collection) {
	tcInformer := tcinformer.NewSharedInformerFactory(c.tcClient, time.Second)
	inf := tcInformer.Wgpolicyk8s().V1alpha1().TargetConfigs().Informer()
	c.informer = inf

	c.configureInformer(targetChan)
}

func (c *TargetConfigClient) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		c.logger.Error("Failed to sync target config cache") // todo
		return
	}
	c.logger.Info("target config cache synced")
}

func NewTargetConfigClient(tcClient tcv1alpha1.Interface, f target.Factory, targets *target.Collection, logger *zap.Logger) *TargetConfigClient {
	return &TargetConfigClient{
		tcClient:      tcClient,
		targetFactory: f,
		targetClients: targets,
		logger:        logger,
	}
}
