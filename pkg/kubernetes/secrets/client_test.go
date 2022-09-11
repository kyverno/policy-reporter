package secrets_test

import (
	"context"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const secretName = "secret-values"

func newFakeClient() v1.SecretInterface {
	return fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"host":            []byte("http://localhost:9200"),
			"username":        []byte("username"),
			"password":        []byte("password"),
			"webhook":         []byte("http://localhost:9200/webhook"),
			"accessKeyID":     []byte("accessKeyID"),
			"secretAccessKey": []byte("secretAccessKey"),
			"token":           []byte("token"),
		},
	}).CoreV1().Secrets("default")
}

func Test_Client(t *testing.T) {
	client := secrets.NewClient(newFakeClient())

	t.Run("Get values from existing secret", func(t *testing.T) {
		values, err := client.Get(context.Background(), secretName)
		if err != nil {
			t.Errorf("Unexpected error while fetching secret: %s", err)
		}

		if values.Host != "http://localhost:9200" {
			t.Errorf("Unexpected Host: %s", values.Host)
		}

		if values.Webhook != "http://localhost:9200/webhook" {
			t.Errorf("Unexpected Webhook: %s", values.Webhook)
		}

		if values.Username != "username" {
			t.Errorf("Unexpected Username: %s", values.Username)
		}

		if values.Password != "password" {
			t.Errorf("Unexpected Password: %s", values.Password)
		}

		if values.AccessKeyID != "accessKeyID" {
			t.Errorf("Unexpected AccessKeyID: %s", values.AccessKeyID)
		}

		if values.SecretAccessKey != "secretAccessKey" {
			t.Errorf("Unexpected SecretAccessKey: %s", values.SecretAccessKey)
		}

		if values.Token != "token" {
			t.Errorf("Unexpected Token: %s", values.Token)
		}
	})

	t.Run("Get values from not existing secret", func(t *testing.T) {
		_, err := client.Get(context.Background(), "not-exist")
		if !errors.IsNotFound(err) {
			t.Errorf("Expected not found error")
		}
	})
}
