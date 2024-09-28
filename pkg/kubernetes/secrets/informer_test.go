package secrets_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	metafake "k8s.io/client-go/metadata/fake"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/factory"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

func NewFakeMetaClient() (*metafake.FakeMetadataClient, metafake.MetadataClient) {
	s := metafake.NewTestScheme()
	metav1.AddMetaToScheme(s)

	client := metafake.NewSimpleMetadataClient(s)
	return client, client.Resource(schema.GroupVersionResource{Version: "v1", Resource: "secrets"}).Namespace("default").(metafake.MetadataClient)
}

func Test_SecretInformer(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	t.Run("update secretRef", func(t *testing.T) {
		collection := target.NewCollection(
			&target.Target{
				ID:   uuid.NewString(),
				Type: target.Webhook,
				Client: webhook.NewClient(webhook.Options{
					ClientOptions: target.ClientOptions{
						Name: "Webhook",
					},
				}),
				Config: &target.Config[target.WebhookOptions]{
					Name:      "Webhook",
					SecretRef: secretName,
					Config:    &target.WebhookOptions{},
				},
				ParentConfig: &target.Config[target.WebhookOptions]{Config: &target.WebhookOptions{}},
			},
		)

		client, secret := NewFakeMetaClient()

		informer := secrets.NewInformer(client, factory.NewFactory(secrets.NewClient(newFakeClient()), target.NewResultFilterFactory(nil)), "default")

		err := informer.Sync(collection, stop)
		assert.Nil(t, err)

		assert.True(t, informer.HasSynced())

		secret.CreateFake(&metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: "default"}}, metav1.CreateOptions{})
		time.Sleep(1 * time.Second)

		secret.UpdateFake(&metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: "default"}}, metav1.UpdateOptions{})
		time.Sleep(1 * time.Second)

		assert.Equal(t, collection.Targets()[0].Config.(*target.Config[target.WebhookOptions]).Config.Webhook, "http://localhost:9200/webhook")
	})
}
