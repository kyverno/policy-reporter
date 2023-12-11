package config_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
)

const (
	secretName    = "secret-values"
	mountedSecret = "/tmp/secrets-9999"
)

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
			"credentials":     []byte(`{"token": "token", "type": "authorized_user"}`),
			"database":        []byte("database"),
			"dsn":             []byte(""),
		},
	}).CoreV1().Secrets("default")
}

func mountSecret() {
	secretValues := secrets.Values{
		Host:            "http://localhost:9200",
		Webhook:         "http://localhost:9200/webhook",
		Username:        "username",
		Password:        "password",
		ApiKey:          "apiKey",
		AccessKeyID:     "accessKeyId",
		SecretAccessKey: "secretAccessKey",
		KmsKeyID:        "kmsKeyId",
		Token:           "token",
		Credentials:     `{"token": "token", "type": "authorized_user"}`,
		Database:        "database",
		DSN:             "",
	}
	file, _ := json.MarshalIndent(secretValues, "", " ")
	_ = os.WriteFile(mountedSecret, file, 0o644)
}

var logger = zap.NewNop()

func Test_ResolveTarget(t *testing.T) {
	factory := config.NewTargetFactory(nil)

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
		if len(clients) != 3 {
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
	t.Run("Telegram", func(t *testing.T) {
		clients := factory.TelegramClients(testConfig.Telegram)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
	t.Run("GoogleChat", func(t *testing.T) {
		clients := factory.GoogleChatClients(testConfig.GoogleChat)
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
	t.Run("GCS", func(t *testing.T) {
		clients := factory.GCSClients(testConfig.GCS)
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
	t.Run("SecurityHub", func(t *testing.T) {
		clients := factory.SecurityHubs(testConfig.SecurityHub)
		if len(clients) != 2 {
			t.Errorf("Expected 2 Client, got %d clients", len(clients))
		}
	})
}

func Test_ResolveTargetWithoutHost(t *testing.T) {
	factory := config.NewTargetFactory(nil)

	t.Run("Loki", func(t *testing.T) {
		if len(factory.LokiClients(&config.Loki{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Elasticsearch", func(t *testing.T) {
		if len(factory.ElasticsearchClients(&config.Elasticsearch{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Slack", func(t *testing.T) {
		if len(factory.SlackClients(&config.Slack{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Discord", func(t *testing.T) {
		if len(factory.DiscordClients(&config.Discord{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Teams", func(t *testing.T) {
		if len(factory.TeamsClients(&config.Teams{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Webhook", func(t *testing.T) {
		if len(factory.WebhookClients(&config.Webhook{})) != 0 {
			t.Error("Expected Client to be nil if no host is configured")
		}
	})
	t.Run("Telegram", func(t *testing.T) {
		if len(factory.TelegramClients(&config.Telegram{})) != 0 {
			t.Error("Expected Client to be nil if no chatID is configured")
		}
	})
	t.Run("GoogleChat", func(t *testing.T) {
		if len(factory.GoogleChatClients(&config.GoogleChat{})) != 0 {
			t.Error("Expected Client to be nil if no webhook is configured")
		}
	})
	t.Run("S3.Endoint", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{})) != 0 {
			t.Error("Expected Client to be nil if no endpoint is configured")
		}
	})
	t.Run("S3.AccessKey", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net"}})) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access"}})) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("S3.Region", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret"}})) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("S3.Bucket", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"}})) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})
	t.Run("S3.SSE-S3", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"}, ServerSideEncryption: "AES256"})) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})
	t.Run("S3.SSE-KMS", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"}, ServerSideEncryption: "aws:kms"})) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})
	t.Run("S3.SSE-KMS-S3-KEY", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"}, BucketKeyEnabled: true, ServerSideEncryption: "aws:kms"})) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})
	t.Run("S3.SSE-KMS-KEY-ID", func(t *testing.T) {
		if len(factory.S3Clients(&config.S3{AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"}, ServerSideEncryption: "aws:kms", KmsKeyID: "kmsKeyId"})) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})
	t.Run("Kinesis.Endpoint", func(t *testing.T) {
		if len(factory.KinesisClients(&config.Kinesis{})) != 0 {
			t.Error("Expected Client to be nil if no endpoint is configured")
		}
	})
	t.Run("Kinesis.AccessKey", func(t *testing.T) {
		if len(factory.KinesisClients(&config.Kinesis{AWSConfig: config.AWSConfig{Endpoint: "https://yds.serverless.yandexcloud.net"}})) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("Kinesis.SecretAccessKey", func(t *testing.T) {
		if len(factory.KinesisClients(&config.Kinesis{AWSConfig: config.AWSConfig{Endpoint: "https://yds.serverless.yandexcloud.net", AccessKeyID: "access"}})) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("Kinesis.Region", func(t *testing.T) {
		if len(factory.KinesisClients(&config.Kinesis{AWSConfig: config.AWSConfig{Endpoint: "https://yds.serverless.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret"}})) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("Kinesis.StreamName", func(t *testing.T) {
		if len(factory.KinesisClients(&config.Kinesis{AWSConfig: config.AWSConfig{Endpoint: "https://yds.serverless.yandexcloud.net", AccessKeyID: "access", SecretAccessKey: "secret", Region: "ru-central1"}})) != 0 {
			t.Error("Expected Client to be nil if no stream name is configured")
		}
	})
	t.Run("SecurityHub.AccountID", func(t *testing.T) {
		if len(factory.SecurityHubs(&config.SecurityHub{})) != 0 {
			t.Error("Expected Client to be nil if no accountID is configured")
		}
	})
	t.Run("SecurityHub.AccessKey", func(t *testing.T) {
		if len(factory.SecurityHubs(&config.SecurityHub{AccountID: "accountID"})) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
	t.Run("SecurityHub.SecretAccessKey", func(t *testing.T) {
		if len(factory.SecurityHubs(&config.SecurityHub{AccountID: "accountID", AWSConfig: config.AWSConfig{AccessKeyID: "access"}})) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})
	t.Run("SecurityHub.Region", func(t *testing.T) {
		if len(factory.SecurityHubs(&config.SecurityHub{AccountID: "accountID", AWSConfig: config.AWSConfig{AccessKeyID: "access", SecretAccessKey: "secret"}})) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
	t.Run("GCS.Bucket", func(t *testing.T) {
		if len(factory.GCSClients(&config.GCS{})) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})
	t.Run("GCS.Credentials", func(t *testing.T) {
		if len(factory.GCSClients(&config.GCS{Bucket: "policy-reporter"})) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
}

func Test_GetValuesFromSecret(t *testing.T) {
	factory := config.NewTargetFactory(secrets.NewClient(newFakeClient()))

	t.Run("Get Loki values from Secret", func(t *testing.T) {
		clients := factory.LokiClients(&config.Loki{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
		if len(clients) != 1 {
			t.Fatal("Expected one client created")
		}

		fv := reflect.ValueOf(clients[0]).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/api/prom/push" {
			t.Errorf("Expected host from secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from Secret", func(t *testing.T) {
		clients := factory.ElasticsearchClients(&config.Elasticsearch{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
		if len(clients) != 1 {
			t.Fatal("Expected one client created")
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

		rotation := client.FieldByName("rotation").String()
		if rotation != "daily" {
			t.Errorf("Expected rotation from secret, got %s", rotation)
		}

		index := client.FieldByName("index").String()
		if index != "policy-reporter" {
			t.Errorf("Expected rotation from secret, got %s", index)
		}

		password := client.FieldByName("password").String()
		if password != "password" {
			t.Errorf("Expected password from secret, got %s", password)
		}

		apiKey := client.FieldByName("apiKey").String()
		if apiKey != "apiKey" {
			t.Errorf("Expected apiKey from secret, got %s", apiKey)
		}
	})

	t.Run("Get Discord values from Secret", func(t *testing.T) {
		clients := factory.DiscordClients(&config.Discord{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
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
		clients := factory.TeamsClients(&config.Teams{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
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
		clients := factory.SlackClients(&config.Slack{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
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
		clients := factory.WebhookClients(&config.Webhook{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
		}
	})

	t.Run("Get Telegram Token from Secret", func(t *testing.T) {
		clients := factory.TelegramClients(&config.Telegram{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}, ChatID: "1234"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200/bottoken/sendMessage" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})
	t.Run("Get GoogleChat Webhook from Secret", func(t *testing.T) {
		clients := factory.GoogleChatClients(&config.GoogleChat{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		host := client.FieldByName("webhook").String()
		if host != "http://localhost:9200/webhook" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get S3 values from Secret", func(t *testing.T) {
		clients := factory.S3Clients(&config.S3{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}, AWSConfig: config.AWSConfig{Endpoint: "endoint", Region: "region"}, Bucket: "bucket"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get S3 values from Secret with KMS", func(t *testing.T) {
		clients := factory.S3Clients(&config.S3{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}, AWSConfig: config.AWSConfig{Endpoint: "endoint", Region: "region"}, Bucket: "bucket", BucketKeyEnabled: true, ServerSideEncryption: "aws:kms"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get Kinesis values from Secret", func(t *testing.T) {
		clients := factory.KinesisClients(&config.Kinesis{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}, AWSConfig: config.AWSConfig{Endpoint: "endpoint", Region: "region"}, StreamName: "stream"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get GCS values from Secret", func(t *testing.T) {
		clients := factory.GCSClients(&config.GCS{TargetBaseOptions: config.TargetBaseOptions{SecretRef: secretName}, Bucket: "bucket"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get none existing secret skips target", func(t *testing.T) {
		clients := factory.LokiClients(&config.Loki{TargetBaseOptions: config.TargetBaseOptions{SecretRef: "no-exist"}})
		if len(clients) != 0 {
			t.Error("Expected client are skipped")
		}
	})

	t.Run("Get CustomFields from Slack", func(t *testing.T) {
		clients := factory.SlackClients(&config.Slack{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Webhook: "http://localhost"})
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
		clients := factory.DiscordClients(&config.Discord{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Webhook: "http://localhost"})
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
		clients := factory.TeamsClients(&config.Teams{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Webhook: "http://localhost"})
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
		clients := factory.ElasticsearchClients(&config.Elasticsearch{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Host: "http://localhost"})
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
		clients := factory.WebhookClients(&config.Webhook{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Host: "http://localhost"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Telegram", func(t *testing.T) {
		clients := factory.TelegramClients(&config.Telegram{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Token: "XXX", ChatID: "1234"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from GoogleChat", func(t *testing.T) {
		clients := factory.GoogleChatClients(&config.GoogleChat{TargetBaseOptions: config.TargetBaseOptions{CustomFields: map[string]string{"field": "value"}}, Webhook: "http;//googlechat.webhook"})
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
		clients := factory.LokiClients(&config.Loki{
			CustomLabels: map[string]string{"label": "value"},
			Host:         "http://localhost",
		})
		if len(clients) < 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customLabels").MapKeys()
		if customFields[0].String() != "label" {
			t.Errorf("Expected customLabels are added")
		}
	})
	t.Run("Get CustomFields from GCS", func(t *testing.T) {
		clients := factory.GCSClients(testConfig.GCS)
		if len(clients) < 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
}

func Test_GetValuesFromMountedSecret(t *testing.T) {
	factory := config.NewTargetFactory(nil)
	mountSecret()
	defer os.Remove(mountedSecret)

	t.Run("Get Loki values from MountedSecret", func(t *testing.T) {
		clients := factory.LokiClients(&config.Loki{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		fv := reflect.ValueOf(clients[0]).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/api/prom/push" {
			t.Errorf("Expected host from mounted secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from MountedSecret", func(t *testing.T) {
		clients := factory.ElasticsearchClients(&config.Elasticsearch{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200" {
			t.Errorf("Expected host from mounted secret, got %s", host)
		}

		username := client.FieldByName("username").String()
		if username != "username" {
			t.Errorf("Expected username from mounted secret, got %s", username)
		}

		password := client.FieldByName("password").String()
		if password != "password" {
			t.Errorf("Expected password from mounted secret, got %s", password)
		}

		apiKey := client.FieldByName("apiKey").String()
		if apiKey != "apiKey" {
			t.Errorf("Expected apiKey from secret, got %s", apiKey)
		}
	})

	t.Run("Get Discord values from MountedSecret", func(t *testing.T) {
		clients := factory.DiscordClients(&config.Discord{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from mounted secret, got %s", webhook)
		}
	})

	t.Run("Get MS Teams values from MountedSecret", func(t *testing.T) {
		clients := factory.TeamsClients(&config.Teams{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from mounted secret, got %s", webhook)
		}
	})

	t.Run("Get Slack values from MountedSecret", func(t *testing.T) {
		clients := factory.SlackClients(&config.Slack{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from mounted secret, got %s", webhook)
		}
	})

	t.Run("Get Webhook Authentication Token from MountedSecret", func(t *testing.T) {
		clients := factory.WebhookClients(&config.Webhook{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from mounted secret, got %s", token)
		}
	})

	t.Run("Get Telegram Token from MountedSecret", func(t *testing.T) {
		clients := factory.TelegramClients(&config.Telegram{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}, ChatID: "123"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		token := client.FieldByName("host").String()
		if token != "http://localhost:9200/bottoken/sendMessage" {
			t.Errorf("Expected token from mounted secret, got %s", token)
		}
	})

	t.Run("Get GoogleChat Webhook from MountedSecret", func(t *testing.T) {
		clients := factory.GoogleChatClients(&config.GoogleChat{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}

		client := reflect.ValueOf(clients[0]).Elem()

		token := client.FieldByName("webhook").String()
		if token != "http://localhost:9200/webhook" {
			t.Errorf("Expected token from mounted secret, got %s", token)
		}
	})

	t.Run("Get S3 values from MountedSecret", func(t *testing.T) {
		clients := factory.S3Clients(&config.S3{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}, AWSConfig: config.AWSConfig{Endpoint: "endpoint", Region: "region"}, Bucket: "bucket"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get S3 values from MountedSecret with KMS", func(t *testing.T) {
		clients := factory.S3Clients(&config.S3{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}, AWSConfig: config.AWSConfig{Endpoint: "endpoint", Region: "region"}, Bucket: "bucket", BucketKeyEnabled: true, ServerSideEncryption: "aws:kms"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get Kinesis values from MountedSecret", func(t *testing.T) {
		clients := factory.KinesisClients(&config.Kinesis{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}, AWSConfig: config.AWSConfig{Endpoint: "endpoint", Region: "region"}, StreamName: "stream"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get GCS values from MountedSecret", func(t *testing.T) {
		clients := factory.GCSClients(&config.GCS{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: mountedSecret}, Bucket: "bucket"})
		if len(clients) != 1 {
			t.Error("Expected one client created")
		}
	})

	t.Run("Get none existing mounted secret skips target", func(t *testing.T) {
		clients := factory.LokiClients(&config.Loki{TargetBaseOptions: config.TargetBaseOptions{MountedSecret: "no-exists"}})
		if len(clients) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}
