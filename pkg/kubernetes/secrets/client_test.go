package secrets_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
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
			"apiKey":          []byte("apiKey"),
			"webhook":         []byte("http://localhost:9200/webhook"),
			"accessKeyID":     []byte("accessKeyID"),
			"secretAccessKey": []byte("secretAccessKey"),
			"kmsKeyId":        []byte("kmsKeyId"),
			"token":           []byte("token"),
			"accountID":       []byte("accountID"),
			"database":        []byte("database"),
			"dsn":             []byte("dsn"),
			"typelessApi":     []byte("false"),
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

		if values.APIKey != "apiKey" {
			t.Errorf("Unexpected ApiKey: %s", values.APIKey)
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

		if values.KmsKeyID != "kmsKeyId" {
			t.Errorf("Unexpected KmsKeyId: %s", values.KmsKeyID)
		}

		if values.AccountID != "accountID" {
			t.Errorf("Unexpected AccountID: %s", values.AccountID)
		}

		if values.Database != "database" {
			t.Errorf("Unexpected Database: %s", values.Database)
		}

		if values.DSN != "dsn" {
			t.Errorf("Unexpected DSN: %s", values.DSN)
		}

		if values.TypelessAPI {
			t.Errorf("Unexpected TypelessAPI: %t", values.TypelessAPI)
		}
	})

	t.Run("Get values from not existing secret", func(t *testing.T) {
		_, err := client.Get(context.Background(), "not-exist")
		if !errors.IsNotFound(err) {
			t.Errorf("Expected not found error")
		}
	})
}
