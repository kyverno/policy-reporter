package slack_test

import (
	"testing"

	goslack "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
)

type testClient struct {
	callback   func(req *goslack.WebhookMessage)
	statusCode int
}

func (c testClient) PostMessage(req *goslack.WebhookMessage) error {
	c.callback(req)

	return nil
}

func Test_SlackTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 1, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(fixtures.CompleteTargetSendResult)
	})

	t.Run("Send Batch Results", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 3, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})

		client.BatchSend(fixtures.ScopePolicyReport, []v1alpha2.PolicyReportResult{
			fixtures.CompleteTargetSendResult,
			fixtures.CritcalSendResult,
		})
	})

	t.Run("Send Batch Results without scope", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 1, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})

		client.BatchSend(fixtures.EmptyPolicyReport, []v1alpha2.PolicyReportResult{
			fixtures.CompleteTargetSendResult,
			fixtures.CritcalSendResult,
		})
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 1, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			HTTPClient: testClient{callback, 200},
		})
		client.Send(fixtures.MinimalTargetSendResult)
	})

	t.Run("Send enforce Result", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 1, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			HTTPClient: testClient{callback, 200},
		})
		client.Send(fixtures.EnforceTargetSendResult)
	})

	t.Run("Send incomplete Result", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 1, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(fixtures.MissingUIDSendResult)
	})

	t.Run("Send incomplete Result2", func(t *testing.T) {
		callback := func(req *goslack.WebhookMessage) {
			assert.Equal(t, 1, len(req.Attachments))
		}

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(fixtures.MissingAPIVersionSendResult)
	})

	t.Run("Name", func(t *testing.T) {
		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{},
		})

		if client.Name() != "Slack" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
