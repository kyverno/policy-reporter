package googlechat_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/googlechat"
)

type testClient struct {
	callback   func(req *http.Request) error
	statusCode int
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	err := c.callback(req)

	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}, err
}

func Test_GoogleChatTarget(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(req *http.Request) error {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "https://googlechat.webhook" {
				t.Errorf("Unexpected Host: %s", url)
			}

			if value := req.Header.Get("X-Code"); value != "1234" {
				t.Errorf("Unexpected Header X-Code: %s", value)
			}

			return nil
		}

		client := googlechat.NewClient(googlechat.Options{
			ClientOptions: target.ClientOptions{
				Name: "GoogleChat",
			},
			Webhook:      "https://googlechat.webhook",
			Headers:      map[string]string{"X-Code": "1234"},
			CustomFields: map[string]string{"cluster": "name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(fixtures.DefaultPolicyReport, fixtures.CompleteTargetSendResult)

		if len(fixtures.CompleteTargetSendResult.Properties) > 1 || fixtures.CompleteTargetSendResult.Properties["cluster"] != "" {
			t.Error("expected customFields are not added to the actuel result")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := googlechat.NewClient(googlechat.Options{
			ClientOptions: target.ClientOptions{
				Name: "GoogleChat",
			},
			Webhook:    "https://googlechat.webhook",
			Headers:    map[string]string{"X-Code": "1234"},
			HTTPClient: testClient{},
		})

		if client.Name() != "GoogleChat" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
