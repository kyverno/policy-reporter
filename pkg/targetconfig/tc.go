package targetconfig

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	tcv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned"
	tcinformer "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/informers/externalversions"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type TargetConfigClient struct {
	tcClient      tcv1alpha1.Interface
	targetFactory target.Factory
	targetClients *target.Collection
	logger        *zap.Logger
	informer      cache.SharedIndexInformer
	tcCount       int
	hasSynced     bool
}

type EventType string

const (
	DeleteTcEvent = "delete"
	CreateTcEvent = "create"
)

type TcEvent struct {
	Type                EventType
	Targets             *target.Collection
	RestartPolrInformer bool
}

func (c *TargetConfigClient) TargetConfigCount() int {
	return c.tcCount
}

func (c *TargetConfigClient) configureInformer(targetChan chan TcEvent) {
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			c.logger.Info(fmt.Sprintf("new target: %s, type: %s", tc.Name, tc.Spec.TargetType))

			t, err := c.targetFactory.CreateSingleClient(tc)
			if err != nil {
				c.logger.Error("unable to create target from TargetConfig: " + err.Error())
				return
			}

			c.targetClients.AddTarget(tc.Name, t)
			targetChan <- TcEvent{Type: CreateTcEvent, Targets: c.targetClients, RestartPolrInformer: !tc.Spec.SkipExisting}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			c.logger.Info(fmt.Sprintf("deleting target: %s, type: %s", tc.Name, tc.Spec.TargetType))

			c.targetClients.RemoveTarget(tc.Name)
			targetChan <- TcEvent{Type: DeleteTcEvent, Targets: c.targetClients}
		},
	})
}

func (c *TargetConfigClient) CreateInformer(targetChan chan TcEvent) error {
	tcInformer := tcinformer.NewSharedInformerFactory(c.tcClient, time.Second)
	inf := tcInformer.Wgpolicyk8s().V1alpha1().TargetConfigs().Informer()
	c.informer = inf

	tcs, err := tcInformer.Wgpolicyk8s().V1alpha1().TargetConfigs().Lister().List(labels.Everything())
	if err != nil {
		return err
	}

	c.tcCount = len(tcs)
	c.configureInformer(targetChan)
	return nil
}

func (c *TargetConfigClient) HasSynced() bool {
	return c.hasSynced
}

func (c *TargetConfigClient) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		c.logger.Error("Failed to sync target config cache") // todo
		return
	}

	c.hasSynced = true
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
