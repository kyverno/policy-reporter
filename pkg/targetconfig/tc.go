package targetconfig

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			c.logger.Info(fmt.Sprintf("new target: %s", tc.Name))

			t, err := c.targetFactory.CreateSingleClient(tc)
			if err != nil {
				c.logger.Error("unable to create target from TargetConfig: " + err.Error())
				return
			}

			c.targetClients.AddTarget(tc.Name, t)
			targetChan <- TcEvent{Type: CreateTcEvent, Targets: c.targetClients}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			tc := newObj.(*v1alpha1.TargetConfig)
			c.logger.Info(fmt.Sprintf("update target: %s", tc.Name))

			t, err := c.targetFactory.CreateSingleClient(tc)
			if err != nil {
				c.logger.Error("unable to create target from TargetConfig: " + err.Error())
				return
			}

			c.targetClients.AddTarget(tc.Name, t)
			targetChan <- TcEvent{Type: CreateTcEvent, Targets: c.targetClients}
		},
		DeleteFunc: func(obj interface{}) {
			tc := obj.(*v1alpha1.TargetConfig)
			c.logger.Info(fmt.Sprintf("deleting target: %s", tc.Name))

			c.targetClients.RemoveTarget(tc.Name)
			targetChan <- TcEvent{Type: DeleteTcEvent, Targets: c.targetClients}
		},
	})
}

func (c *TargetConfigClient) CreateInformer(targetChan chan TcEvent) error {
	tcInformer := tcinformer.NewSharedInformerFactory(c.tcClient, 0)
	inf := tcInformer.Policyreporter().V1alpha1().TargetConfigs().Informer()
	c.informer = inf

	tcs, err := c.tcClient.PolicyreporterV1alpha1().TargetConfigs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	c.tcCount = len(tcs.Items)
	c.configureInformer(targetChan)
	return nil
}

func (c *TargetConfigClient) HasSynced() bool {
	return c.hasSynced
}

func (c *TargetConfigClient) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		c.logger.Error("Failed to sync target config cache")
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
