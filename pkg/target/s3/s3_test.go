package s3_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
)

type testClient struct {
	err      error
	callback func(body *bytes.Buffer, key string)
}

func (c *testClient) Upload(_ *bytes.Buffer, _ string) error {
	return c.err
}

var testCallback = func(body *bytes.Buffer, key string) {}

func Test_S3Target(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(body *bytes.Buffer, key string) {
			report := new(bytes.Buffer)
			json.NewEncoder(report).Encode(fixtures.CompleteTargetSendResult)

			if body != report {
				buf := new(bytes.Buffer)
				buf.ReadFrom(body)

				t.Errorf("Unexpected Body Content: %s", buf.String())
			}
		}

		client := s3.NewClient(s3.Options{
			ClientOptions: target.ClientOptions{
				Name: "S3",
			},
			CustomFields: map[string]string{"cluster": "name"},
			S3:           &testClient{nil, callback},
		})
		client.Send(fixtures.CompleteTargetSendResult)

		if len(fixtures.CompleteTargetSendResult.Properties) > 1 || fixtures.CompleteTargetSendResult.Properties["cluster"] != "" {
			t.Error("expected customFields are not added to the actuel result")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := s3.NewClient(s3.Options{
			ClientOptions: target.ClientOptions{
				Name: "S3",
			},
			S3: &testClient{},
		})

		if client.Name() != "S3" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
