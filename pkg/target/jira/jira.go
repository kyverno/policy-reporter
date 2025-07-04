package jira

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	v2 "github.com/ctreminiom/go-atlassian/v2/jira/v2"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	targethttp "github.com/kyverno/policy-reporter/pkg/target/http"
)

const summaryTmplate = "{{ if .result.ResourceString }}{{ .result.ResourceString }}: {{ end }}Policy Violation: {{ .result.Policy }}"

// Options to configure the JIRA target
type Options struct {
	target.ClientOptions
	Host            string
	Username        string
	Password        string
	APIToken        string
	ProjectKey      string
	IssueType       string
	SummaryTemplate string
	SkipTLS         bool
	Certificate     string
	CustomFields    map[string]string
	Components      []string
	HTTPClient      targethttp.Client
}

type client struct {
	target.BaseClient
	projectKey      string
	issueType       string
	summaryTemplate string
	customFields    map[string]string
	compoenents     []string
	jira            *v2.Client
}

func (e *client) Send(result openreports.ResultAdapter) {
	// Create a detailed description
	description := fmt.Sprintf("*Policy*: %s\n*Severity*: %s\n*Status*: %s\n", result.Policy, result.Severity, result.Result)

	if result.Category != "" {
		description += fmt.Sprintf("*Category*: %s\n", result.Category)
	}

	if result.Source != "" {
		description += fmt.Sprintf("*Source*: %s\n", result.Source)
	}

	if result.Description != "" {
		description += fmt.Sprintf("\n*Message*:\n%s\n", result.Description)
	}

	if result.GetResource() != nil {
		resource := result.GetResource()
		description += fmt.Sprintf("\n*Resource*:\n- Kind: %s\n- Name: %s\n",
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
		description += "\n*Additional Properties*:\n"
		for k, v := range result.Properties {
			description += fmt.Sprintf("- %s: %s\n", k, v)
		}
	}

	customFields := models.CustomFields{}

	// Add custom fields if any
	if len(e.customFields) > 0 {
		for property, value := range e.customFields {
			if !strings.HasPrefix(property, "customfield_") {
				continue
			}

			err := customFields.Text(property, value)
			if err != nil {
				zap.L().Error("failed to add jira custom field", zap.String("name", e.Name()), zap.String("field", property), zap.Error(err), zap.Any("result", result))
			}
		}
	}

	var summary bytes.Buffer

	t, err := template.New("summary").Parse(e.summaryTemplate)
	if err != nil {
		zap.L().Error("failed to parse summary template", zap.String("name", e.Name()), zap.Error(err), zap.Any("result", result))
		return
	}

	if err := t.Execute(&summary, map[string]any{"result": &result, "customfield": e.customFields}); err != nil {
		zap.L().Error("failed to execute summary template", zap.String("name", e.Name()), zap.Error(err), zap.Any("result", result))
		return
	}

	issue := &models.IssueSchemeV2{
		Fields: &models.IssueFieldsSchemeV2{
			Project:     &models.ProjectScheme{Key: e.projectKey},
			IssueType:   &models.IssueTypeScheme{Name: helper.Defaults(e.issueType, "Task")},
			Summary:     summary.String(),
			Labels:      []string{"policy-reporter", "policy-violation"},
			Description: description,
			Components: helper.Map(e.compoenents, func(s string) *models.ComponentScheme {
				return &models.ComponentScheme{Name: s}
			}),
		},
	}

	// Add labels
	if result.Policy != "" {
		issue.Fields.Labels = append(issue.Fields.Labels, fmt.Sprintf("policy-%s", result.Policy))
	}
	if string(result.Severity) != "" {
		issue.Fields.Labels = append(issue.Fields.Labels, fmt.Sprintf("severity-%s", result.Severity))
	}
	if result.HasResource() {
		issue.Fields.Labels = append(issue.Fields.Labels, fmt.Sprintf("resource-%s", openreports.ToResourceID(result.GetResource())))
	}

	s, resp, err := e.jira.Issue.Create(context.Background(), issue, &customFields)
	if err == nil {
		zap.L().Debug("JIRA issue created", zap.String("key", s.Key), zap.String("id", s.ID))
	}

	targethttp.ProcessHTTPResponse(e.Name(), resp.Response, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new jira.client to send Results to JIRA
func NewClient(options Options) (target.Client, error) {
	jira, err := v2.New(options.HTTPClient, options.Host)
	if err != nil {
		return nil, err
	}

	if options.APIToken != "" {
		jira.Auth.SetBearerToken(options.APIToken)
	} else {
		jira.Auth.SetBasicAuth(options.Username, options.Password)
	}

	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.ProjectKey,
		options.IssueType,
		helper.Defaults(options.SummaryTemplate, summaryTmplate),
		options.CustomFields,
		options.Components,
		jira,
	}, nil
}
