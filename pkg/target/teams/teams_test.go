package teams_test

import (
	"testing"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
)

type testClient struct {
	callback func(msg *adaptivecard.Message)
	send     bool
}

func (c *testClient) PostMessage(msg *adaptivecard.Message) error {
	c.callback(msg)
	c.send = true

	return nil
}

func Test_TeamsTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.Send(fixtures.CompleteTargetSendResult)

		assert.True(t, tc.send)
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.Send(fixtures.MinimalTargetSendResult)

		assert.True(t, tc.send)
	})
	t.Run("Send Minimal InfoResult", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.Send(fixtures.InfoSendResult)

		assert.True(t, tc.send)
	})
	t.Run("Send Minimal ErrorResult", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.Send(fixtures.ErrorSendResult)

		assert.True(t, tc.send)
	})
	t.Run("Send Minimal Debug Result", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.Send(fixtures.DebugSendResult)

		assert.True(t, tc.send)
	})
	t.Run("Send Scope Results", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.BatchSend(fixtures.ScopePolicyReport, fixtures.ScopePolicyReport.Results)

		assert.True(t, tc.send)
	})
	t.Run("Send Batch Results Without Scope", func(t *testing.T) {
		tc := &testClient{callback: func(msg *adaptivecard.Message) {
			if len(msg.Attachments) < 1 {
				t.Errorf("missing msg attachment")
			}
		}}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   tc,
		})
		client.BatchSend(fixtures.DefaultPolicyReport, fixtures.DefaultPolicyReport.Results)

		assert.True(t, tc.send)
	})
	t.Run("Name", func(t *testing.T) {
		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   &testClient{},
		})

		assert.Equal(t, "Teams", client.Name())
	})

	t.Run("Name", func(t *testing.T) {
		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   &testClient{},
		})

		assert.True(t, client.SupportsBatchSend())
	})
}
