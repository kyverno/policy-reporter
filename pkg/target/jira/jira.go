package jira

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	v2 "github.com/ctreminiom/go-atlassian/v2/jira/v2"
	v3 "github.com/ctreminiom/go-atlassian/v2/jira/v3"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/ctreminiom/go-atlassian/v2/service/common"
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
	APIVersion      string
	SummaryTemplate string
	SkipTLS         bool
	Certificate     string
	Labels          []string
	Components      []string
	CustomFields    map[string]string
	HTTPClient      targethttp.Client
}

type client struct {
	target.BaseClient
	projectKey      string
	issueType       string
	summaryTemplate string
	labels          []string
	compoenents     []string
	customFields    map[string]string
	jiraV2          *v2.Client
	jiraV3          *v3.Client
}

func (e *client) Send(result openreports.ResultAdapter) {
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

	labels := []string{"policy-reporter", "policy-violation"}

	// Add labels
	if result.Policy != "" {
		labels = append(labels, fmt.Sprintf("policy-%s", result.Policy))
	}
	if string(result.Severity) != "" {
		labels = append(labels, fmt.Sprintf("severity-%s", result.Severity))
	}
	if result.HasResource() {
		labels = append(labels, fmt.Sprintf("resource-%s", openreports.ToResourceID(result.GetResource())))
	}
	for _, label := range e.labels {
		labels = append(labels, label)
	}

	if e.jiraV2 != nil {
		resp, err := e.sendV2(summary.String(), result, labels)
		targethttp.ProcessHTTPResponse(e.Name(), resp, err)
	}

	if e.jiraV3 != nil {
		resp, err := e.sendV3(summary.String(), result, labels)
		targethttp.ProcessHTTPResponse(e.Name(), resp, err)
	}
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

func (e *client) sendV2(summary string, result openreports.ResultAdapter, labels []string) (*http.Response, error) {
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

	fields := make(map[string]string)
	for k, v := range e.customFields {
		if strings.HasPrefix(k, "customfield_") {
			continue
		}

		fields[k] = v
	}

	if len(fields) > 0 {
		description += "\n*Custom Fields*:\n"

		for k, v := range fields {
			description += fmt.Sprintf("- %s: %s\n", k, v)
		}
	}

	issue := &models.IssueSchemeV2{
		Fields: &models.IssueFieldsSchemeV2{
			Project:     &models.ProjectScheme{Key: e.projectKey},
			IssueType:   &models.IssueTypeScheme{Name: helper.Defaults(e.issueType, "Task")},
			Summary:     summary,
			Labels:      labels,
			Description: description,
			Components: helper.Map(e.compoenents, func(s string) *models.ComponentScheme {
				return &models.ComponentScheme{Name: s}
			}),
		},
	}

	s, resp, err := e.jiraV2.Issue.Create(context.Background(), issue, e.jiraCustomFields())
	if err == nil {
		zap.L().Debug("JIRA issue created", zap.String("key", s.Key), zap.String("id", s.ID))
	}

	return resp.Response, err
}

func (e *client) sendV3(summary string, result openreports.ResultAdapter, labels []string) (*http.Response, error) {
	document := &models.CommentNodeScheme{
		Version: 1,
		Type:    "doc",
		Content: []*models.CommentNodeScheme{},
	}

	AppendProperty(document, "Policy", result.Policy)
	AppendProperty(document, "Status", string(result.Result))
	AppendProperty(document, "Category", result.Category)
	AppendProperty(document, "Source", result.Source)
	AppendProperty(document, "Message", result.Description)

	if result.HasResource() {
		document.AppendNode(&models.CommentNodeScheme{
			Type:    "paragraph",
			Content: []*models.CommentNodeScheme{{Type: "text", Text: "Resource", Marks: []*models.MarkScheme{{Type: "strong"}}}},
		})

		document.AppendNode(&models.CommentNodeScheme{
			Type: "bulletList",
			Content: []*models.CommentNodeScheme{
				{Type: "listItem", Content: []*models.CommentNodeScheme{{Type: "paragraph", Content: []*models.CommentNodeScheme{{Type: "text", Text: fmt.Sprintf("Kind: %s", result.GetResource().Kind)}}}}},
				{Type: "listItem", Content: []*models.CommentNodeScheme{{Type: "paragraph", Content: []*models.CommentNodeScheme{{Type: "text", Text: fmt.Sprintf("Name: %s", result.GetResource().Name)}}}}},
				{Type: "listItem", Content: []*models.CommentNodeScheme{{Type: "paragraph", Content: []*models.CommentNodeScheme{{Type: "text", Text: fmt.Sprintf("Namespace: %s", result.GetResource().Namespace)}}}}},
				{Type: "listItem", Content: []*models.CommentNodeScheme{{Type: "paragraph", Content: []*models.CommentNodeScheme{{Type: "text", Text: fmt.Sprintf("UID: %s", result.GetResource().UID)}}}}},
			},
		})
	}

	if len(result.Properties) > 0 {
		document.AppendNode(&models.CommentNodeScheme{
			Type:    "paragraph",
			Content: []*models.CommentNodeScheme{{Type: "text", Text: "Additional Properties", Marks: []*models.MarkScheme{{Type: "strong"}}}},
		})

		props := make([]*models.CommentNodeScheme, 0, len(result.Properties))
		for k, v := range result.Properties {
			props = append(props, &models.CommentNodeScheme{Type: "listItem", Content: []*models.CommentNodeScheme{{Type: "paragraph", Content: []*models.CommentNodeScheme{{Type: "text", Text: fmt.Sprintf("%s: %s", k, v)}}}}})
		}

		document.AppendNode(&models.CommentNodeScheme{
			Type:    "bulletList",
			Content: props,
		})
	}

	fields := make(map[string]string)
	for k, v := range e.customFields {
		if strings.HasPrefix(k, "customfield_") {
			continue
		}

		fields[k] = v
	}

	if len(fields) > 0 {
		document.AppendNode(&models.CommentNodeScheme{
			Type:    "paragraph",
			Content: []*models.CommentNodeScheme{{Type: "text", Text: "Custom Fields", Marks: []*models.MarkScheme{{Type: "strong"}}}},
		})

		props := make([]*models.CommentNodeScheme, 0, len(fields))
		for k, v := range fields {
			props = append(props, &models.CommentNodeScheme{Type: "listItem", Content: []*models.CommentNodeScheme{{Type: "paragraph", Content: []*models.CommentNodeScheme{{Type: "text", Text: fmt.Sprintf("%s: %s", k, v)}}}}})
		}

		document.AppendNode(&models.CommentNodeScheme{
			Type:    "bulletList",
			Content: props,
		})
	}

	issue := &models.IssueScheme{
		Fields: &models.IssueFieldsScheme{
			Project:     &models.ProjectScheme{Key: e.projectKey},
			IssueType:   &models.IssueTypeScheme{Name: helper.Defaults(e.issueType, "Task")},
			Summary:     summary,
			Labels:      labels,
			Description: document,
			Components: helper.Map(e.compoenents, func(s string) *models.ComponentScheme {
				return &models.ComponentScheme{Name: s}
			}),
		},
	}

	s, resp, err := e.jiraV3.Issue.Create(context.Background(), issue, e.jiraCustomFields())
	if err == nil {
		zap.L().Debug("JIRA issue created", zap.String("key", s.Key), zap.String("id", s.ID))
	}

	return resp.Response, err
}

func (e *client) jiraCustomFields() *models.CustomFields {
	customFields := models.CustomFields{}
	if len(e.customFields) > 0 {
		for property, value := range e.customFields {
			if !strings.HasPrefix(property, "customfield_") {
				continue
			}

			err := customFields.Text(property, value)
			if err != nil {
				zap.L().Error("failed to add jira custom field", zap.String("name", e.Name()), zap.String("field", property), zap.Error(err))
			}
		}
	}

	return &customFields
}

// NewClient creates a new jira.client to send Results to JIRA
func NewClient(options Options) (target.Client, error) {
	var jiraV2 *v2.Client
	var jiraV3 *v3.Client
	var auth common.Authentication
	var err error

	if options.APIVersion == "v2" {
		jiraV2, err = v2.New(options.HTTPClient, options.Host)
		if err != nil {
			return nil, err
		}

		auth = jiraV2.Auth
	} else {
		jiraV3, err = v3.New(options.HTTPClient, options.Host)
		if err != nil {
			return nil, err
		}

		auth = jiraV3.Auth
	}

	if options.APIToken != "" && options.Username == "" {
		auth.SetBearerToken(options.APIToken)
	} else if options.APIToken != "" && options.Username != "" {
		auth.SetBasicAuth(options.Username, options.APIToken)
	} else {
		auth.SetBasicAuth(options.Username, options.Password)
	}

	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.ProjectKey,
		options.IssueType,
		helper.Defaults(options.SummaryTemplate, summaryTmplate),
		options.Labels,
		options.Components,
		options.CustomFields,
		jiraV2,
		jiraV3,
	}, nil
}

func AppendProperty(doc *models.CommentNodeScheme, key, value string) {
	if doc == nil || value == "" {
		return
	}

	doc.AppendNode(&models.CommentNodeScheme{
		Type: "paragraph",
		Content: []*models.CommentNodeScheme{
			{Type: "text", Text: fmt.Sprintf("%s: ", key), Marks: []*models.MarkScheme{{Type: "strong"}}},
			{Type: "text", Text: value},
		},
	})
}
