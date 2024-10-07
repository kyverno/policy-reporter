package target_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

func TestCollection(t *testing.T) {
	collection := target.NewCollection(
		&target.Target{
			ID:   uuid.NewString(),
			Type: target.Webhook,
			Client: webhook.NewClient(webhook.Options{
				ClientOptions: target.ClientOptions{
					Name: "Webhook",
				},
			}),
			Config:       &target.Config[target.WebhookOptions]{SecretRef: "webhook-secret"},
			ParentConfig: &target.Config[target.WebhookOptions]{},
		},
		&target.Target{
			ID:   uuid.NewString(),
			Type: target.Slack,
			Client: slack.NewClient(slack.Options{
				ClientOptions: target.ClientOptions{
					Name: "Slack",
				},
			}),
			Config:       &target.Config[target.SlackOptions]{},
			ParentConfig: &target.Config[target.SlackOptions]{SecretRef: "slack-secret"},
		},
		&target.Target{
			ID:   uuid.NewString(),
			Type: target.Discord,
			Client: discord.NewClient(discord.Options{
				ClientOptions: target.ClientOptions{
					Name: "Discord",
				},
			}),
			Config:       &target.Config[target.WebhookOptions]{},
			ParentConfig: &target.Config[target.WebhookOptions]{SecretRef: "slack-secret"},
		},
	)

	t.Run("empty returns if the collection has any target", func(t *testing.T) {
		assert.True(t, target.NewCollection().Empty())
		assert.False(t, collection.Empty())
	})

	t.Run("length returns the amount of targets within a collection", func(t *testing.T) {
		assert.Equal(t, collection.Length(), 3)
	})

	t.Run("clients returns all clients of the given targets", func(t *testing.T) {
		assert.Equal(t, len(collection.Clients()), 3)
	})

	t.Run("client searches for a configured target with the given name", func(t *testing.T) {
		assert.NotNil(t, collection.Client("Webhook"))
		assert.NotNil(t, collection.Client("Discord"))
		assert.NotNil(t, collection.Client("Slack"))
		assert.Nil(t, collection.Client("Invalid"))
	})

	t.Run("usesSecret checks if at least on target has a secretRef configured", func(t *testing.T) {
		assert.False(t, target.NewCollection().UsesSecrets())
		assert.True(t, collection.UsesSecrets())
	})

	t.Run("SingleSendClients only returns clients which do not support batch sending", func(t *testing.T) {
		for _, c := range collection.SingleSendClients() {
			assert.Equal(t, target.SingleSend, c.Type())
		}
	})

	t.Run("BatchSendClients only returns clients which do support batch sending", func(t *testing.T) {
		for _, c := range collection.BatchSendClients() {
			assert.Equal(t, target.BatchSend, c.Type())
		}
	})
}
