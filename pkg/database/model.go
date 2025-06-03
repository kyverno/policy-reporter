package database

import (
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	"github.com/uptrace/bun"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

type Config struct {
	bun.BaseModel `bun:"table:policy_report_config,alias:c"`

	ID      int `bun:"id,pk,autoincrement" json:"id"`
	Version string
}

type PolicyReport struct {
	bun.BaseModel `bun:"table:policy_report,alias:pr" json:"-"`

	ID        string            `bun:",pk" json:"id"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Source    string            `json:"source"`
	Labels    map[string]string `bun:",type:json" json:"labels"`
	Skip      int               `json:"skip"`
	Pass      int               `json:"pass"`
	Warn      int               `json:"warn"`
	Fail      int               `json:"fail"`
	Error     int               `json:"error"`
	Created   int64             `json:"created"`
}

type Resource struct {
	APIVersion string `bun:"api_version"`
	Kind       string
	Name       string
	Namespace  string
	UID        string
}

func (r Resource) GetID() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.Namespace)
	h1 = fnv1a.AddString64(h1, r.Name)
	h1 = fnv1a.AddString64(h1, r.Kind)
	h1 = fnv1a.AddString64(h1, r.APIVersion)

	return strconv.FormatUint(h1, 10)
}

type PolicyReportResult struct {
	bun.BaseModel `bun:"table:policy_report_result,alias:r" json:"-"`

	ID             string   `bun:",pk" json:"id"`
	PolicyReportID string   `bun:"policy_report_id" json:"-"`
	ResourceID     string   `bun:"resource_id"`
	Resource       Resource `bun:"embed:resource_"`
	Policy         string
	Rule           string
	Message        string
	Scored         bool
	Result         string
	Severity       string
	Category       string
	Source         string
	Properties     map[string]string `bun:",type:json"`
	Created        int64
}

type ResourceResult struct {
	bun.BaseModel `bun:"table:policy_report_resource,alias:res" json:"-"`

	ID             string   `bun:",pk"`
	PolicyReportID string   `bun:"policy_report_id,pk"`
	Resource       Resource `bun:"embed:resource_"`
	Source         string   `bun:",pk"`
	Category       string   `bun:"category,pk"`
	Pass           int
	Warn           int
	Fail           int
	Error          int
	Skip           int
	Info           int
	Low            int
	Medium         int
	High           int
	Critical       int
	Unknown        int
}

type PolicyReportFilter struct {
	bun.BaseModel `bun:"table:policy_report_filter,alias:f"`

	PolicyReportID string `bun:"policy_report_id"`
	Namespace      string `bun:"resource_namespace"`
	Kind           string `bun:"resource_kind"`
	Policy         string
	Result         string
	Severity       string
	Category       string
	Source         string
	Count          int
}

func (r *PolicyReportFilter) Hash() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.PolicyReportID)
	h1 = fnv1a.AddString64(h1, r.Namespace)
	h1 = fnv1a.AddString64(h1, r.Source)
	h1 = fnv1a.AddString64(h1, r.Kind)
	h1 = fnv1a.AddString64(h1, r.Category)
	h1 = fnv1a.AddString64(h1, r.Policy)
	h1 = fnv1a.AddString64(h1, r.Severity)
	h1 = fnv1a.AddString64(h1, r.Result)

	return strconv.FormatUint(h1, 10)
}

func MapPolicyReport(r v1alpha1.ReportInterface) *PolicyReport {
	return &PolicyReport{
		ID:        r.GetID(),
		Type:      report.GetType(r),
		Name:      r.GetName(),
		Namespace: r.GetNamespace(),
		Source:    r.GetSource(),
		Labels:    r.GetLabels(),
		Skip:      r.GetSummary().Skip,
		Pass:      r.GetSummary().Pass,
		Warn:      r.GetSummary().Warn,
		Fail:      r.GetSummary().Fail,
		Error:     r.GetSummary().Error,
		Created:   r.GetCreationTimestamp().Unix(),
	}
}

func MapPolicyReportResults(polr v1alpha1.ReportInterface) []*PolicyReportResult {
	list := make([]*PolicyReportResult, 0, len(polr.GetResults()))
	for _, r := range polr.GetResults() {
		res := result.Resource(polr, r)

		ns := res.Namespace
		if ns == "" {
			ns = polr.GetNamespace()
		}

		resource := Resource{
			APIVersion: res.APIVersion,
			Kind:       res.Kind,
			Name:       res.Name,
			Namespace:  ns,
			UID:        string(res.UID),
		}

		list = append(list, &PolicyReportResult{
			ID:             r.GetID(),
			PolicyReportID: polr.GetID(),
			ResourceID:     resource.GetID(),
			Resource:       resource,
			Policy:         r.Policy,
			Rule:           r.Rule,
			Source:         r.Source,
			Scored:         r.Scored,
			Message:        r.Description,
			Result:         string(r.Result),
			Severity:       string(r.Severity),
			Category:       r.Category,
			Properties:     r.Properties,
			Created:        r.Timestamp.Seconds,
		})
	}

	return list
}

func MapPolicyReportFilter(polr v1alpha1.ReportInterface) []*PolicyReportFilter {
	mapping := make(map[string]*PolicyReportFilter)
	for _, res := range polr.GetResults() {
		kind := res.GetKind()
		if kind == "" && polr.GetScope() != nil {
			kind = polr.GetScope().Kind
		}

		value := &PolicyReportFilter{
			PolicyReportID: polr.GetID(),
			Namespace:      polr.GetNamespace(),
			Source:         res.Source,
			Kind:           kind,
			Category:       res.Category,
			Policy:         res.Policy,
			Severity:       string(res.Severity),
			Result:         string(res.Result),
			Count:          1,
		}

		if item, ok := mapping[value.Hash()]; ok {
			item.Count = item.Count + 1
		} else {
			mapping[value.Hash()] = value
		}
	}
	list := make([]*PolicyReportFilter, 0, len(mapping))
	for _, v := range mapping {
		list = append(list, v)
	}

	return list
}

func MapPolicyReportResource(polr v1alpha1.ReportInterface) []*ResourceResult {
	mapping := make(map[string]*ResourceResult)
	for _, res := range polr.GetResults() {
		resource := polr.GetScope()
		if res.HasResource() {
			resource = res.GetResource()
		}

		if resource == nil {
			continue
		}

		r := Resource{
			APIVersion: resource.APIVersion,
			Kind:       resource.Kind,
			UID:        string(resource.UID),
			Namespace:  resource.Namespace,
			Name:       resource.Name,
		}

		id := r.GetID() + res.Category + polr.GetID()

		value, ok := mapping[id]
		if !ok {
			value = &ResourceResult{
				ID:             r.GetID(),
				PolicyReportID: polr.GetID(),
				Resource:       r,
				Source:         res.Source,
				Category:       res.Category,
			}

			mapping[id] = value
		}

		switch res.Result {
		case v1alpha1.StatusPass:
			value.Pass = value.Pass + 1
		case v1alpha1.StatusSkip:
			value.Skip = value.Skip + 1
		case v1alpha1.StatusWarn:
			value.Warn = value.Warn + 1
		case v1alpha1.StatusFail:
			value.Fail = value.Fail + 1
		case v1alpha1.StatusError:
			value.Error = value.Error + 1
		}

		switch res.Severity {
		case v1alpha1.SeverityInfo:
			value.Info = value.Info + 1
		case v1alpha1.SeverityLow:
			value.Low = value.Low + 1
		case v1alpha1.SeverityMedium:
			value.Medium = value.Medium + 1
		case v1alpha1.SeverityHigh:
			value.High = value.High + 1
		case v1alpha1.SeverityCritical:
			value.Critical = value.Critical + 1
		default:
			value.Unknown = value.Unknown + 1
		}
	}

	return helper.ToList(mapping)
}

type Filter struct {
	Kinds       []string
	Categories  []string
	Namespaces  []string
	Sources     []string
	Policies    []string
	Rules       []string
	Severities  []string
	Status      []string
	Resources   []string
	ResourceID  string
	ReportLabel map[string]string
	Exclude     map[string][]string
	Namespaced  bool
	Search      string
}

type Pagination struct {
	Page      int
	Offset    int
	SortBy    []string
	Direction string
}
