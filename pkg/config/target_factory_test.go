package config_test

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/config"
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
			"webhook":         []byte("http://localhost:9200/webhook"),
			"accessKeyID":     []byte("accessKeyID"),
			"secretAccessKey": []byte("secretAccessKey"),
			"token":           []byte("token"),
		},
	}).CoreV1().Secrets("default")
}

func Test_ResolveTarget(t *testing.T) {
	factory := config.NewTargetFactory("", nil)

	t.Run("Loki", func(t *testing.T) {
		clients := factory.LokiClients(testConfig.Loki)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("Elasticsearch", func(t *testing.T) {
		clients := factory.ElasticsearchClients(testConfig.Elasticsearch)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("Slack", func(t *testing.T) {
		clients := factory.SlackClients(testConfig.Slack)
		if len(clients) != 2 {
			t.Error("Expected Client, got nil")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		clients := factory.DiscordClients(testConfig.Discord)
		if len(clients) != 2 {
			t.Error("Expected Client, got nil")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		clients := factory.TeamsClients(testConfig.Teams)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("Webhook", func(t *testing.T) {
		clients := factory.WebhookClients(testConfig.Webhook)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("S3", func(t *testing.T) {
		clients := factory.S3Clients(testConfig.S3)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("Kinesis", func(t *testing.T) {
		clients := factory.KinesisClients(testConfig.Kinesis)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
}

func Test_ResolveTargetWithoutHost(t *testing.T) {
	factory := config.NewTargetFactory("", nil)

	t.Run("Loki", func(t *testing.T) {
		if len(factory.LokiClients(config.Loki{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Elasticsearch", func(t *testing.T) {
		if len(factory.ElasticsearchClients(config.Elasticsearch{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Slack", func(t *testing.T) {
		if len(factory.SlackClients(config.Slack{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		if len(factory.DiscordClients(config.Discord{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		if len(factory.TeamsClients(config.Teams{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Webhook", func(t *testing.T) {
		if len(factory.WebhookClients(config.Webhook{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("S3.Endoint", func(t *testing.T) {
		if len(factory.S3Clients(config.S3{})) != 0 {
			t.Error("Expected Client to be nil if no endpoint is configured")
		}
	})
	t.Run("S3.AccessKey", func(t *testing.T) {
		if len(factory.S3Clients(config.S3{Endpoint: "https://storage.yandexcloud.net"})) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		if len(factory.S3Clients(config.S3{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access"})) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("S3.Region", func(t *testing.T) {
		if len(factory.S3Clients(config.S3{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret"})) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("S3.Bucket", func(t *testing.T) {
		if len(factory.S3Clients(config.S3{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"})) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})
	t.Run("Kinesis.Endoint", func(t *testing.T) {
		if len(factory.KinesisClients(config.Kinesis{})) != 0 {
			t.Error("Expected Client to be nil if no endpoint is configured")
		}
	})
	t.Run("Kinesis.AccessKey", func(t *testing.T) {
		if len(factory.KinesisClients(config.Kinesis{Endpoint: "https://yds.serverless.yandexcloud.net"})) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("Kinesis.SecretAccessKey", func(t *testing.T) {
		if len(factory.KinesisClients(config.Kinesis{Endpoint: "https://yds.serverless.yandexcloud.net", AccessKeyID: "access"})) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("Kinesis.Region", func(t *testing.T) {
		if len(factory.KinesisClients(config.Kinesis{Endpoint: "https://yds.serverless.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret"})) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("Kinesis.StreamName", func(t *testing.T) {
		if len(factory.KinesisClients(config.Kinesis{Endpoint: "https://yds.serverless.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"})) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})
}

func Test_GetValuesFromSecret(t *testing.T) {
	factory := config.NewTargetFactory("default", secrets.NewClient(newFakeClient()))

	t.Run("Get Loki values from Secret", func(t *testing.T) {
		clients := factory.LokiClients(config.Loki{SecretRef: secretName})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		fv := reflect.ValueOf(clients[0]).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/api/prom/push" {
			t.Errorf("Expected host from secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from Secret", func(t *testing.T) {
		clients := factory.ElasticsearchClients(config.Elasticsearch{SecretRef: secretName})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200" {
			t.Errorf("Expected host from secret, got %s", host)
		}

		username := client.FieldByName("username").String()
		if username != "username" {
			t.Errorf("Expected username from secret, got %s", username)
		}

		password := client.FieldByName("password").String()
		if password != "password" {
			t.Errorf("Expected password from secret, got %s", password)
		}
	})

	t.Run("Get Discord values from Secret", func(t *testing.T) {
		clients := factory.DiscordClients(config.Discord{SecretRef: secretName})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get MS Teams values from Secret", func(t *testing.T) {
		clients := factory.TeamsClients(config.Teams{SecretRef: secretName})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get Slack values from Secret", func(t *testing.T) {
		clients := factory.SlackClients(config.Slack{SecretRef: secretName})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get Webhook Authentication Token from Secret", func(t *testing.T) {
		clients := factory.WebhookClients(config.Webhook{SecretRef: secretName})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
		}
	})

	t.Run("Get S3 values from Secret", func(t *testing.T) {
		clients := factory.S3Clients(config.S3{SecretRef: secretName, Endpoint: "endoint", Bucket: "bucket", Region: "region"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get Kinesis values from Secret", func(t *testing.T) {
		clients := factory.KinesisClients(config.Kinesis{SecretRef: secretName, Endpoint: "endpoint", StreamName: "stream", Region: "region"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get none existing secret skips target", func(t *testing.T) {
		clients := factory.LokiClients(config.Loki{SecretRef: "no-exist"})
		if len(clients) != 0 {
			t.Error("Expected client are skipped")
		}
	})

	t.Run("Get CustomFields from Slack", func(t *testing.T) {
		clients := factory.SlackClients(config.Slack{CustomFields: map[string]string{"field": "value"}, Webhook: "http://localhost"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Discord", func(t *testing.T) {
		clients := factory.DiscordClients(config.Discord{CustomFields: map[string]string{"field": "value"}, Webhook: "http://localhost"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from MS Teams", func(t *testing.T) {
		clients := factory.TeamsClients(config.Teams{CustomFields: map[string]string{"field": "value"}, Webhook: "http://localhost"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Elasticsearch", func(t *testing.T) {
		clients := factory.ElasticsearchClients(config.Elasticsearch{CustomFields: map[string]string{"field": "value"}, Host: "http://localhost"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Webhook", func(t *testing.T) {
		clients := factory.WebhookClients(config.Webhook{CustomFields: map[string]string{"field": "value"}, Host: "http://localhost"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Kinesis", func(t *testing.T) {
		clients := factory.KinesisClients(testConfig.Kinesis)
		if len(clients) < 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from S3", func(t *testing.T) {
		clients := factory.S3Clients(testConfig.S3)
		if len(clients) < 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomLabels from Loki", func(t *testing.T) {
		clients := factory.LokiClients(config.Loki{CustomLabels: map[string]string{"label": "value"}, Host: "http://localhost"})
		if len(clients) < 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customLabels").MapKeys()
		if customFields[0].String() != "label" {
			t.Errorf("Expected customLabels are added")
		}
	})
}
