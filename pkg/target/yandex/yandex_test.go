package yandex_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target/yandex"
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

func Test_YandexTarget(t *testing.T) {
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

		client := yandex.NewClient(&testClient{nil, callback}, "", "", true)
		client.Send(completeResult)
	})
	t.Run("Send with ignored Priority", func(t *testing.T) {
		callback := func(body *bytes.Buffer, key string) {
			t.Errorf("Unexpected Call")
		}

		client := yandex.NewClient(&testClient{nil, callback}, "", "error", true)
		client.Send(completeResult)
	})
	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		client := yandex.NewClient(&testClient{nil, testCallback}, "", "", true)

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := yandex.NewClient(&testClient{nil, testCallback}, "", "", false)

		if client.Name() != "Yandex" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := yandex.NewClient(&testClient{nil, testCallback}, "", "debug", false)

		if client.MinimumPriority() != "debug" {
			t.Errorf("Unexpected MinimumPriority %s", client.MinimumPriority())
		}
	})
}
