package loki_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
)

type testClient struct {
	callback   func(req *http.Request)
	statusCode int
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	c.callback(req)

	return &http.Response{
		StatusCode: c.statusCode,
	}, nil
}

func Test_LokiTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "unexpected Content-Type")
			assert.Equal(t, "http://loki", req.Header.Get("X-Forward"), "unexpected X-Forward")
			assert.Equal(t, "Policy-Reporter", req.Header.Get("User-Agent"), "unexpected Agent")
			assert.Equal(t, "http://localhost:3100/loki/api/v1/push", req.URL.String(), "unexpected Host")
			assert.NotEqual(t, "", req.Header.Get("Authorization"), "unexpected auth header")

			expectedLine := fmt.Sprintf("[%s] %s", strings.ToUpper(string(fixtures.CompleteTargetSendResult.Severity)), fixtures.CompleteTargetSendResult.Description)

			stream := convertAndValidateBody(req, t)

			assert.Equal(t, expectedLine, stream.Values[0][1])
			assert.Equal(t, fixtures.CompleteTargetSendResult.Rule, stream.Stream["rule"])
			assert.Equal(t, fixtures.CompleteTargetSendResult.Policy, stream.Stream["policy"])
			assert.Equal(t, fixtures.CompleteTargetSendResult.Category, stream.Stream["category"])
			assert.Equal(t, string(fixtures.CompleteTargetSendResult.Result), stream.Stream["status"])
			assert.Equal(t, string(fixtures.CompleteTargetSendResult.Severity), stream.Stream["severity"])
			or := fixtures.CompleteTargetSendResult
			res := or.GetResource()
			assert.Equal(t, res.Kind, stream.Stream["kind"])
			assert.Equal(t, res.Name, stream.Stream["name"])
			assert.Equal(t, string(res.UID), stream.Stream["uid"])
			assert.Equal(t, res.Namespace, stream.Stream["namespace"])

			assert.Equal(t, fixtures.CompleteTargetSendResult.Properties["version"], stream.Stream["version"])
		}

		client := loki.NewClient(loki.Options{
			ClientOptions: target.ClientOptions{
				Name: "Loki",
			},
			Host:         "http://localhost:3100/loki/api/v1/push",
			CustomFields: map[string]string{"custom": "label"},
			HTTPClient:   testClient{callback, 200},
			Username:     "username",
			Password:     "password",
			Headers:      map[string]string{"X-Forward": "http://loki"},
		})
		client.Send(fixtures.DefaultPolicyReport, fixtures.CompleteTargetSendResult)
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "unexpected Content-Type")
			assert.Equal(t, "Policy-Reporter", req.Header.Get("User-Agent"), "unexpected Agent")
			assert.Equal(t, "http://localhost:3100/loki/api/v1/push", req.URL.String(), "unexpected Host")

			expectedLine := fmt.Sprintf("[%s] %s", strings.ToUpper(string(fixtures.MinimalTargetSendResult.Severity)), fixtures.MinimalTargetSendResult.Description)
			stream := convertAndValidateBody(req, t)

			assert.Equal(t, expectedLine, stream.Values[0][1])
			assert.Equal(t, fixtures.MinimalTargetSendResult.Rule, stream.Stream["rule"])
			assert.Equal(t, fixtures.MinimalTargetSendResult.Policy, stream.Stream["policy"])
			assert.Equal(t, fixtures.MinimalTargetSendResult.Category, stream.Stream["category"])
			assert.Equal(t, string(fixtures.MinimalTargetSendResult.Result), stream.Stream["status"])
			assert.Equal(t, string(fixtures.MinimalTargetSendResult.Severity), stream.Stream["severity"])

			assert.Equal(t, "policy-reporter", stream.Stream["createdBy"])
		}

		client := loki.NewClient(loki.Options{
			ClientOptions: target.ClientOptions{
				Name: "Loki",
			},
			Host:         "http://localhost:3100/loki/api/v1/push",
			CustomFields: map[string]string{"custom": "label"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(fixtures.DefaultPolicyReport, fixtures.MinimalTargetSendResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := loki.NewClient(loki.Options{
			ClientOptions: target.ClientOptions{
				Name: "Loki",
			},
			Host:         "http://localhost:3100/loki/api/v1/push",
			CustomFields: map[string]string{"custom": "label"},
			HTTPClient:   testClient{},
		})

		if client.Name() != "Loki" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}

func convertAndValidateBody(req *http.Request, t *testing.T) loki.Stream {
	payload := loki.Payload{}

	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, payload.Streams[0].Values, 1)
	assert.Len(t, payload.Streams[0].Values[0], 2)

	return payload.Streams[0]
}
