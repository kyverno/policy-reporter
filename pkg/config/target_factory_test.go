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
			"accountId":       []byte("accountID"),
			"typelessApi":     []byte("true"),
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
		APIKey:          "apiKey",
		AccountID:       "accountID",
		AccessKeyID:     "accessKeyId",
		SecretAccessKey: "secretAccessKey",
		KmsKeyID:        "kmsKeyId",
		Token:           "token",
		Credentials:     `{"token": "token", "type": "authorized_user"}`,
		Database:        "database",
		TypelessAPI:     true,
		DSN:             "",
	}
	file, _ := json.MarshalIndent(secretValues, "", " ")
	_ = os.WriteFile(mountedSecret, file, 0o644)
}

var logger = zap.NewNop()

func Test_ResolveTarget(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	clients := factory.CreateClients(&testConfig.Targets)
	if len(clients) != 25 {
		t.Errorf("Expected 25 Client, got %d clients", len(clients))
	}
}

func Test_ResolveTargetsWithoutRequiredConfiguration(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	targets := config.Targets{
		Loki:          &config.Target[config.LokiOptions]{},
		Elasticsearch: &config.Target[config.ElasticsearchOptions]{},
		Slack:         &config.Target[config.SlackOptions]{},
		Discord:       &config.Target[config.WebhookOptions]{},
		Teams:         &config.Target[config.WebhookOptions]{},
		GoogleChat:    &config.Target[config.WebhookOptions]{},
		Webhook:       &config.Target[config.WebhookOptions]{},
		Telegram:      &config.Target[config.TelegramOptions]{},
		S3:            &config.Target[config.S3Options]{},
		Kinesis:       &config.Target[config.KinesisOptions]{},
		SecurityHub:   &config.Target[config.SecurityHubOptions]{},
	}

	if len(factory.CreateClients(&targets)) != 0 {
		t.Error("Expected Client to be nil if no required fields are configured")
	}

	targets = config.Targets{}
	if len(factory.CreateClients(&targets)) != 0 {
		t.Error("Expected Client to be nil if no target is configured")
	}

	targets.S3 = &config.Target[config.S3Options]{
		Config: &config.S3Options{
			AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
		},
	}
}

func Test_S3Validation(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	targets := config.Targets{
		S3: &config.Target[config.S3Options]{
			Config: &config.S3Options{
				AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("S3.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.S3.Config.AWSConfig.AccessKeyID = "access"
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.S3.Config.AWSConfig.SecretAccessKey = "secret"
	t.Run("S3.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})

	targets.S3.Config.AWSConfig.Region = "ru-central1"
	t.Run("S3.Bucket", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})

	targets.S3.Config.ServerSideEncryption = "AES256"
	t.Run("S3.SSE-S3", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})

	targets.S3.Config.ServerSideEncryption = "aws:kms"
	t.Run("S3.SSE-KMS", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})

	targets.S3.Config.BucketKeyEnabled = true
	t.Run("S3.SSE-KMS-S3-KEY", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})

	targets.S3.Config.KmsKeyID = "kmsKeyId"
	t.Run("S3.SSE-KMS-KEY-ID", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})
}

