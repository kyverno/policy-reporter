package factory_test

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

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/factory"
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
			"channel":         []byte("general"),
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
		Channel:         "general",
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

var targets = target.Targets{
	Loki: &target.Config[target.LokiOptions]{
		Config: &target.LokiOptions{
			HostOptions: target.HostOptions{
				Host:    "http://localhost:3100",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.LokiOptions]{
			{
				CustomFields: map[string]string{"label2": "value2"},
			},
		},
	},
	Elasticsearch: &target.Config[target.ElasticsearchOptions]{
		Config: &target.ElasticsearchOptions{
			HostOptions: target.HostOptions{
				Host:    "http://localhost:9200",
				SkipTLS: true,
			},
			Index:    "policy-reporter",
			Rotation: "daily",
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.ElasticsearchOptions]{{}},
	},
	Slack: &target.Config[target.SlackOptions]{
		Config: &target.SlackOptions{
			WebhookOptions: target.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.SlackOptions]{{
			Config: &target.SlackOptions{
				WebhookOptions: target.WebhookOptions{
					Webhook: "http://localhost:9200",
				},
			},
		}, {
			Config: &target.SlackOptions{
				Channel: "general",
			},
		}},
	},
	Discord: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://discord:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.WebhookOptions]{{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	Teams: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://hook.teams:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.WebhookOptions]{{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	GoogleChat: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://localhost:900/webhook",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.WebhookOptions]{{}},
	},
	Telegram: &target.Config[target.TelegramOptions]{
		Config: &target.TelegramOptions{
			WebhookOptions: target.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
			Token:  "XXX",
			ChatID: "123456",
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.TelegramOptions]{{
			Config: &target.TelegramOptions{
				ChatID: "1234567",
			},
		}},
	},
	Webhook: &target.Config[target.WebhookOptions]{
		Config: &target.WebhookOptions{
			Webhook: "http://localhost:8080",
			SkipTLS: true,
			Headers: map[string]string{
				"X-Custom": "Header",
			},
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*target.Config[target.WebhookOptions]{{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:8081",
				Headers: map[string]string{
					"X-Custom-2": "Header",
				},
			},
		}},
	},
	S3: &target.Config[target.S3Options]{
		Config: &target.S3Options{
			AWSConfig: target.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			Bucket:               "test",
			BucketKeyEnabled:     false,
			KmsKeyID:             "",
			ServerSideEncryption: "",
			PathStyle:            true,
			Prefix:               "prefix",
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.S3Options]{{}},
	},
	Kinesis: &target.Config[target.KinesisOptions]{
		Config: &target.KinesisOptions{
			AWSConfig: target.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			StreamName: "policy-reporter",
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.KinesisOptions]{{}},
	},
	SecurityHub: &target.Config[target.SecurityHubOptions]{
		Config: &target.SecurityHubOptions{
			AWSConfig: target.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			AccountID: "AccountID",
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.SecurityHubOptions]{{}},
	},
	GCS: &target.Config[target.GCSOptions]{
		Config: &target.GCSOptions{
			Credentials: `{"token": "token", "type": "authorized_user"}`,
			Bucket:      "test",
			Prefix:      "prefix",
		},
		SkipExisting:    true,
		MinimumSeverity: v1alpha2.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*target.Config[target.GCSOptions]{{}},
	},
}

func Test_ResolveTarget(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	clients := factory.CreateClients(&targets)
	if len(clients.Clients()) != 25 {
		t.Errorf("Expected 25 Client, got %d clients", len(clients.Clients()))
	}
}

func Test_ResolveTargetsWithoutRequiredConfiguration(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		Loki:          &target.Config[target.LokiOptions]{},
		Elasticsearch: &target.Config[target.ElasticsearchOptions]{},
		Slack:         &target.Config[target.SlackOptions]{},
		Discord:       &target.Config[target.WebhookOptions]{},
		Teams:         &target.Config[target.WebhookOptions]{},
		GoogleChat:    &target.Config[target.WebhookOptions]{},
		Webhook:       &target.Config[target.WebhookOptions]{},
		Telegram:      &target.Config[target.TelegramOptions]{},
		S3:            &target.Config[target.S3Options]{},
		Kinesis:       &target.Config[target.KinesisOptions]{},
		SecurityHub:   &target.Config[target.SecurityHubOptions]{},
	}

	if len(factory.CreateClients(&targets).Clients()) != 0 {
		t.Error("Expected Client to be nil if no required fields are configured")
	}

	targets = target.Targets{}
	if len(factory.CreateClients(&targets).Clients()) != 0 {
		t.Error("Expected Client to be nil if no target is configured")
	}

	targets.S3 = &target.Config[target.S3Options]{
		Config: &target.S3Options{
			AWSConfig: target.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
		},
	}
}

