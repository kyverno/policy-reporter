package secrets

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Values struct {
	Host            string `json:"host,omitempty"`
	Webhook         string `json:"webhook,omitempty"`
	Channel         string `json:"channel,omitempty"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	KmsKeyID        string `json:"kmsKeyId,omitempty"`
	Token           string `json:"token,omitempty"`
	Credentials     string `json:"credentials,omitempty"`
}

type Client interface {
	Get(context.Context, string) (Values, error)
}

type k8sClient struct {
	client v1.SecretInterface
}

func (c *k8sClient) Get(ctx context.Context, name string) (Values, error) {
	secret, err := c.client.Get(ctx, name, metav1.GetOptions{})
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

	if accessKeyID, ok := secret.Data["accessKeyID"]; ok {
		values.AccessKeyID = string(accessKeyID)
	}

	if secretAccessKey, ok := secret.Data["secretAccessKey"]; ok {
		values.SecretAccessKey = string(secretAccessKey)
	}

	if kmsKeyID, ok := secret.Data["kmsKeyId"]; ok {
		values.KmsKeyID = string(kmsKeyID)
	}

	if token, ok := secret.Data["token"]; ok {
		values.Token = string(token)
	}

	if credentials, ok := secret.Data["credentials"]; ok {
		values.Credentials = string(credentials)
	}

	return values, nil
}

func NewClient(secretClient v1.SecretInterface) Client {
	return &k8sClient{secretClient}
}
