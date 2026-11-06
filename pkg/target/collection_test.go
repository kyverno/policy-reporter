package target_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

func TestCollection(t *testing.T) {
	t.Parallel()
	collection := target.NewCollection(
		&target.Target{
			ID:   uuid.NewString(),
			Type: target.Webhook,
			Client: webhook.NewClient(webhook.Options{
				ClientOptions: target.ClientOptions{
					Name: "Webhook",
				},
			}),
			Config:       &targetconfig.Config[v1alpha1.WebhookOptions]{SecretRef: "webhook-secret"},
			ParentConfig: &targetconfig.Config[v1alpha1.WebhookOptions]{},
		},
		&target.Target{
			ID:   uuid.NewString(),
			Type: target.Slack,
			Client: slack.NewClient(slack.Options{
				ClientOptions: target.ClientOptions{
					Name: "Slack",
				},
			}),
			Config:       &targetconfig.Config[v1alpha1.SlackOptions]{},
			ParentConfig: &targetconfig.Config[v1alpha1.SlackOptions]{SecretRef: "slack-secret"},
		},
		&target.Target{
			ID:   uuid.NewString(),
			Type: target.Discord,
			Client: discord.NewClient(discord.Options{
				ClientOptions: target.ClientOptions{
					Name: "Discord",
				},
			}),
			Config:       &targetconfig.Config[v1alpha1.WebhookOptions]{},
			ParentConfig: &targetconfig.Config[v1alpha1.WebhookOptions]{SecretRef: "slack-secret"},
		},
	)

	t.Run("empty returns if the collection has any target", func(t *testing.T) {
		t.Parallel()
		assert.True(t, target.NewCollection().Empty())
		assert.False(t, collection.Empty())
	})

	t.Run("length returns the amount of targets within a collection", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, collection.Length(), 3)
	})

	t.Run("clients returns all clients of the given targets", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, len(collection.Clients()), 3)
	})

	t.Run("client searches for a configured target with the given name", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, collection.Client("Webhook"))
		assert.NotNil(t, collection.Client("Discord"))
		assert.NotNil(t, collection.Client("Slack"))
		assert.Nil(t, collection.Client("Invalid"))
	})

	t.Run("usesSecret checks if at least on target has a secretRef configured", func(t *testing.T) {
		t.Parallel()
		assert.False(t, target.NewCollection().UsesSecrets())
		assert.True(t, collection.UsesSecrets())
	})

	t.Run("SingleSendClients only returns clients which do not support batch sending", func(t *testing.T) {
		t.Parallel()
		for _, c := range collection.SingleSendClients() {
			assert.Equal(t, target.SingleSend, c.Type())
		}
	})

	t.Run("BatchSendClients only returns clients which do support batch sending", func(t *testing.T) {
		t.Parallel()
		for _, c := range collection.BatchSendClients() {
			assert.Equal(t, target.BatchSend, c.Type())
		}
	})
}

func TestCollectionUpdate(t *testing.T) {
	t.Parallel()

	newTarget := func(id, name string) *target.Target {
		return &target.Target{
			ID:   id,
			Type: target.Slack,
			Client: slack.NewClient(slack.Options{
				ClientOptions: target.ClientOptions{Name: name},
			}),
			Config:       &targetconfig.Config[v1alpha1.SlackOptions]{Name: name},
			ParentConfig: &targetconfig.Config[v1alpha1.SlackOptions]{},
		}
	}

	t.Run("replaces a target added with AddTarget instead of adding a second entry", func(t *testing.T) {
		t.Parallel()
		collection := target.NewCollection()

		original := newTarget(uuid.NewString(), "original")
		collection.AddTarget("target-config", original)

		updated := newTarget(original.ID, "updated")
		collection.Update(updated)

		assert.Equal(t, 1, collection.Length())
		assert.Len(t, collection.Clients(), 1)
		assert.Equal(t, "updated", collection.Targets()[0].Client.Name())
	})

	t.Run("keeps replacing the same entry on repeated updates", func(t *testing.T) {
		t.Parallel()
		collection := target.NewCollection()

		original := newTarget(uuid.NewString(), "original")
		collection.AddTarget("target-config", original)

		collection.Update(newTarget(original.ID, "first"))
		collection.Update(newTarget(original.ID, "second"))

		assert.Equal(t, 1, collection.Length())
		assert.Equal(t, "second", collection.Targets()[0].Client.Name())
	})

	t.Run("adds the target under its ID if it does not exist yet", func(t *testing.T) {
		t.Parallel()
		collection := target.NewCollection()

		collection.Update(newTarget(uuid.NewString(), "new"))

		assert.Equal(t, 1, collection.Length())
	})

	t.Run("does not touch other targets", func(t *testing.T) {
		t.Parallel()
		collection := target.NewCollection()

		first := newTarget(uuid.NewString(), "first")
		second := newTarget(uuid.NewString(), "second")
		collection.AddTarget("first-config", first)
		collection.AddTarget("second-config", second)

		collection.Update(newTarget(first.ID, "first-updated"))

		assert.Equal(t, 2, collection.Length())
		assert.NotNil(t, collection.Client("first-updated"))
		assert.NotNil(t, collection.Client("second"))
		assert.Nil(t, collection.Client("first"))
	})
}