func Test_S3Validation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		S3: &target.Config[target.S3Options]{
			Config: &target.S3Options{
				AWSConfig: target.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("S3.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.S3.Config.AWSConfig.AccessKeyID = "access"
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.S3.Config.AWSConfig.SecretAccessKey = "secret"
	t.Run("S3.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})

	targets.S3.Config.AWSConfig.Region = "ru-central1"
	t.Run("S3.Bucket", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})

	targets.S3.Config.ServerSideEncryption = "AES256"
	t.Run("S3.SSE-S3", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})

	targets.S3.Config.ServerSideEncryption = "aws:kms"
	t.Run("S3.SSE-KMS", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})

	targets.S3.Config.BucketKeyEnabled = true
	t.Run("S3.SSE-KMS-S3-KEY", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})

	targets.S3.Config.KmsKeyID = "kmsKeyId"
	t.Run("S3.SSE-KMS-KEY-ID", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if server side encryption is not configured")
		}
	})
}

func Test_KinesisValidation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		Kinesis: &target.Config[target.KinesisOptions]{
			Config: &target.KinesisOptions{
				AWSConfig: target.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("Kinesis.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.Kinesis.Config.AWSConfig.AccessKeyID = "access"
	t.Run("Kinesis.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.Kinesis.Config.AWSConfig.SecretAccessKey = "secret"

	t.Run("Kinesis.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})

	targets.Kinesis.Config.AWSConfig.Region = "ru-central1"

	t.Run("Kinesis.StreamName", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no stream name is configured")
		}
	})
}

func Test_SecurityHubValidation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		SecurityHub: &target.Config[target.SecurityHubOptions]{
			Config: &target.SecurityHubOptions{
				AWSConfig: target.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("SecurityHub.AccountID", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accountID is configured")
		}
	})

	targets.SecurityHub.Config.AccountID = "accountID"
	t.Run("SecurityHub.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.SecurityHub.Config.AWSConfig.AccessKeyID = "access"
	t.Run("SecurityHub.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.SecurityHub.Config.AWSConfig.SecretAccessKey = "secret"
	t.Run("SecurityHub.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
}

func Test_GCSValidation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		GCS: &target.Config[target.GCSOptions]{
			Config: &target.GCSOptions{
				Credentials: "{}",
			},
		},
	}

	t.Run("GCS.Bucket", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no bucket is configured")
		}
	})

	targets.GCS.Config.Bucket = "policy-reporter"
	t.Run("GCS.Credentials", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})
}

