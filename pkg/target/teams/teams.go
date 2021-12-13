package teams

import (
	"net/http"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type section struct {
	Title    string `json:"activityTitle"`
	SubTitle string `json:"activitySubtitle"`
	Text     string `json:"text"`
	Facts    []fact `json:"facts,omitempty"`
}

type payload struct {
	Type       string    `json:"@type"`
	Context    string    `json:"@context"`
	Summary    string    `json:"summary,omitempty"`
	ThemeColor string    `json:"themeColor,omitempty"`
	Sections   []section `json:"sections"`
}

var colors = map[report.Priority]string{
	report.DebugPriority:    "68c2ff",
	report.InfoPriority:     "36a64f",
	report.WarningPriority:  "f2c744",
	report.CriticalPriority: "b80707",
	report.ErrorPriority:    "e20b0b",
}

func newPayload(result *report.Result) payload {
	facts := make([]fact, 0)

	facts = append(facts, fact{"Policy", result.Policy})

	if result.Rule != "" {
		facts = append(facts, fact{"Rule", result.Rule})
	}

	facts = append(facts, fact{"Priority", result.Priority.String()})

	if result.Category != "" {
		facts = append(facts, fact{"Category", result.Category})
	}
	if result.Severity != "" {
		facts = append(facts, fact{"Severity", result.Severity})
	}
	res := &report.Resource{}
	if result.HasResource() {
		res = result.Resource
	}

	if res.UID != "" {
		facts = append(facts, fact{"Kind", res.Kind})
		facts = append(facts, fact{"Name", res.Name})
		facts = append(facts, fact{"UID", res.UID})
		if res.Namespace != "" {
			facts = append(facts, fact{"Namespace", res.Namespace})
		}
		facts = append(facts, fact{"API Version", res.APIVersion})
	}

	for property, value := range result.Properties {
		facts = append(facts, fact{strings.Title(property), value})
	}

	timestamp := time.Now()
	if !result.Timestamp.IsZero() {
		timestamp = result.Timestamp
	}

	sections := make([]section, 0, 1)
	sections = append(sections, section{
		Title:    "New Policy Report Result",
		SubTitle: timestamp.Format(time.RFC3339),
		Text:     result.Message,
		Facts:    facts,
	})

	return payload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		Summary:    result.Message,
		ThemeColor: colors[result.Priority],
		Sections:   sections,
	}
}

type client struct {
	target.BaseClient
	webhook string
	client  httpClient
}

func (s *client) Send(result *report.Result) {
	req, err := helper.CreateJSONRequest(s.Name(), "POST", s.webhook, newPayload(result))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := s.client.Do(req)
	helper.ProcessHTTPResponse(s.Name(), resp, err)
}

func (s *client) Name() string {
	return "Teams"
}

// NewClient creates a new teams.client to send Results to MS Teams
func NewClient(host, minimumPriority string, sources []string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		target.NewBaseClient(minimumPriority, sources, skipExistingOnStartup),
		host,
		httpClient,
	}
}
