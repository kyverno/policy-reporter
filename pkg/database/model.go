package database

import (
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	"github.com/uptrace/bun"
	corev1 "k8s.io/api/core/v1"

	api "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
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

func (r Resource) ID() string {
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
	PolicyReportID string   `bund:"policy_report_id" json:"-"`
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
	Pass           int
	Warn           int
	Fail           int
	Error          int
	Skip           int
}

type PolicyReportFilter struct {
	bun.BaseModel `bun:"table:policy_report_filter,alias:f"`

	PolicyReportID string `bun:"policy_report_id"`
	Namespace      string
	Policy         string
	Kind           string
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

func MapPolicyReport(r v1alpha2.ReportInterface) *PolicyReport {
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

func MapPolicyReportResults(polr v1alpha2.ReportInterface) []*PolicyReportResult {
	list := make([]*PolicyReportResult, 0, len(polr.GetResults()))
	for _, result := range polr.GetResults() {
		res := result.GetResource()
		if res == nil && polr.GetScope() != nil {
			res = polr.GetScope()
		} else if res == nil {
			res = &corev1.ObjectReference{}
		}

		ns := res.Namespace
		if ns == "" {
			ns = polr.GetNamespace()
		}

		list = append(list, &PolicyReportResult{
			ID:             result.GetID(),
			PolicyReportID: polr.GetID(),
			Resource: Resource{
				APIVersion: res.APIVersion,
				Kind:       res.Kind,
				Name:       res.Name,
				Namespace:  ns,
				UID:        string(res.UID),
			},
			Policy:     result.Policy,
			Rule:       result.Rule,
			Source:     result.Source,
			Scored:     result.Scored,
			Message:    result.Message,
			Result:     string(result.Result),
			Severity:   string(result.Severity),
			Category:   result.Category,
			Properties: result.Properties,
			Created:    result.Timestamp.Seconds,
		})
	}

	return list
}

func MapPolicyReportFilter(polr v1alpha2.ReportInterface) []*PolicyReportFilter {
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

func MapListResult(results []*PolicyReportResult) []*api.ListResult {
	list := make([]*api.ListResult, 0, len(results))
	for _, res := range results {
		list = append(list, &api.ListResult{
			ID:         res.ID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
			Message:    res.Message,
			Category:   res.Category,
			Policy:     res.Policy,
			Rule:       res.Rule,
			Status:     res.Result,
			Severity:   res.Severity,
			Timestamp:  res.Created,
			Properties: res.Properties,
		})
	}

	return list
}

func MapResourceResult(results []*ResourceResult) []*api.ResourceResult {
	list := make([]*api.ResourceResult, 0, len(results))
	for _, res := range results {
		list = append(list, &api.ResourceResult{
			ID:         res.ID,
			UID:        res.Resource.UID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
			Source:     res.Source,
			Pass:       res.Pass,
			Skip:       res.Skip,
			Warn:       res.Warn,
			Fail:       res.Fail,
			Error:      res.Error,
		})
	}

	return list
}
func MapPolicyReportResource(polr v1alpha2.ReportInterface) []*ResourceResult {
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

		id := r.ID()

		value, ok := mapping[id]
		if !ok {
			value = &ResourceResult{
				ID:             id,
				PolicyReportID: polr.GetID(),
				Resource:       r,
				Source:         res.Source,
			}

			mapping[id] = value
		}

		switch res.Result {
		case v1alpha2.StatusPass:
			value.Pass = value.Pass + 1
		case v1alpha2.StatusSkip:
			value.Skip = value.Skip + 1
		case v1alpha2.StatusWarn:
			value.Warn = value.Warn + 1
		case v1alpha2.StatusFail:
			value.Fail = value.Fail + 1
		case v1alpha2.StatusError:
			value.Error = value.Error + 1
		}
	}

	return helper.ToList(mapping)
}
