package splunk

import (
	"net/http"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type testClient struct {
	callBack   func(req *http.Request) error
	statusCode int
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	err := c.callBack(req)
	return &http.Response{
		StatusCode: c.statusCode,
	}, err
}

func TestSplunkTarget(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(req *http.Request) error {
			if agent := req.Header.Get("User-Agent"); agent != "Policy-Reporter" {
				t.Errorf("Unexpected Agent: %s", agent)
			}

			if url := req.URL.String(); url != "http://localhost:8088/services/collector" {
				t.Errorf("Unexpected Host: %s", url)
			}

			if value := req.Header.Get("Authorization"); value != "Splunk my-token" {
				t.Errorf("Unexpected Header Authorization: %s", value)
			}

			return nil
		}

		client := NewClient(Options{
			ClientOptions: target.ClientOptions{
				Name: "Test",
			},
			Host:       "http://localhost:8088/services/collector",
			Token:      "my-token",
			Headers:    map[string]string{"Authorization": "Splunk my-token"},
			HTTPClient: testClient{callback, 200},
		})
		client.Send(fixtures.DefaultPolicyReport, fixtures.CompleteTargetSendResult)
	})
}
