package targetconfig_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corefake "k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned/fake"
	tcv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned/typed/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/factory"
	"github.com/kyverno/policy-reporter/pkg/targetconfig"
)

const secretName = "secret-values"

func newSecretClient() v1.SecretInterface {
	return corefake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"host":            []byte("http://localhost:9200"),
			"username":        []byte("username"),
			"password":        []byte("password"),
			"apiKey":          []byte("apiKey"),
			"webhook":         []byte("http://localhost:9200/webhook"),
			"accessKeyId":     []byte("accessKeyId"),
			"secretAccessKey": []byte("secretAccessKey"),
			"kmsKeyId":        []byte("kmsKeyId"),
			"token":           []byte("token"),
			"accountId":       []byte("accountId"),
			"database":        []byte("database"),
			"dsn":             []byte("dsn"),
			"typelessApi":     []byte("false"),
		},
	}).CoreV1().Secrets("default")
}

func NewFakeClient() (*fake.Clientset, tcv1alpha1.TargetConfigInterface) {
	client := fake.NewSimpleClientset()

	return client, client.PolicyreporterV1alpha1().TargetConfigs("")
}

func Test_TargetConfig_TargetCreation(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})

	defer close(stop)

	kclient, tclient := NewFakeClient()
	collection := target.NewCollection()
	factory := factory.NewFactory(secrets.NewClient(newSecretClient()), target.NewResultFilterFactory(nil))

	client := targetconfig.NewClient(kclient, factory, collection, zap.L())
	client.ConfigureInformer()

	go func() {
		client.Run(stop)
	}()

	tclient.Create(ctx, &v1alpha1.TargetConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.TargetConfigSpec{
			Slack: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:8080",
				},
			},
		},
	}, metav1.CreateOptions{})

	time.Sleep(10 * time.Millisecond)

	assert.Len(t, collection.Targets(), 1)

	target := reflect.ValueOf(collection.Client("test")).Elem()

	assert.NotNil(t, target)

	webhook := target.FieldByName("webhook").String()
	assert.Equal(t, "http://localhost:8080", webhook)
}

func Test_TargetConfig_TargetUpdates(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})

	defer close(stop)

	kclient, tclient := NewFakeClient()
	collection := target.NewCollection()
	factory := factory.NewFactory(secrets.NewClient(newSecretClient()), target.NewResultFilterFactory(nil))

	client := targetconfig.NewClient(kclient, factory, collection, zap.L())
	client.ConfigureInformer()

	go func() {
		client.Run(stop)
	}()

	tclient.Create(ctx, &v1alpha1.TargetConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.TargetConfigSpec{
			Slack: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:8080",
				},
			},
		},
	}, metav1.CreateOptions{})

	time.Sleep(10 * time.Millisecond)

	tclient.Update(ctx, &v1alpha1.TargetConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.TargetConfigSpec{
			Slack: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:9090",
				},
			},
		},
	}, metav1.UpdateOptions{})

	time.Sleep(10 * time.Millisecond)

	assert.Len(t, collection.Targets(), 1)

	target := reflect.ValueOf(collection.Client("test")).Elem()

	assert.NotNil(t, target)

	webhook := target.FieldByName("webhook").String()
	assert.Equal(t, "http://localhost:9090", webhook)
}

func Test_TargetConfig_TargetDeletion(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})

	defer close(stop)

	kclient, tclient := NewFakeClient()
	collection := target.NewCollection()
	factory := factory.NewFactory(secrets.NewClient(newSecretClient()), target.NewResultFilterFactory(nil))

	client := targetconfig.NewClient(kclient, factory, collection, zap.L())
	client.ConfigureInformer()

	go func() {
		client.Run(stop)
	}()

	tclient.Create(ctx, &v1alpha1.TargetConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.TargetConfigSpec{
			Slack: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:8080",
				},
			},
		},
	}, metav1.CreateOptions{})

	time.Sleep(10 * time.Millisecond)

	tclient.Delete(ctx, "test", metav1.DeleteOptions{})

	time.Sleep(10 * time.Millisecond)

	assert.Len(t, collection.Targets(), 0)

	target := collection.Client("test")

	assert.Nil(t, target)
}
