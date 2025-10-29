package kinesis_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/kinesis"
)

type testClient struct {
	err      error
	callback func(body *bytes.Buffer, key string)
}

func (c *testClient) Upload(_ *bytes.Buffer, _ string) error {
	return c.err
}

var testCallback = func(body *bytes.Buffer, key string) {}

func Test_KinesisTarget(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(body *bytes.Buffer, key string) {
			report := new(bytes.Buffer)
			if err := json.NewEncoder(report).Encode(fixtures.CompleteTargetSendResult); err != nil {
				t.Errorf("Failed to encode report message: %s", err)
			}

			if body != report {
				buf := new(bytes.Buffer)
				if _, err := buf.ReadFrom(body); err != nil {
					t.Errorf("Failed to read from body: %s", err)
				}

				t.Errorf("Unexpected Body Content: %s", buf.String())
			}
		}

		client := kinesis.NewClient(kinesis.Options{
			ClientOptions: target.ClientOptions{
				Name: "Kinesis",
			},
			CustomFields: map[string]string{"cluster": "name"},
			Kinesis:      &testClient{nil, callback},
		})
		client.Send(fixtures.DefaultPolicyReport, fixtures.CompleteTargetSendResult)

		if len(fixtures.CompleteTargetSendResult.Properties) > 1 || fixtures.CompleteTargetSendResult.Properties["cluster"] != "" {
			t.Error("expected customFields are not added to the actuel result")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := kinesis.NewClient(kinesis.Options{
			ClientOptions: target.ClientOptions{
				Name: "Kinesis",
			},
			Kinesis: &testClient{},
		})

		if client.Name() != "Kinesis" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