func Test_GetValuesFromSecret(t *testing.T) {
	factory := factory.NewFactory(secrets.NewClient(newFakeClient()), nil)

	targets := target.Targets{
		Loki:          &target.Config[target.LokiOptions]{SecretRef: secretName},
		Elasticsearch: &target.Config[target.ElasticsearchOptions]{SecretRef: secretName},
		Slack:         &target.Config[target.SlackOptions]{SecretRef: secretName},
		Discord:       &target.Config[target.WebhookOptions]{SecretRef: secretName},
		Teams:         &target.Config[target.WebhookOptions]{SecretRef: secretName},
		GoogleChat:    &target.Config[target.WebhookOptions]{SecretRef: secretName},
		Webhook:       &target.Config[target.WebhookOptions]{SecretRef: secretName},
		Telegram: &target.Config[target.TelegramOptions]{
			SecretRef: secretName,
			Config: &target.TelegramOptions{
				ChatID: "1234",
			},
		},
		S3: &target.Config[target.S3Options]{
			SecretRef: secretName,
			Config: &target.S3Options{
				AWSConfig: target.AWSConfig{Endpoint: "endoint", Region: "region"},
				Bucket:    "bucket",
			},
		},
		Kinesis: &target.Config[target.KinesisOptions]{
			SecretRef: secretName,
			Config: &target.KinesisOptions{
				AWSConfig:  target.AWSConfig{Endpoint: "endoint", Region: "region"},
				StreamName: "stream",
			},
		},
		SecurityHub: &target.Config[target.SecurityHubOptions]{
			SecretRef: secretName,
			Config: &target.SecurityHubOptions{
				AWSConfig: target.AWSConfig{Endpoint: "endoint", Region: "region"},
				AccountID: "accountID",
			},
		},
		GCS: &target.Config[target.GCSOptions]{
			SecretRef: secretName,
			Config: &target.GCSOptions{
				Bucket: "policy-reporter",
			},
		},
	}

	clients := factory.CreateClients(&targets)
	if len(clients.Clients()) != 12 {
		t.Fatalf("expected 12 clients created, got %d", len(clients.Clients()))
	}

	t.Run("Get Loki values from Secret", func(t *testing.T) {
		fv := reflect.ValueOf(clients.Client("Loki")).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/loki/api/v1/push" {
			t.Errorf("Expected host from secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Elasticsearch")).Elem()

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
		client := reflect.ValueOf(clients.Client("Slack")).Elem()

		webhook := client.FieldByName("channel").String()
		if webhook != "general" {
			t.Errorf("Expected channel from secret, got %s", webhook)
		}
	})

	t.Run("Get Discord values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Discord")).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get MS Teams values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Teams")).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get GoogleChat Webhook from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("GoogleChat")).Elem()

		host := client.FieldByName("webhook").String()
		if host != "http://localhost:9200/webhook" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Telegram Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Telegram")).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200/bottoken/sendMessage" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Webhook Authentication Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Webhook")).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
		}
	})

	t.Run("Get none existing secret skips target", func(t *testing.T) {
		clients := factory.CreateClients(&target.Targets{
			Loki: &target.Config[target.LokiOptions]{SecretRef: "not-exist"},
		})

		if len(clients.Clients()) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}

func Test_CustomFields(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := &target.Targets{
		Loki: &target.Config[target.LokiOptions]{
			Config: &target.LokiOptions{
				HostOptions: target.HostOptions{
					Host: "http://localhost:3100",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Elasticsearch: &target.Config[target.ElasticsearchOptions]{
			Config: &target.ElasticsearchOptions{
				HostOptions: target.HostOptions{
					Host: "http://localhost:9200",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Slack: &target.Config[target.SlackOptions]{
			Config: &target.SlackOptions{
				WebhookOptions: target.WebhookOptions{
					Webhook: "http://localhost:80",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Discord: &target.Config[target.WebhookOptions]{
			Config: &target.WebhookOptions{
				Webhook: "http://discord:80",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Teams: &target.Config[target.WebhookOptions]{
			Config: &target.WebhookOptions{
				Webhook: "http://hook.teams:80",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		GoogleChat: &target.Config[target.WebhookOptions]{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:900/webhook",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Telegram: &target.Config[target.TelegramOptions]{
			Config: &target.TelegramOptions{
				WebhookOptions: target.WebhookOptions{
					Webhook: "http://localhost:80",
				},
				Token:  "XXX",
				ChatID: "123456",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Webhook: &target.Config[target.WebhookOptions]{
			Config: &target.WebhookOptions{
				Webhook: "http://localhost:8080",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		S3: &target.Config[target.S3Options]{
			Config: &target.S3Options{
				AWSConfig: target.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				Bucket: "test",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Kinesis: &target.Config[target.KinesisOptions]{
			Config: &target.KinesisOptions{
				AWSConfig: target.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				StreamName: "policy-reporter",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		SecurityHub: &target.Config[target.SecurityHubOptions]{
			Config: &target.SecurityHubOptions{
				AWSConfig: target.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				AccountID: "AccountID",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		GCS: &target.Config[target.GCSOptions]{
			Config: &target.GCSOptions{
				Credentials: `{"token": "token", "type": "authorized_user"}`,
				Bucket:      "test",
				Prefix:      "prefix",
			},
			CustomFields: map[string]string{"field": "value"},
		},
	}

	clients := factory.CreateClients(targets)

	if len(clients.Clients()) != 12 {
		t.Fatalf("expected 12 client created, got %d", len(clients.Clients()))
	}

	t.Run("Get CustomFields from Loki", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Loki")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Elasticsearch", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Elasticsearch")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Slack", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Slack")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Discord", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Discord")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from MS Teams", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Teams")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from GoogleChat", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("GoogleChat")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Telegram", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Telegram")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})

	t.Run("Get CustomFields from Webhook", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Webhook")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from S3", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("S3")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from Kinesis", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Kinesis")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
	t.Run("Get CustomFields from GCS", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("GoogleCloudStorage")).Elem()

		customFields := client.FieldByName("customFields").MapKeys()
		if customFields[0].String() != "field" {
			t.Errorf("Expected customFields are added")
		}
	})
}

func Test_GetValuesFromMountedSecret(t *testing.T) {
	factory := factory.NewFactory(secrets.NewClient(newFakeClient()), nil)

	mountSecret()
	defer os.Remove(mountedSecret)

	targets := target.Targets{
		Loki:          &target.Config[target.LokiOptions]{MountedSecret: mountedSecret},
		Elasticsearch: &target.Config[target.ElasticsearchOptions]{MountedSecret: mountedSecret},
		Slack:         &target.Config[target.SlackOptions]{MountedSecret: mountedSecret},
		Discord:       &target.Config[target.WebhookOptions]{MountedSecret: mountedSecret},
		Teams:         &target.Config[target.WebhookOptions]{MountedSecret: mountedSecret},
		GoogleChat:    &target.Config[target.WebhookOptions]{MountedSecret: mountedSecret},
		Webhook:       &target.Config[target.WebhookOptions]{MountedSecret: mountedSecret},
		Telegram: &target.Config[target.TelegramOptions]{
			MountedSecret: mountedSecret,
			Config: &target.TelegramOptions{
				ChatID: "1234",
			},
		},
		S3: &target.Config[target.S3Options]{
			MountedSecret: mountedSecret,
			Config: &target.S3Options{
				AWSConfig: target.AWSConfig{Endpoint: "endoint", Region: "region"},
				Bucket:    "bucket",
			},
		},
		Kinesis: &target.Config[target.KinesisOptions]{
			MountedSecret: mountedSecret,
			Config: &target.KinesisOptions{
				AWSConfig:  target.AWSConfig{Endpoint: "endoint", Region: "region"},
				StreamName: "stream",
			},
		},
		SecurityHub: &target.Config[target.SecurityHubOptions]{
			MountedSecret: mountedSecret,
			Config: &target.SecurityHubOptions{
				AWSConfig: target.AWSConfig{Endpoint: "endoint", Region: "region"},
				AccountID: "accountID",
			},
		},
		GCS: &target.Config[target.GCSOptions]{
			MountedSecret: mountedSecret,
			Config: &target.GCSOptions{
				Bucket: "policy-reporter",
			},
		},
	}

	clients := factory.CreateClients(&targets)
	if len(clients.Clients()) != 12 {
		t.Fatalf("expected 12 client created, got %d", len(clients.Clients()))
	}

	t.Run("Get Loki values from Secret", func(t *testing.T) {
		fv := reflect.ValueOf(clients.Client("Loki")).Elem().FieldByName("host")
		if v := fv.String(); v != "http://localhost:9200/loki/api/v1/push" {
			t.Errorf("Expected host from secret, got %s", v)
		}
	})

	t.Run("Get Elasticsearch values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Elasticsearch")).Elem()

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
		client := reflect.ValueOf(clients.Client("Slack")).Elem()

		webhook := client.FieldByName("channel").String()
		if webhook != "general" {
			t.Errorf("Expected channel from secret, got %s", webhook)
		}
	})

	t.Run("Get Discord values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Discord")).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get MS Teams values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Teams")).Elem()

		webhook := client.FieldByName("webhook").String()
		if webhook != "http://localhost:9200/webhook" {
			t.Errorf("Expected webhook from secret, got %s", webhook)
		}
	})

	t.Run("Get GoogleChat Webhook from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("GoogleChat")).Elem()

		host := client.FieldByName("webhook").String()
		if host != "http://localhost:9200/webhook" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Telegram Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Telegram")).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200/bottoken/sendMessage" {
			t.Errorf("Expected host with token from secret, got %s", host)
		}
	})

	t.Run("Get Webhook Authentication Token from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Webhook")).Elem()

		token := client.FieldByName("headers").MapIndex(reflect.ValueOf("Authorization")).String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
		}
	})

	t.Run("Get none existing secret skips target", func(t *testing.T) {
		clients := factory.CreateClients(&target.Targets{
			Loki: &target.Config[target.LokiOptions]{SecretRef: "not-exist"},
		})

		if len(clients.Clients()) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}