func Test_KinesisValidation(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	targets := config.Targets{
		Kinesis: &config.Target[config.KinesisOptions]{
			Config: &config.KinesisOptions{
				AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("Kinesis.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.Kinesis.Config.AWSConfig.AccessKeyID = "access"
	t.Run("Kinesis.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.Kinesis.Config.AWSConfig.SecretAccessKey = "secret"

	t.Run("Kinesis.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})

	targets.Kinesis.Config.AWSConfig.Region = "ru-central1"

	t.Run("Kinesis.StreamName", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no stream name is configured")
		}
	})
}

func Test_SecurityHubValidation(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	targets := config.Targets{
		SecurityHub: &config.Target[config.SecurityHubOptions]{
			Config: &config.SecurityHubOptions{
				AWSConfig: config.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("SecurityHub.AccountID", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no accountID is configured")
		}
	})

	targets.SecurityHub.Config.AccountID = "accountID"
	t.Run("SecurityHub.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.SecurityHub.Config.AWSConfig.AccessKeyID = "access"
	t.Run("SecurityHub.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.SecurityHub.Config.AWSConfig.SecretAccessKey = "secret"
	t.Run("SecurityHub.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
}

func Test_GCSValidation(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	targets := config.Targets{
		GCS: &config.Target[config.GCSOptions]{
			Config: &config.GCSOptions{},
		},
	}

	t.Run("GCS.Bucket", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})

	targets.GCS.Config.Bucket = "policy-reporter"
	t.Run("GCS.Credentials", func(t *testing.T) {
		if len(factory.CreateClients(&targets)) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
}

func Test_GetValuesFromSecret(t *testing.T) {
	factory := config.NewTargetFactory(secrets.NewClient(newFakeClient()), nil)

	targets := config.Targets{
		Loki:          &config.Target[config.LokiOptions]{SecretRef: secretName},
		Elasticsearch: &config.Target[config.ElasticsearchOptions]{SecretRef: secretName},
		Slack:         &config.Target[config.SlackOptions]{SecretRef: secretName},
		Discord:       &config.Target[config.WebhookOptions]{SecretRef: secretName},
		Teams:         &config.Target[config.WebhookOptions]{SecretRef: secretName},
		GoogleChat:    &config.Target[config.WebhookOptions]{SecretRef: secretName},
		Webhook:       &config.Target[config.WebhookOptions]{SecretRef: secretName},
		Telegram: &config.Target[config.TelegramOptions]{
			SecretRef: secretName,
			Config: &config.TelegramOptions{
				ChatID: "1234",
			},
		},
		S3: &config.Target[config.S3Options]{
			SecretRef: secretName,
			Config: &config.S3Options{
				AWSConfig: config.AWSConfig{Endpoint: "endoint", Region: "region"},
				Bucket:    "bucket",
			},
		},
		Kinesis: &config.Target[config.KinesisOptions]{
			SecretRef: secretName,
			Config: &config.KinesisOptions{
				AWSConfig:  config.AWSConfig{Endpoint: "endoint", Region: "region"},
				StreamName: "stream",
			},
		},
		SecurityHub: &config.Target[config.SecurityHubOptions]{
			SecretRef: secretName,
			Config: &config.SecurityHubOptions{
				AWSConfig: config.AWSConfig{Endpoint: "endoint", Region: "region"},
				AccountID: "accountID",
			},
		},
		GCS: &config.Target[config.GCSOptions]{
			SecretRef: secretName,
			Config: &config.GCSOptions{
				Bucket: "policy-reporter",
			},
		},
	}

	clients := factory.CreateClients(&targets)
	if len(clients) != 12 {
		t.Fatalf("expected 12 clients created, got %d", len(clients))
	}

	t.Run("Get Loki values from Secret", func(t *testing.T) {
		fv := reflect.ValueOf(clients[0]).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/api/prom/push" {
			t.Errorf("Expected host from secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[1]).Elem()

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

	t.Run("Get Slack values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[2]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get Discord values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[3]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get MS Teams values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[4]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get GoogleChat Webhook from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[5]).Elem()

		host := client.FieldByName("webhook").String()
		if host != "http://localhost:9200/webhook" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Telegram Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[6]).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200/bottoken/sendMessage" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Webhook Authentication Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[7]).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
		}
	})

	t.Run("Get none existing secret skips target", func(t *testing.T) {
		clients := factory.CreateClients(&config.Targets{
			Loki: &config.Target[config.LokiOptions]{SecretRef: "not-exist"},
		})

		if len(clients) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}

func Test_CustomFields(t *testing.T) {
	factory := config.NewTargetFactory(nil, nil)

	targets := &config.Targets{
		Loki: &config.Target[config.LokiOptions]{
			Config: &config.LokiOptions{
				HostOptions: config.HostOptions{
					Host: "http://localhost:3100",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Elasticsearch: &config.Target[config.ElasticsearchOptions]{
			Config: &config.ElasticsearchOptions{
				HostOptions: config.HostOptions{
					Host: "http://localhost:9200",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Slack: &config.Target[config.SlackOptions]{
			Config: &config.SlackOptions{
				WebhookOptions: config.WebhookOptions{
					Webhook: "http://localhost:80",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Discord: &config.Target[config.WebhookOptions]{
			Config: &config.WebhookOptions{
				Webhook: "http://discord:80",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Teams: &config.Target[config.WebhookOptions]{
			Config: &config.WebhookOptions{
				Webhook: "http://hook.teams:80",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		GoogleChat: &config.Target[config.WebhookOptions]{
			Config: &config.WebhookOptions{
				Webhook: "http://localhost:900/webhook",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Telegram: &config.Target[config.TelegramOptions]{
			Config: &config.TelegramOptions{
				WebhookOptions: config.WebhookOptions{
					Webhook: "http://localhost:80",
				},
				Token:  "XXX",
				ChatID: "123456",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Webhook: &config.Target[config.WebhookOptions]{
			Config: &config.WebhookOptions{
				Webhook: "http://localhost:8080",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		S3: &config.Target[config.S3Options]{
			Config: &config.S3Options{
				AWSConfig: config.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				Bucket: "test",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Kinesis: &config.Target[config.KinesisOptions]{
			Config: &config.KinesisOptions{
				AWSConfig: config.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				StreamName: "policy-reporter",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		SecurityHub: &config.Target[config.SecurityHubOptions]{
			Config: &config.SecurityHubOptions{
				AWSConfig: config.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				AccountID: "AccountID",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		GCS: &config.Target[config.GCSOptions]{
			Config: &config.GCSOptions{
				Credentials: `{"token": "token", "type": "authorized_user"}`,
				Bucket:      "test",
				Prefix:      "prefix",
			},
			CustomFields: map[string]string{"field": "value"},
		},
	}

	clients := factory.CreateClients(targets)

	if len(clients) != 12 {
		t.Fatalf("expected 12 client created, got %d", len(clients))
	}

	t.Run("Get CustomLabels from Loki", func(t *testing.T) {
		client := reflect.ValueOf(clients[0]).Elem()

		customFields := client.FieldByName("customLabels").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customLabels are added")
		}
	})

	t.Run("Get CustomFields from Elasticsearch", func(t *testing.T) {
		client := reflect.ValueOf(clients[1]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Slack", func(t *testing.T) {
		client := reflect.ValueOf(clients[2]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Discord", func(t *testing.T) {
		client := reflect.ValueOf(clients[3]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from MS Teams", func(t *testing.T) {
		client := reflect.ValueOf(clients[4]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from GoogleChat", func(t *testing.T) {
		client := reflect.ValueOf(clients[5]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Telegram", func(t *testing.T) {
		client := reflect.ValueOf(clients[6]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Webhook", func(t *testing.T) {
		client := reflect.ValueOf(clients[7]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from S3", func(t *testing.T) {
		client := reflect.ValueOf(clients[8]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Kinesis", func(t *testing.T) {
		client := reflect.ValueOf(clients[9]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from GCS", func(t *testing.T) {
		client := reflect.ValueOf(clients[11]).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
}

func Test_GetValuesFromMountedSecret(t *testing.T) {
	factory := config.NewTargetFactory(secrets.NewClient(newFakeClient()), nil)

	mountSecret()
	defer os.Remove(mountedSecret)

	targets := config.Targets{
		Loki:          &config.Target[config.LokiOptions]{MountedSecret: mountedSecret},
		Elasticsearch: &config.Target[config.ElasticsearchOptions]{MountedSecret: mountedSecret},
		Slack:         &config.Target[config.SlackOptions]{MountedSecret: mountedSecret},
		Discord:       &config.Target[config.WebhookOptions]{MountedSecret: mountedSecret},
		Teams:         &config.Target[config.WebhookOptions]{MountedSecret: mountedSecret},
		GoogleChat:    &config.Target[config.WebhookOptions]{MountedSecret: mountedSecret},
		Webhook:       &config.Target[config.WebhookOptions]{MountedSecret: mountedSecret},
		Telegram: &config.Target[config.TelegramOptions]{
			MountedSecret: mountedSecret,
			Config: &config.TelegramOptions{
				ChatID: "1234",
			},
		},
		S3: &config.Target[config.S3Options]{
			MountedSecret: mountedSecret,
			Config: &config.S3Options{
				AWSConfig: config.AWSConfig{Endpoint: "endoint", Region: "region"},
				Bucket:    "bucket",
			},
		},
		Kinesis: &config.Target[config.KinesisOptions]{
			MountedSecret: mountedSecret,
			Config: &config.KinesisOptions{
				AWSConfig:  config.AWSConfig{Endpoint: "endoint", Region: "region"},
				StreamName: "stream",
			},
		},
		SecurityHub: &config.Target[config.SecurityHubOptions]{
			MountedSecret: mountedSecret,
			Config: &config.SecurityHubOptions{
				AWSConfig: config.AWSConfig{Endpoint: "endoint", Region: "region"},
				AccountID: "accountID",
			},
		},
		GCS: &config.Target[config.GCSOptions]{
			MountedSecret: mountedSecret,
			Config: &config.GCSOptions{
				Bucket: "policy-reporter",
			},
		},
	}

	clients := factory.CreateClients(&targets)
	if len(clients) != 12 {
		t.Fatalf("expected 12 client created, got %d", len(clients))
	}

	t.Run("Get Loki values from Secret", func(t *testing.T) {
		fv := reflect.ValueOf(clients[0]).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/api/prom/push" {
			t.Errorf("Expected host from secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[1]).Elem()

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

	t.Run("Get Slack values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[2]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get Discord values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[3]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get MS Teams values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[4]).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get GoogleChat Webhook from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[5]).Elem()

		host := client.FieldByName("webhook").String()
		if host != "http://localhost:9200/webhook" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Telegram Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[6]).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200/bottoken/sendMessage" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Webhook Authentication Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients[7]).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
		}
	})

	t.Run("Get none existing secret skips target", func(t *testing.T) {
		clients := factory.CreateClients(&config.Targets{
			Loki: &config.Target[config.LokiOptions]{SecretRef: "not-exist"},
		})

		if len(clients) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}
