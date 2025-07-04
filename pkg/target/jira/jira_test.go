package jira_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/jira"
)

type testClient struct {
	callback   func(req *http.Request)
	statusCode int
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	c.callback(req)

	return &http.Response{
		StatusCode: c.statusCode,
		Request:    req,
		Body:       io.NopCloser(bytes.NewBufferString(``)),
	}, nil
}

func Test_JiraTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			// Verify HTTP headers and method
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

			// Verify URL
			assert.Equal(t, "https://jira.example.com/rest/api/2/issue", req.URL.String())

			// Verify basic auth
			token := req.Header.Get("Authorization")
			assert.Equal(t, "Bearer test-token", token)

			// Verify request body
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			var issueData map[string]interface{}
			err = json.Unmarshal(body, &issueData)
			assert.NoError(t, err)

			fields, ok := issueData["fields"].(map[string]interface{})
			assert.True(t, ok)

			// Check essential fields
			project, ok := fields["project"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "TEST", project["key"])

			issueType, ok := fields["issuetype"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "Task", issueType["name"])
			// Check summary and description are set
			summary, ok := fields["summary"].(string)
			assert.True(t, ok)
			assert.Equal(t, "default/deployment/nginx: Policy Violation: require-requests-and-limits-required", summary)
			_, ok = fields["description"].(string)
			assert.True(t, ok)

			// Check labels
			labels, ok := fields["labels"].([]interface{})
			assert.True(t, ok)
			assert.Contains(t, labels, "policy-reporter")
			assert.Contains(t, labels, "policy-violation")
		}

		client, _ := jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name: "Jira",
			},
			Host:         "https://jira.example.com",
			Username:     "test-user",
			APIToken:     "test-token",
			ProjectKey:   "TEST",
			IssueType:    "Task",
			CustomFields: map[string]string{"customfield_10001": "PolicyReporter"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(fixtures.CompleteTargetSendResult)
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			// Verify HTTP headers and method
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

			// Verify URL
			assert.Equal(t, "https://jira.example.com/rest/api/2/issue", req.URL.String())

			// Verify basic auth
			username, password, ok := req.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test-user", username)
			assert.Equal(t, "test-password", password)

			// Verify request body
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			var issueData map[string]interface{}
			err = json.Unmarshal(body, &issueData)
			assert.NoError(t, err)

			fields, ok := issueData["fields"].(map[string]interface{})
			assert.True(t, ok)

			// Check summary is set
			summary, ok := fields["summary"].(string)
			assert.True(t, ok)
			assert.Equal(t, "test: Policy Violation: require-requests-and-limits-required", summary)

			// Check essential fields
			project, ok := fields["project"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "TEST", project["key"])

			issueType, ok := fields["issuetype"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "Bug", issueType["name"])

			components, ok := fields["compoenents"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, []string{"policy-reporter"}, components["name"])
		}

		client, err := jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name: "Jira",
			},
			Host:           "https://jira.example.com",
			Username:       "test-user",
			Password:       "test-password",
			ProjectKey:     "TEST",
			IssueType:      "Bug",
			HTTPClient:     testClient{callback, 200},
			CustomFields:   map[string]string{"cluster": "test"},
			Components:     []string{"policy-reporter"},
			SummaryTmplate: "{{ customfield.cluster }}: Policy Violation: {{ result.Policy }}",
		})
		if assert.NoError(t, err) {
			client.Send(fixtures.CompleteTargetSendResult)
		}
	})

	t.Run("Default IssueType", func(t *testing.T) {
		callback := func(req *http.Request) {
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			var issueData map[string]interface{}
			err = json.Unmarshal(body, &issueData)
			assert.NoError(t, err)

			fields, ok := issueData["fields"].(map[string]interface{})
			assert.True(t, ok)

			issueType, ok := fields["issuetype"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "Task", issueType["name"]) // Default should be Task
		}

		client, _ := jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name: "Jira",
			},
			Host:       "https://jira.example.com",
			Username:   "test-user",
			Password:   "test-password",
			ProjectKey: "TEST",
			IssueType:  "", // Empty to test default
			HTTPClient: testClient{callback, 200},
		})
		client.Send(fixtures.CompleteTargetSendResult)
	})

	t.Run("Custom Fields", func(t *testing.T) {
		callback := func(req *http.Request) {
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			var issueData map[string]interface{}
			err = json.Unmarshal(body, &issueData)
			assert.NoError(t, err)

			fields, ok := issueData["fields"].(map[string]interface{})
			assert.True(t, ok)

			// Check custom fields
			assert.Equal(t, "TestCluster", fields["customfield_10001"])
			assert.Equal(t, "Kubernetes", fields["customfield_10002"])
		}

		client, _ := jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name: "Jira",
			},
			Host:       "https://jira.example.com",
			Username:   "test-user",
			Password:   "test-password",
			ProjectKey: "TEST",
			IssueType:  "Task",
			CustomFields: map[string]string{
				"customfield_10001": "TestCluster",
				"customfield_10002": "Kubernetes",
			},
			HTTPClient: testClient{callback, 200},
		})
		client.Send(fixtures.CompleteTargetSendResult)
	})

	t.Run("Name", func(t *testing.T) {
		client, _ := jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name: "JiraTarget",
			},
			Host:       "https://jira.example.com",
			Username:   "test-user",
			Password:   "test-password",
			ProjectKey: "TEST",
			IssueType:  "Task",
			HTTPClient: testClient{func(req *http.Request) {}, 200},
		})

		assert.Equal(t, "JiraTarget", client.Name())
	})

	t.Run("Type", func(t *testing.T) {
		client, _ := jira.NewClient(jira.Options{
			ClientOptions: target.ClientOptions{
				Name: "JiraTarget",
			},
			Host:       "https://jira.example.com",
			ProjectKey: "TEST",
			HTTPClient: testClient{func(req *http.Request) {}, 200},
		})

		assert.Equal(t, target.SingleSend, client.Type())
	})
}
