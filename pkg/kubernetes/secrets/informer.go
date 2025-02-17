package secrets

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type Informer interface {
	Sync(targets *target.Collection, stopper chan struct{}) error
	HasSynced() bool
}

type informer struct {
	metaClient metadata.Interface
	synced     bool
	mx         *sync.Mutex
	stopChan   chan struct{}
	factory    target.Factory
	namespace  string
}

func (k *informer) HasSynced() bool {
	return k.synced
}

func (k *informer) Sync(targets *target.Collection, stopper chan struct{}) error {
	k.stopChan = stopper

	factory := metadatainformer.NewFilteredSharedInformerFactory(k.metaClient, 15*time.Minute, k.namespace, nil)

	informer := k.configureInformer(targets, factory.ForResource(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}).Informer())

	factory.Start(stopper)

	if informer != nil && !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		return fmt.Errorf("failed to sync secrets")
	}

	k.synced = true

	zap.L().Info("secret informer sync completed")

	return nil
}

func (k *informer) configureInformer(targets *target.Collection, informer cache.SharedIndexInformer) cache.SharedIndexInformer {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObj interface{}) {
			if s, ok := newObj.(*v1.PartialObjectMetadata); ok {
				for _, t := range targets.Targets() {
					if t.Secret() == s.Name {
						targets.Update(k.UpdateTarget(t, s.Name))
					}
				}
			}
		},
	})

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return informer
}

func (k *informer) UpdateTarget(t *target.Target, secret string) *target.Target {
	updatedTarget := t
	switch t.Type {
	case target.Loki:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateLokiTarget)
	case target.Elasticsearch:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateLokiTarget)
	case target.Slack:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateSlackTarget)
	case target.Discord:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateDiscordTarget)
	case target.Teams:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateTeamsTarget)
	case target.Webhook:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateWebhookTarget)
	case target.Telegram:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateTelegramTarget)
	case target.GoogleChat:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateGoogleChatTarget)
	case target.S3:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateS3Target)
	case target.Kinesis:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateKinesisTarget)
	case target.SecurityHub:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateSecurityHubTarget)
	case target.GCS:
		updatedTarget = createClients(t.Config, t.ParentConfig, k.factory.CreateGCSTarget)
	}

	updatedTarget.ID = t.ID

	return updatedTarget
}

// NewPolicyReportClient new Client for Policy Report Kubernetes API
func NewInformer(metaClient metadata.Interface, factory target.Factory, namespace string) Informer {
	return &informer{
		metaClient: metaClient,
		mx:         new(sync.Mutex),
		factory:    factory,
		namespace:  namespace,
	}
}

func createClients[T any](config, parent any, mapper func(*v1alpha1.Config[T], *v1alpha1.Config[T]) *target.Target) *target.Target {
	return mapper(config.(*v1alpha1.Config[T]), parent.(*v1alpha1.Config[T]))
}
