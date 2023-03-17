package teams

import (
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Slack target
type Options struct {
	target.ClientOptions
	Webhook      string
	CustomFields map[string]string
	HTTPClient   http.Client
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

var colors = map[v1alpha2.Priority]string{
	v1alpha2.DebugPriority:    "68c2ff",
	v1alpha2.InfoPriority:     "36a64f",
	v1alpha2.WarningPriority:  "f2c744",
	v1alpha2.CriticalPriority: "b80707",
	v1alpha2.ErrorPriority:    "e20b0b",
}

func newPayload(result v1alpha2.PolicyReportResult, customFields map[string]string) payload {
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
		facts = append(facts, fact{"Severity", string(result.Severity)})
	}

	if result.HasResource() {
		res := result.GetResource()

		facts = append(facts, fact{"Kind", res.Kind})
		facts = append(facts, fact{"Name", res.Name})
		if res.UID != "" {
			facts = append(facts, fact{"UID", string(res.UID)})
		}
		if res.Namespace != "" {
			facts = append(facts, fact{"Namespace", res.Namespace})
		}
		if res.APIVersion != "" {
			facts = append(facts, fact{"API Version", res.APIVersion})
		}
	}

	for property, value := range result.Properties {
		facts = append(facts, fact{strings.Title(property), value})
	}
	for property, value := range customFields {
		facts = append(facts, fact{strings.Title(property), value})
	}

	timestamp := time.Now()
	if result.Timestamp.Seconds == 0 {
		timestamp = time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
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
	webhook      string
	customFields map[string]string
	client       http.Client
}

func (s *client) Send(result v1alpha2.PolicyReportResult) {
	req, err := http.CreateJSONRequest(s.Name(), "POST", s.webhook, newPayload(result, s.customFields))
	if err != nil {
		return
	}

	resp, err := s.client.Do(req)
	http.ProcessHTTPResponse(s.Name(), resp, err)
}

// NewClient creates a new teams.client to send Results to MS Teams
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.CustomFields,
		options.HTTPClient,
	}
}
