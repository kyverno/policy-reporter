package kinesis_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/kinesis"
)

var completeResult = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

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
			if err := json.NewEncoder(report).Encode(completeResult); err != nil {
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

		client := kinesis.NewClient("Kinesis", &testClient{nil, callback}, true, &target.Filter{})
		client.Send(completeResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := kinesis.NewClient("Kinesis", &testClient{nil, testCallback}, false, &target.Filter{})

		if client.Name() != "Kinesis" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
