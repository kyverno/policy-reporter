package secrets

import (
	"context"

	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/retry"
)

type Values struct {
	Host            string `json:"host,omitempty"`
	Webhook         string `json:"webhook,omitempty"`
	Channel         string `json:"channel,omitempty"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	APIKey          string `json:"apiKey,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	AccountID       string `json:"accountID,omitempty"`
	KmsKeyID        string `json:"kmsKeyId,omitempty"`
	Token           string `json:"token,omitempty"`
	Credentials     string `json:"credentials,omitempty"`
	Database        string `json:"database,omitempty"`
	DSN             string `json:"dsn,omitempty"`
	TypelessAPI     bool   `json:"typelessApi,omitempty"`
}

type Client interface {
	Get(context.Context, string) (Values, error)
}

type k8sClient struct {
	client v1.SecretInterface
}

func (c *k8sClient) Get(ctx context.Context, name string) (Values, error) {
	var secret *corev1.Secret

	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		if _, ok := err.(errors.APIStatus); !ok {
			return true
		}

		if ok := errors.IsTimeout(err); ok {
			return true
		}

		if ok := errors.IsServerTimeout(err); ok {
			return true
		}

		if ok := errors.IsServiceUnavailable(err); ok {
			return true
		}

		return false
	}, func() error {
		var err error
		secret, err = c.client.Get(ctx, name, metav1.GetOptions{})

		return err
	})

	values := Values{}
	if err != nil {
		return values, err
	}

	if host, ok := secret.Data["host"]; ok {
		values.Host = string(host)
	}

	if webhook, ok := secret.Data["webhook"]; ok {
		values.Webhook = string(webhook)
	}

	if channel, ok := secret.Data["channel"]; ok {
		values.Channel = string(channel)
	}

	if username, ok := secret.Data["username"]; ok {
		values.Username = string(username)
	}

	if password, ok := secret.Data["password"]; ok {
		values.Password = string(password)
	}

	if apiKey, ok := secret.Data["apiKey"]; ok {
		values.APIKey = string(apiKey)
	}

	if database, ok := secret.Data["database"]; ok {
		values.Database = string(database)
	}

	if dsn, ok := secret.Data["dsn"]; ok {
		values.DSN = string(dsn)
	}

	if accessKeyID, ok := secret.Data["accessKeyID"]; ok {
		values.AccessKeyID = string(accessKeyID)
	}

	if secretAccessKey, ok := secret.Data["secretAccessKey"]; ok {
		values.SecretAccessKey = string(secretAccessKey)
	}

	if kmsKeyID, ok := secret.Data["kmsKeyId"]; ok {
		values.KmsKeyID = string(kmsKeyID)
	}

	if accountID, ok := secret.Data["accountID"]; ok {
		values.AccountID = string(accountID)
	}

	if token, ok := secret.Data["token"]; ok {
		values.Token = string(token)
	}

	if credentials, ok := secret.Data["credentials"]; ok {
		values.Credentials = string(credentials)
	}

	if typelessAPI, ok := secret.Data["typelessApi"]; ok {
		values.TypelessAPI, err = strconv.ParseBool(string(typelessAPI))
		if err != nil {
			values.TypelessAPI = false
		}
	}

	return values, nil
}

func NewClient(secretClient v1.SecretInterface) Client {
	return &k8sClient{secretClient}
}
