package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	targethttp "github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the JIRA target
type Options struct {
	target.ClientOptions
	Host         string
	Username     string
	Password     string
	APIToken     string
	ProjectKey   string
	IssueType    string
	SkipTLS      bool
	Certificate  string
	CustomFields map[string]string
	HTTPClient   targethttp.Client
}

type client struct {
	target.BaseClient
	host         string
	username     string
	password     string
	apiToken     string
	projectKey   string
	issueType    string
	skipTLS      bool
	certificate  string
	customFields map[string]string
	client       targethttp.Client
}

// Issue represents a JIRA issue to be created
type Issue struct {
	Fields struct {
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Labels       []string               `json:"labels,omitempty"`
		CustomFields map[string]interface{} `json:"-"`
	} `json:"fields"`
}

func (e *client) Send(result *openreports.ORResultAdapter) {
	issue := Issue{}
	issue.Fields.Project.Key = e.projectKey
	issue.Fields.IssueType.Name = e.issueType
	if e.issueType == "" {
		issue.Fields.IssueType.Name = "Task"
	}

	// Set summary and description based on policy result
	issue.Fields.Summary = fmt.Sprintf("Policy Violation: %s", result.Policy)

	// Create a detailed description
	description := fmt.Sprintf("**Policy**: %s\n**Severity**: %s\n**Status**: %s\n",
		result.Policy, result.Severity, result.Result)

	if result.Category != "" {
		description += fmt.Sprintf("**Category**: %s\n", result.Category)
	}

	if result.Source != "" {
		description += fmt.Sprintf("**Source**: %s\n", result.Source)
	}

	if result.Description != "" {
		description += fmt.Sprintf("\n**Message**:\n%s\n", result.Description)
	}

	if result.GetResource() != nil {
		resource := result.GetResource()
		description += fmt.Sprintf("\n**Resource**:\n- Kind: %s\n- Name: %s\n",
			resource.Kind, resource.Name)

		if resource.Namespace != "" {
			description += fmt.Sprintf("- Namespace: %s\n", resource.Namespace)
		}

		if resource.UID != "" {
			description += fmt.Sprintf("- UID: %s\n", resource.UID)
		}
	}

	// Add properties as additional information
	if len(result.Properties) > 0 {
		description += "\n**Additional Properties**:\n"
		for k, v := range result.Properties {
			description += fmt.Sprintf("- %s: %s\n", k, v)
		}
	}

	issue.Fields.Description = description

	// Add custom fields if any
	if len(e.customFields) > 0 {
		issue.Fields.CustomFields = make(map[string]interface{})
		for property, value := range e.customFields {
			issue.Fields.CustomFields[property] = value
		}
	}

	// Add labels
	issue.Fields.Labels = []string{"policy-reporter", "policy-violation"}
	if result.Policy != "" {
		issue.Fields.Labels = append(issue.Fields.Labels, fmt.Sprintf("policy-%s", result.Policy))
	}
	if string(result.Severity) != "" {
		issue.Fields.Labels = append(issue.Fields.Labels, fmt.Sprintf("severity-%s", result.Severity))
	}

	// Custom fields need to be mapped to the top level of fields for the JIRA API
	issueData := make(map[string]interface{})
	fieldsData := make(map[string]interface{})

	rawIssue, _ := json.Marshal(issue)
	json.Unmarshal(rawIssue, &issueData)

	// Get the fields section
	if fieldsRaw, ok := issueData["fields"].(map[string]interface{}); ok {
		fieldsData = fieldsRaw
	}

	// Add custom fields directly to fields section
	for k, v := range issue.Fields.CustomFields {
		fieldsData[k] = v
	}

	issueData["fields"] = fieldsData

	// Create the JSON request body
	jsonBody, err := json.Marshal(issueData)
	if err != nil {
		return
	}

	// Create HTTP request directly
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/api/2/issue", strings.TrimRight(e.host, "/")), bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}

	// JIRA API requires Content-Type to be exactly "application/json" (without charset)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Policy-Reporter")

	// Set authentication
	if e.apiToken != "" && e.username != "" {
		req.SetBasicAuth(e.username, e.apiToken)
	} else if e.username != "" && e.password != "" {
		req.SetBasicAuth(e.username, e.password)
	}

	// Execute the request
	resp, err := e.client.Do(req)
	targethttp.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new jira.client to send Results to JIRA
func NewClient(options Options) target.Client {
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = targethttp.NewClient(options.Certificate, options.SkipTLS)
	}

	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Username,
		options.Password,
		options.APIToken,
		options.ProjectKey,
		options.IssueType,
		options.SkipTLS,
		options.Certificate,
		options.CustomFields,
		httpClient,
	}
}
