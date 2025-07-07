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

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
	"github.com/kyverno/policy-reporter/pkg/openreports"
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
			"accountId":       []byte("accountId"),
			"typelessApi":     []byte("true"),
			"accessKeyId":     []byte("accessKeyId"),
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
		AccountID:       "accountId",
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
	Loki: &targetconfig.Config[v1alpha1.LokiOptions]{
		Config: &v1alpha1.LokiOptions{
			HostOptions: v1alpha1.HostOptions{
				Host:    "http://localhost:3100",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.LokiOptions]{
			{
				CustomFields: map[string]string{"label2": "value2"},
			},
		},
	},
	Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{
		Config: &v1alpha1.ElasticsearchOptions{
			HostOptions: v1alpha1.HostOptions{
				Host:    "http://localhost:9200",
				SkipTLS: true,
			},
			Index:    "policy-reporter",
			Rotation: "daily",
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.ElasticsearchOptions]{{}},
	},
	Slack: &targetconfig.Config[v1alpha1.SlackOptions]{
		Config: &v1alpha1.SlackOptions{
			WebhookOptions: v1alpha1.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.SlackOptions]{{
			Config: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:9200",
				},
			},
		}, {
			Config: &v1alpha1.SlackOptions{
				Channel: "general",
			},
		}},
	},
	Discord: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://discord:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.WebhookOptions]{{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	Teams: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://hook.teams:80",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.WebhookOptions]{{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:9200",
			},
		}},
	},
	GoogleChat: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://localhost:900/webhook",
			SkipTLS: true,
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.WebhookOptions]{{}},
	},
	Telegram: &targetconfig.Config[v1alpha1.TelegramOptions]{
		Config: &v1alpha1.TelegramOptions{
			WebhookOptions: v1alpha1.WebhookOptions{
				Webhook: "http://localhost:80",
				SkipTLS: true,
			},
			Token:  "XXX",
			ChatID: "123456",
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.TelegramOptions]{{
			Config: &v1alpha1.TelegramOptions{
				ChatID: "1234567",
			},
		}},
	},
	Webhook: &targetconfig.Config[v1alpha1.WebhookOptions]{
		Config: &v1alpha1.WebhookOptions{
			Webhook: "http://localhost:8080",
			SkipTLS: true,
			Headers: map[string]string{
				"X-Custom": "Header",
			},
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels: []*targetconfig.Config[v1alpha1.WebhookOptions]{{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:8081",
				Headers: map[string]string{
					"X-Custom-2": "Header",
				},
			},
		}},
	},
	S3: &targetconfig.Config[v1alpha1.S3Options]{
		Config: &v1alpha1.S3Options{
			AWSConfig: v1alpha1.AWSConfig{
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
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.S3Options]{{}},
	},
	Kinesis: &targetconfig.Config[v1alpha1.KinesisOptions]{
		Config: &v1alpha1.KinesisOptions{
			AWSConfig: v1alpha1.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			StreamName: "policy-reporter",
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.KinesisOptions]{{}},
	},
	SecurityHub: &targetconfig.Config[v1alpha1.SecurityHubOptions]{
		Config: &v1alpha1.SecurityHubOptions{
			AWSConfig: v1alpha1.AWSConfig{
				AccessKeyID:     "AccessKey",
				SecretAccessKey: "SecretAccessKey",
				Endpoint:        "https://storage.yandexcloud.net",
				Region:          "ru-central1",
			},
			AccountID: "AccountID",
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.SecurityHubOptions]{{}},
	},
	GCS: &targetconfig.Config[v1alpha1.GCSOptions]{
		Config: &v1alpha1.GCSOptions{
			Credentials: `{"token": "token", "type": "authorized_user"}`,
			Bucket:      "test",
			Prefix:      "prefix",
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
		Channels:        []*targetconfig.Config[v1alpha1.GCSOptions]{{}},
	},
	Splunk: &targetconfig.Config[v1alpha1.SplunkOptions]{
		Config: &v1alpha1.SplunkOptions{
			HostOptions: v1alpha1.HostOptions{
				Host: "http://localhost:9200",
			},
			Token: "token",
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
	},
	Jira: &targetconfig.Config[v1alpha1.JiraOptions]{
		Config: &v1alpha1.JiraOptions{
			ProjectKey: "PR",
			Host:       "http://localhost:9200",
			APIToken:   "token",
			APIVersion: "v2",
			IssueType:  "Bug",
			Username:   "username",
			Labels:     []string{"dev-cluster"},
			Components: []string{"component1"},
		},
		SkipExisting:    true,
		MinimumSeverity: openreports.SeverityInfo,
		CustomFields:    map[string]string{"field": "value"},
	},
}

func Test_ResolveTarget(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	clients := factory.CreateClients(&targets)
	if len(clients.Clients()) != 26 {
		t.Errorf("Expected 26 Client, got %d clients", len(clients.Clients()))
	}
}

func Test_ResolveTargetsWithoutRequiredConfiguration(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		Loki:          &targetconfig.Config[v1alpha1.LokiOptions]{},
		Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{},
		Slack:         &targetconfig.Config[v1alpha1.SlackOptions]{},
		Discord:       &targetconfig.Config[v1alpha1.WebhookOptions]{},
		Teams:         &targetconfig.Config[v1alpha1.WebhookOptions]{},
		GoogleChat:    &targetconfig.Config[v1alpha1.WebhookOptions]{},
		Webhook:       &targetconfig.Config[v1alpha1.WebhookOptions]{},
		Telegram:      &targetconfig.Config[v1alpha1.TelegramOptions]{},
		S3:            &targetconfig.Config[v1alpha1.S3Options]{},
		Kinesis:       &targetconfig.Config[v1alpha1.KinesisOptions]{},
		SecurityHub:   &targetconfig.Config[v1alpha1.SecurityHubOptions]{},
		Jira:          &targetconfig.Config[v1alpha1.JiraOptions]{},
	}

	if len(factory.CreateClients(&targets).Clients()) != 0 {
		t.Error("Expected Client to be nil if no required fields are configured")
	}

	targets = target.Targets{}
	if len(factory.CreateClients(&targets).Clients()) != 0 {
		t.Error("Expected Client to be nil if no target is configured")
	}

	targets.S3 = &targetconfig.Config[v1alpha1.S3Options]{
		Config: &v1alpha1.S3Options{
			AWSConfig: v1alpha1.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
		},
	}
}

func Test_S3Validation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		S3: &targetconfig.Config[v1alpha1.S3Options]{
			Config: &v1alpha1.S3Options{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("S3.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.S3.Config.AccessKeyID = "access"
	t.Run("S3.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.S3.Config.SecretAccessKey = "secret"
	t.Run("S3.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})

	targets.S3.Config.Region = "ru-central1"
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
		Kinesis: &targetconfig.Config[v1alpha1.KinesisOptions]{
			Config: &v1alpha1.KinesisOptions{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("Kinesis.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.Kinesis.Config.AccessKeyID = "access"
	t.Run("Kinesis.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.Kinesis.Config.SecretAccessKey = "secret"

	t.Run("Kinesis.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})

	targets.Kinesis.Config.Region = "ru-central1"

	t.Run("Kinesis.StreamName", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no stream name is configured")
		}
	})
}

func Test_SecurityHubValidation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		SecurityHub: &targetconfig.Config[v1alpha1.SecurityHubOptions]{
			Config: &v1alpha1.SecurityHubOptions{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "https://storage.yandexcloud.net"},
			},
		},
	}

	t.Run("SecurityHub.AccountId", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accountId is configured")
		}
	})

	targets.SecurityHub.Config.AccountID = "accountId"
	t.Run("SecurityHub.AccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no accessKey is configured")
		}
	})

	targets.SecurityHub.Config.AccessKeyID = "access"
	t.Run("SecurityHub.SecretAccessKey", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no secretAccessKey is configured")
		}
	})

	targets.SecurityHub.Config.SecretAccessKey = "secret"
	t.Run("SecurityHub.Region", func(t *testing.T) {
		if len(factory.CreateClients(&targets).Clients()) != 0 {
			t.Error("Expected Client to be nil if no region is configured")
		}
	})
}

