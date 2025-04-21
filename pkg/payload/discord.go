package payload

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type embed struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Color       string       `json:"color"`
	Fields      []embedField `json:"fields"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

var discordColors = map[v1alpha2.PolicySeverity]string{
	v1alpha2.SeverityInfo:     "12370112",
	v1alpha2.SeverityLow:      "3066993",
	v1alpha2.StatusWarn:       "15105570",
	v1alpha2.SeverityHigh:     "15158332",
	v1alpha2.SeverityCritical: "15158332",
}

type DiscordPayload struct {
	Content string  `json:"content"`
	Embeds  []embed `json:"embeds"`
}

func (s *PolicyReportResultPayload) ToDiscord() DiscordPayload {
	color := discordColors[s.Result.Severity]

	embedFields := make([]embedField, 0)

	embedFields = append(embedFields, embedField{"Policy", s.Result.Policy, true})

	if s.Result.Rule != "" {
		embedFields = append(embedFields, embedField{"Rule", s.Result.Rule, true})
	}

	if s.Result.Category != "" {
		embedFields = append(embedFields, embedField{"Category", s.Result.Category, true})
	}
	if s.Result.Severity != "" {
		embedFields = append(embedFields, embedField{"Severity", string(s.Result.Severity), true})
	}

	if s.Result.HasResource() {
		res := s.Result.GetResource()

		embedFields = append(embedFields, embedField{"Kind", res.Kind, true})
		embedFields = append(embedFields, embedField{"Name", res.Name, true})
		if res.Namespace != "" {
			embedFields = append(embedFields, embedField{"Namespace", res.Namespace, true})
		}
		if res.APIVersion != "" {
			embedFields = append(embedFields, embedField{"API Version", res.APIVersion, true})
		}
	}

	for property, value := range s.Result.Properties {
		embedFields = append(embedFields, embedField{strings.Title(property), value, true})
	}

	embeds := make([]embed, 0, 1)
	embeds = append(embeds, embed{
		Title:       "New Policy Report Result",
		Description: s.Result.Message,
		Color:       color,
		Fields:      embedFields,
	})

	return DiscordPayload{
		Content: "",
		Embeds:  embeds,
	}
}
