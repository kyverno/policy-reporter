package slack

import (
	"net/http"

	"github.com/slack-go/slack"
)

type APIClient interface {
	PostMessage(*slack.WebhookMessage) error
}

type apiClient struct {
	webhook string
	client  *http.Client
}

func (c *apiClient) PostMessage(message *slack.WebhookMessage) error {
	return slack.PostWebhookCustomHTTP(c.webhook, c.client, message)
}

func NewAPIClient(webhook string, client *http.Client) APIClient {
	return &apiClient{webhook: webhook, client: client}
}
