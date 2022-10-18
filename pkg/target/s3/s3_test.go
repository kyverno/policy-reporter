package s3_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/s3"
)

var completeResult = report.Result{
	Message:    "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:     "require-requests-and-limits-required",
	Rule:       "autogen-check-for-requests-and-limits",
	Priority:   report.WarningPriority,
	Status:     report.Fail,
	Severity:   report.High,
	Category:   "resources",
	Scored:     true,
	Properties: map[string]string{"version": "1234"},
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

func Test_S3Target(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(body *bytes.Buffer, key string) {
			report := new(bytes.Buffer)
			json.NewEncoder(report).Encode(completeResult)

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
		client.Send(completeResult)

		if len(completeResult.Properties) > 1 || completeResult.Properties["cluster"] != "" {
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