func Test_GCSValidation(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := target.Targets{
		GCS: &targetconfig.Config[v1alpha1.GCSOptions]{
			Config: &v1alpha1.GCSOptions{
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
		Loki:          &targetconfig.Config[v1alpha1.LokiOptions]{SecretRef: secretName},
		Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{SecretRef: secretName},
		Slack:         &targetconfig.Config[v1alpha1.SlackOptions]{SecretRef: secretName},
		Discord:       &targetconfig.Config[v1alpha1.WebhookOptions]{SecretRef: secretName},
		Teams:         &targetconfig.Config[v1alpha1.WebhookOptions]{SecretRef: secretName},
		GoogleChat:    &targetconfig.Config[v1alpha1.WebhookOptions]{SecretRef: secretName},
		Webhook:       &targetconfig.Config[v1alpha1.WebhookOptions]{SecretRef: secretName},
		Telegram: &targetconfig.Config[v1alpha1.TelegramOptions]{
			SecretRef: secretName,
			Config: &v1alpha1.TelegramOptions{
				ChatID: "1234",
			},
		},
		S3: &targetconfig.Config[v1alpha1.S3Options]{
			SecretRef: secretName,
			Config: &v1alpha1.S3Options{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "endpoint", Region: "region"},
				Bucket:    "bucket",
			},
		},
		Kinesis: &targetconfig.Config[v1alpha1.KinesisOptions]{
			SecretRef: secretName,
			Config: &v1alpha1.KinesisOptions{
				AWSConfig:  v1alpha1.AWSConfig{Endpoint: "endpoint", Region: "region"},
				StreamName: "stream",
			},
		},
		SecurityHub: &targetconfig.Config[v1alpha1.SecurityHubOptions]{
			SecretRef: secretName,
			Config: &v1alpha1.SecurityHubOptions{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "endpoint", Region: "region"},
				AccountID: "accountId",
			},
		},
		GCS: &targetconfig.Config[v1alpha1.GCSOptions]{
			SecretRef: secretName,
			Config: &v1alpha1.GCSOptions{
				Bucket: "policy-reporter",
			},
		},
		Splunk: &targetconfig.Config[v1alpha1.SplunkOptions]{
			SecretRef: secretName,
		},
	}

	clients := factory.CreateClients(&targets)
	if len(clients.Clients()) != 13 {
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

	t.Run("Get Splunk values from Secret", func(t *testing.T) {
		client := reflect.ValueOf(clients.Client("Splunk")).Elem()

		host := client.FieldByName("host").String()
		if host != "http://localhost:9200" {
			t.Errorf("Expected host from secret, got %s", host)
		}

		token := client.FieldByName("token").String()
		if token != "token" {
			t.Errorf("Expected token from secret, got %s", token)
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
			Loki: &targetconfig.Config[v1alpha1.LokiOptions]{SecretRef: "not-exist"},
		})

		if len(clients.Clients()) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}

func Test_CustomFields(t *testing.T) {
	factory := factory.NewFactory(nil, nil)

	targets := &target.Targets{
		Loki: &targetconfig.Config[v1alpha1.LokiOptions]{
			Config: &v1alpha1.LokiOptions{
				HostOptions: v1alpha1.HostOptions{
					Host: "http://localhost:3100",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{
			Config: &v1alpha1.ElasticsearchOptions{
				HostOptions: v1alpha1.HostOptions{
					Host: "http://localhost:9200",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Slack: &targetconfig.Config[v1alpha1.SlackOptions]{
			Config: &v1alpha1.SlackOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:80",
				},
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Discord: &targetconfig.Config[v1alpha1.WebhookOptions]{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://discord:80",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Teams: &targetconfig.Config[v1alpha1.WebhookOptions]{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://hook.teams:80",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		GoogleChat: &targetconfig.Config[v1alpha1.WebhookOptions]{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:900/webhook",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Telegram: &targetconfig.Config[v1alpha1.TelegramOptions]{
			Config: &v1alpha1.TelegramOptions{
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://localhost:80",
				},
				Token:  "XXX",
				ChatID: "123456",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Webhook: &targetconfig.Config[v1alpha1.WebhookOptions]{
			Config: &v1alpha1.WebhookOptions{
				Webhook: "http://localhost:8080",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		S3: &targetconfig.Config[v1alpha1.S3Options]{
			Config: &v1alpha1.S3Options{
				AWSConfig: v1alpha1.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				Bucket: "test",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		Kinesis: &targetconfig.Config[v1alpha1.KinesisOptions]{
			Config: &v1alpha1.KinesisOptions{
				AWSConfig: v1alpha1.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				StreamName: "policy-reporter",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		SecurityHub: &targetconfig.Config[v1alpha1.SecurityHubOptions]{
			Config: &v1alpha1.SecurityHubOptions{
				AWSConfig: v1alpha1.AWSConfig{
					AccessKeyID:     "AccessKey",
					SecretAccessKey: "SecretAccessKey",
					Endpoint:        "https://storage.yandexcloud.net",
					Region:          "ru-central1",
				},
				AccountID: "AccountId",
			},
			CustomFields: map[string]string{"field": "value"},
		},
		GCS: &targetconfig.Config[v1alpha1.GCSOptions]{
			Config: &v1alpha1.GCSOptions{
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
		Loki:          &targetconfig.Config[v1alpha1.LokiOptions]{MountedSecret: mountedSecret},
		Elasticsearch: &targetconfig.Config[v1alpha1.ElasticsearchOptions]{MountedSecret: mountedSecret},
		Slack:         &targetconfig.Config[v1alpha1.SlackOptions]{MountedSecret: mountedSecret},
		Discord:       &targetconfig.Config[v1alpha1.WebhookOptions]{MountedSecret: mountedSecret},
		Teams:         &targetconfig.Config[v1alpha1.WebhookOptions]{MountedSecret: mountedSecret},
		GoogleChat:    &targetconfig.Config[v1alpha1.WebhookOptions]{MountedSecret: mountedSecret},
		Webhook:       &targetconfig.Config[v1alpha1.WebhookOptions]{MountedSecret: mountedSecret},
		Telegram: &targetconfig.Config[v1alpha1.TelegramOptions]{
			MountedSecret: mountedSecret,
			Config: &v1alpha1.TelegramOptions{
				ChatID: "1234",
			},
		},
		S3: &targetconfig.Config[v1alpha1.S3Options]{
			MountedSecret: mountedSecret,
			Config: &v1alpha1.S3Options{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "endpoint", Region: "region"},
				Bucket:    "bucket",
			},
		},
		Kinesis: &targetconfig.Config[v1alpha1.KinesisOptions]{
			MountedSecret: mountedSecret,
			Config: &v1alpha1.KinesisOptions{
				AWSConfig:  v1alpha1.AWSConfig{Endpoint: "endpoint", Region: "region"},
				StreamName: "stream",
			},
		},
		SecurityHub: &targetconfig.Config[v1alpha1.SecurityHubOptions]{
			MountedSecret: mountedSecret,
			Config: &v1alpha1.SecurityHubOptions{
				AWSConfig: v1alpha1.AWSConfig{Endpoint: "endpoint", Region: "region"},
				AccountID: "accountId",
			},
		},
		GCS: &targetconfig.Config[v1alpha1.GCSOptions]{
			MountedSecret: mountedSecret,
			Config: &v1alpha1.GCSOptions{
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
			Loki: &targetconfig.Config[v1alpha1.LokiOptions]{SecretRef: "not-exist"},
		})

		if len(clients.Clients()) != 0 {
			t.Error("Expected client are skipped")
		}
	})
}
