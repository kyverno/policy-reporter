package v2

import (
	"fmt"
	"net/url"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	db "github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/filters"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type StatusList struct {
	Pass  int `json:"pass"`
	Skip  int `json:"skip"`
	Warn  int `json:"warn"`
	Error int `json:"error"`
	Fail  int `json:"fail"`
}

type SeverityList struct {
	Unknown  int `json:"unknown"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Medium   int `json:"medium"`
	High     int `json:"high"`
	Critical int `json:"critical"`
}

type Category struct {
	Name       string        `json:"name"`
	Status     *StatusList   `json:"status"`
	Severities *SeverityList `json:"severities"`
}

type SourceDetails struct {
	Name       string      `json:"name"`
	Categories []*Category `json:"categories"`
}

func MapToSourceDetails(categories []db.Category) []*SourceDetails {
	list := make(map[string]*SourceDetails, 0)

	for _, r := range categories {
		if s, ok := list[r.Source]; ok {
			UpdateCategory(r, s)
			continue
		}

		list[r.Source] = &SourceDetails{
			Name: r.Source,
			Categories: []*Category{{
				Name:       helper.Defaults(r.Name, "Other"),
				Status:     &StatusList{},
				Severities: &SeverityList{},
			}},
		}

		UpdateCategory(r, list[r.Source])
	}

	return helper.ToList(list)
}

func UpdateCategory(result db.Category, source *SourceDetails) {
	for _, c := range source.Categories {
		if c.Name == helper.Defaults(result.Name, "Other") {
			MapResultToCategory(result, c)
			MapSeverityToCategory(result, c)
			return
		}
	}

	category := &Category{
		Name:       helper.Defaults(result.Name, "Other"),
		Status:     &StatusList{},
		Severities: &SeverityList{},
	}

	category = MapResultToCategory(result, category)
	category = MapSeverityToCategory(result, category)

	source.Categories = append(source.Categories, category)
}

func MapResultToCategory(result db.Category, category *Category) *Category {
	switch result.Result {
	case v1alpha2.StatusPass:
		category.Status.Pass += result.Count
	case v1alpha2.StatusWarn:
		category.Status.Warn += result.Count
	case v1alpha2.StatusFail:
		category.Status.Fail += result.Count
	case v1alpha2.StatusError:
		category.Status.Error += result.Count
	case v1alpha2.StatusSkip:
		category.Status.Skip += result.Count
	}

	return category
}

func MapSeverityToCategory(result db.Category, category *Category) *Category {
	switch result.Severity {
	case v1alpha2.SeverityLow:
		category.Severities.Low += result.Count
	case v1alpha2.SeverityInfo:
		category.Severities.Info += result.Count
	case v1alpha2.SeverityMedium:
		category.Severities.Medium += result.Count
	case v1alpha2.SeverityHigh:
		category.Severities.High += result.Count
	case v1alpha2.SeverityCritical:
		category.Severities.Critical += result.Count
	default:
		category.Severities.Unknown += result.Count
	}

	return category
}

type Resource struct {
	ID         string `json:"id,omitempty"`
	UID        string `json:"uid,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

func MapResource(result db.ResourceResult) Resource {
	return Resource{
		ID:         result.ID,
		UID:        result.Resource.UID,
		APIVersion: result.Resource.APIVersion,
		Kind:       result.Resource.Kind,
		Name:       result.Resource.Name,
		Namespace:  result.Resource.Namespace,
	}
}

type ResourceStatusCount struct {
	Source string `json:"source,omitempty"`
	Pass   int    `json:"pass"`
	Warn   int    `json:"warn"`
	Fail   int    `json:"fail"`
	Error  int    `json:"error"`
	Skip   int    `json:"skip"`
}

func MapResourceStatusCounts(results []db.ResourceStatusCount) []ResourceStatusCount {
	list := make([]ResourceStatusCount, 0, len(results))
	for _, result := range results {
		list = append(list, ResourceStatusCount{
			Source: result.Source,
			Pass:   result.Pass,
			Fail:   result.Fail,
			Warn:   result.Warn,
			Error:  result.Error,
			Skip:   result.Skip,
		})
	}

	return list
}

type ResourceSeverityCount struct {
	Source   string `json:"source,omitempty"`
	Info     int    `json:"info"`
	Low      int    `json:"low"`
	Medium   int    `json:"medium"`
	High     int    `json:"high"`
	Critical int    `json:"critical"`
	Unknown  int    `json:"unknown"`
}

func MapResourceSeverityCounts(results []db.ResourceSeverityCount) []ResourceSeverityCount {
	list := make([]ResourceSeverityCount, 0, len(results))
	for _, result := range results {
		list = append(list, ResourceSeverityCount{
			Source:   result.Source,
			Info:     result.Info,
			Low:      result.Low,
			Medium:   result.Medium,
			High:     result.High,
			Critical: result.Critical,
			Unknown:  result.Unknown,
		})
	}

	return list
}

type Status struct {
	Pass  int `json:"pass"`
	Skip  int `json:"skip"`
	Warn  int `json:"warn"`
	Fail  int `json:"fail"`
	Error int `json:"error"`
}

type Severities struct {
	Info     int `json:"info"`
	Low      int `json:"low"`
	Medium   int `json:"medium"`
	High     int `json:"high"`
	Critical int `json:"critical"`
	Unknown  int `json:"unknown"`
}

type ResourceResult struct {
	ID         string     `json:"id"`
	UID        string     `json:"uid"`
	Name       string     `json:"name"`
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
	Namespace  string     `json:"namespace,omitempty"`
	Source     string     `json:"source,omitempty"`
	Status     Status     `json:"status"`
	Severities Severities `json:"severities"`
}

func MapResourceResults(results []db.ResourceResult) []ResourceResult {
	return helper.Map(results, func(res db.ResourceResult) ResourceResult {
		return ResourceResult{
			ID:         res.ID,
			UID:        res.Resource.UID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
			Source:     res.Source,
			Status: Status{
				Pass:  res.Pass,
				Skip:  res.Skip,
				Warn:  res.Warn,
				Fail:  res.Fail,
				Error: res.Error,
			},
			Severities: Severities{
				Unknown:  res.Unknown,
				Info:     res.Info,
				Low:      res.Low,
				Medium:   res.Medium,
				High:     res.High,
				Critical: res.Critical,
			},
		}
	})
}

type Paginated[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}

type StatusCount struct {
	Namespace string `json:"namespace,omitempty"`
	Source    string `json:"source,omitempty"`
	Status    string `json:"status"`
	Count     int    `json:"count"`
}

func MapClusterSeverityCounts(results []db.SeverityCount) map[string]int {
	mapping := map[string]int{
		"unknown":                 0,
		v1alpha2.SeverityLow:      0,
		v1alpha2.SeverityInfo:     0,
		v1alpha2.SeverityMedium:   0,
		v1alpha2.SeverityHigh:     0,
		v1alpha2.SeverityCritical: 0,
	}

	for _, result := range results {
		mapping[result.Severity] = result.Count
	}

	return mapping
}

func MapClusterStatusCounts(results []db.StatusCount) map[string]int {
	mapping := map[string]int{
		v1alpha2.StatusPass:  0,
		v1alpha2.StatusFail:  0,
		v1alpha2.StatusWarn:  0,
		v1alpha2.StatusError: 0,
		v1alpha2.StatusSkip:  0,
	}

	for _, result := range results {
		mapping[result.Status] = result.Count
	}

	return mapping
}

func MapNamespaceStatusCounts(results []db.StatusCount) map[string]map[string]int {
	mapping := map[string]map[string]int{}

	for _, result := range results {
		if _, ok := mapping[result.Namespace]; !ok {
			mapping[result.Namespace] = map[string]int{
				v1alpha2.StatusPass:  0,
				v1alpha2.StatusFail:  0,
				v1alpha2.StatusWarn:  0,
				v1alpha2.StatusError: 0,
				v1alpha2.StatusSkip:  0,
			}
		}

		mapping[result.Namespace][result.Status] = result.Count
	}

	return mapping
}

func MapNamespaceSeverityCounts(results []db.SeverityCount) map[string]map[string]int {
	mapping := map[string]map[string]int{}

	for _, result := range results {
		if _, ok := mapping[result.Namespace]; !ok {
			mapping[result.Namespace] = map[string]int{
				"unknown":                 0,
				v1alpha2.SeverityInfo:     0,
				v1alpha2.SeverityLow:      0,
				v1alpha2.SeverityMedium:   0,
				v1alpha2.SeverityHigh:     0,
				v1alpha2.SeverityCritical: 0,
			}
		}

		mapping[result.Namespace][result.Severity] = result.Count
	}

	return mapping
}

type Policy struct {
	Source   string         `json:"source,omitempty"`
	Category string         `json:"category,omitempty"`
	Name     string         `json:"policy"`
	Severity string         `json:"severity,omitempty"`
	Results  map[string]int `json:"results"`
}

func MapPolicies(results []db.PolicyReportFilter) []*Policy {
	list := make(map[string]*Policy)

	for idx := range results {
		category := results[idx].Category
		if category == "" {
			category = "Other"
		}

		if _, ok := list[results[idx].Policy]; ok {
			list[results[idx].Policy].Results[results[idx].Result] = results[idx].Count
			continue
		}

		list[results[idx].Policy] = &Policy{
			Source:   results[idx].Source,
			Category: category,
			Name:     results[idx].Policy,
			Severity: results[idx].Severity,
			Results: map[string]int{
				results[idx].Result: results[idx].Count,
			},
		}
	}

	return helper.ToList(list)
}

type PolicyResult struct {
	ID         string            `json:"id"`
	Namespace  string            `json:"namespace,omitempty"`
	Kind       string            `json:"kind"`
	APIVersion string            `json:"apiVersion"`
	Name       string            `json:"name"`
	ResourceID string            `json:"resourceId"`
	Message    string            `json:"message"`
	Category   string            `json:"category,omitempty"`
	Policy     string            `json:"policy"`
	Rule       string            `json:"rule"`
	Status     string            `json:"status"`
	Source     string            `json:"source"`
	Severity   string            `json:"severity,omitempty"`
	Timestamp  int64             `json:"timestamp,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

func MapPolicyResults(results []db.PolicyReportResult) []PolicyResult {
	return helper.Map(results, func(res db.PolicyReportResult) PolicyResult {
		return PolicyResult{
			ID:         res.ID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
			ResourceID: res.Resource.GetID(),
			Message:    res.Message,
			Category:   res.Category,
			Policy:     res.Policy,
			Rule:       res.Rule,
			Status:     res.Result,
			Source:     res.Source,
			Severity:   res.Severity,
			Timestamp:  res.Created,
			Properties: res.Properties,
		}
	})
}

type FindingCounts struct {
	Total  int            `json:"total"`
	Source string         `json:"source"`
	Counts map[string]int `json:"counts"`
}

type Findings struct {
	Total     int              `json:"total"`
	PerResult map[string]int   `json:"perResult"`
	Counts    []*FindingCounts `json:"counts"`
}

func MapFindings(results []db.StatusCount) Findings {
	findings := make(map[string]*FindingCounts, 0)
	totals := make(map[string]int, 5)
	total := 0

	for _, count := range results {
		if finding, ok := findings[count.Source]; ok {
			finding.Counts[count.Status] = count.Count
			finding.Total += count.Count
		} else {
			findings[count.Source] = &FindingCounts{
				Source: count.Source,
				Total:  count.Count,
				Counts: map[string]int{
					count.Status: count.Count,
				},
			}
		}

		totals[count.Status] += count.Count
		total += count.Count
	}

	return Findings{Counts: helper.ToList(findings), Total: total, PerResult: totals}
}

func MapSeverityFindings(results []db.SeverityCount) Findings {
	findings := make(map[string]*FindingCounts, 0)
	totals := make(map[string]int, 5)
	total := 0

	for _, count := range results {
		if finding, ok := findings[count.Source]; ok {
			finding.Counts[count.Severity] = count.Count
			finding.Total += count.Count
		} else {
			findings[count.Source] = &FindingCounts{
				Source: count.Source,
				Total:  count.Count,
				Counts: map[string]int{
					count.Severity: count.Count,
				},
			}
		}

		totals[count.Severity] += count.Count
		total += count.Count
	}

	return Findings{Counts: helper.ToList(findings), Total: total, PerResult: totals}
}

func MapResourceCategoryToSourceDetails(categories []db.ResourceCategory) []*SourceDetails {
	list := make(map[string]*SourceDetails, 0)

	for _, r := range categories {
		if s, ok := list[r.Source]; ok {
			s.Categories = append(s.Categories, &Category{
				Name: r.Name,
				Status: &StatusList{
					Pass:  r.Pass,
					Fail:  r.Fail,
					Warn:  r.Warn,
					Error: r.Error,
					Skip:  r.Skip,
				},
			})
			continue
		}

		list[r.Source] = &SourceDetails{
			Name: r.Source,
			Categories: []*Category{{
				Name: r.Name,
				Status: &StatusList{
					Pass:  r.Pass,
					Fail:  r.Fail,
					Warn:  r.Warn,
					Error: r.Error,
					Skip:  r.Skip,
				},
			}},
		}
	}

	return helper.ToList(list)
}

type ValueFilter struct {
	Include  []string       `json:"include,omitempty"`
	Exclude  []string       `json:"exclude,omitempty"`
	Selector map[string]any `json:"selector,omitempty"`
}

type TargetFilter struct {
	Namespaces   *ValueFilter `json:"namespaces,omitempty"`
	Severities   *ValueFilter `json:"severities,omitempty"`
	Status       *ValueFilter `json:"status,omitempty"`
	Policies     *ValueFilter `json:"policies,omitempty"`
	ReportLabels *ValueFilter `json:"reportLabels,omitempty"`
	Sources      *ValueFilter `json:"sources,omitempty"`
}

type Target struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	SecretRef       string            `json:"secretRef,omitempty"`
	MountedSecret   string            `json:"mountedSecret,omitempty"`
	MinimumSeverity string            `json:"minimumSeverity"`
	Filter          TargetFilter      `json:"filter"`
	CustomFields    map[string]string `json:"customFields"`
	Properties      map[string]any    `json:"properties"`
	Host            string            `json:"host,omitempty"`
	SkipTLS         bool              `json:"skipTLS,omitempty"`
	UseTLS          bool              `json:"useTLS,omitempty"`
	Auth            bool              `json:"auth"`
}

func MapValueFilter(f filters.ValueFilter) *ValueFilter {
	if len(f.Exclude)+len(f.Include) == 0+len(f.Selector) {
		return nil
	}

	s := map[string]any{}
	for k, v := range f.Selector {
		s[k] = v
	}

	return &ValueFilter{
		Include:  f.Include,
		Exclude:  f.Exclude,
		Selector: s,
	}
}

func MapBaseToTarget[T any](t *v1alpha1.Config[T]) *Target {
	fields := t.CustomFields
	if fields == nil {
		fields = make(map[string]string, 0)
	}

	return &Target{
		Name:            t.Name,
		MinimumSeverity: t.MinimumSeverity,
		SecretRef:       t.SecretRef,
		MountedSecret:   t.MountedSecret,
		CustomFields:    fields,
		Properties:      make(map[string]any),
		Filter: TargetFilter{
			Namespaces:   MapValueFilter(t.Filter.Namespaces),
			Severities:   MapValueFilter(t.Filter.Severities),
			Status:       MapValueFilter(t.Filter.Status),
			Policies:     MapValueFilter(t.Filter.Policies),
			ReportLabels: MapValueFilter(t.Filter.ReportLabels),
			Sources: MapValueFilter(filters.ValueFilter{
				Include: t.Sources,
			}),
		},
	}
}

func MapSlackToTarget(ta *v1alpha1.Config[v1alpha1.SlackOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "Slack"
	t.Properties["channel"] = ta.Config.Channel

	return t
}

func MapLokiToTarget(ta *v1alpha1.Config[v1alpha1.LokiOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "Loki"
	t.Host = ta.Config.Host
	t.SkipTLS = ta.Config.SkipTLS
	t.UseTLS = ta.Config.Certificate != ""
	t.Properties["api"] = ta.Config.Path

	if v, ok := ta.Config.Headers["Authorization"]; ok && v != "" {
		t.Auth = true
	} else if ta.Config.Username != "" && ta.Config.Password != "" {
		t.Auth = true
	}

	return t
}

func MapElasticsearchToTarget(ta *v1alpha1.Config[v1alpha1.ElasticsearchOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "Elasticsearch"
	t.Host = ta.Config.Host
	t.SkipTLS = ta.Config.SkipTLS
	t.UseTLS = ta.Config.Certificate != ""
	t.Auth = (ta.Config.Username != "" && ta.Config.Password != "") || ta.Config.APIKey != ""
	t.Properties["rotation"] = ta.Config.Rotation
	t.Properties["index"] = ta.Config.Index

	if v, ok := ta.Config.Headers["Authorization"]; ok && v != "" {
		t.Auth = true
	}

	return t
}

func MapWebhhokToTarget(typeName string) func(ta *v1alpha1.Config[v1alpha1.WebhookOptions]) *Target {
	return func(ta *v1alpha1.Config[v1alpha1.WebhookOptions]) *Target {
		t := MapBaseToTarget(ta)
		t.Type = typeName
		t.SkipTLS = ta.Config.SkipTLS
		t.UseTLS = ta.Config.Certificate != ""

		if u, err := url.Parse(ta.Config.Webhook); err == nil {
			t.Host = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			t.Auth = u.User != nil
		}

		if v, ok := ta.Config.Headers["Authorization"]; ok && v != "" {
			t.Auth = true
		}

		return t
	}
}

func MapTelegramToTarget(ta *v1alpha1.Config[v1alpha1.TelegramOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "Telegram"
	t.Host = ta.Config.Webhook
	t.SkipTLS = ta.Config.SkipTLS
	t.UseTLS = ta.Config.Certificate != ""
	t.Properties["chatId"] = ta.Config.ChatID

	return t
}

func MapS3ToTarget(ta *v1alpha1.Config[v1alpha1.S3Options]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "S3"
	t.Host = ta.Config.Endpoint
	t.Properties["prefix"] = ta.Config.Prefix
	t.Properties["bucket"] = ta.Config.Bucket
	t.Properties["region"] = ta.Config.Region
	t.Auth = true

	return t
}

func MapKinesisToTarget(ta *v1alpha1.Config[v1alpha1.KinesisOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "Kinesis"
	t.Host = ta.Config.Endpoint
	t.Properties["stream"] = ta.Config.StreamName
	t.Properties["region"] = ta.Config.Region
	t.Auth = true

	return t
}

func MapSecurityHubToTarget(ta *v1alpha1.Config[v1alpha1.SecurityHubOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "SecurityHub"
	t.Host = ta.Config.Endpoint
	t.Properties["region"] = ta.Config.Region
	t.Properties["synchronize"] = ta.Config.Synchronize
	t.Auth = true

	return t
}

func MapSplunkToTarget(ta *v1alpha1.Config[v1alpha1.SplunkOptions]) *Target {
	t := MapBaseToTarget(ta)

	t.Type = "Splunk"
	t.Properties["host"] = ta.Config.Host
	return t
}

func MapGCSToTarget(ta *v1alpha1.Config[v1alpha1.GCSOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "GoogleCloudStore"
	t.Properties["prefix"] = ta.Config.Prefix
	t.Properties["bucket"] = ta.Config.Bucket
	t.Auth = true

	return t
}

func MapAlertManagerToTarget(ta *v1alpha1.Config[v1alpha1.AlertManagerOptions]) *Target {
	t := MapBaseToTarget(ta)
	t.Type = "AlertManager"
	t.Host = ta.Config.Host
	t.SkipTLS = ta.Config.SkipTLS
	t.UseTLS = ta.Config.Certificate != ""

	if v, ok := ta.Config.Headers["Authorization"]; ok && v != "" {
		t.Auth = true
	}

	return t
}

func MapTargets[T any](c *v1alpha1.Config[T], mapper func(*v1alpha1.Config[T]) *Target) []*Target {
	targets := make([]*Target, 0)

	if c == nil {
		return targets
	}

	if c.Valid {
		targets = append(targets, mapper(c))
	}

	for _, channel := range c.Channels {
		if channel.Valid {
			targets = append(targets, mapper(channel))
		}
	}

	return targets
}

func MapConfigTargets(c target.Targets) map[string][]*Target {
	targets := make(map[string][]*Target)

	targets["loki"] = MapTargets(c.Loki, MapLokiToTarget)
	targets["elasticsearch"] = MapTargets(c.Elasticsearch, MapElasticsearchToTarget)
	targets["slack"] = MapTargets(c.Slack, MapSlackToTarget)
	targets["discord"] = MapTargets(c.Discord, MapWebhhokToTarget("Discord"))
	targets["teams"] = MapTargets(c.Teams, MapWebhhokToTarget("MS Teams"))
	targets["googleChat"] = MapTargets(c.GoogleChat, MapWebhhokToTarget("GoogleChat"))
	targets["webhook"] = MapTargets(c.Webhook, MapWebhhokToTarget("Webhook"))
	targets["telegram"] = MapTargets(c.Telegram, MapTelegramToTarget)
	targets["s3"] = MapTargets(c.S3, MapS3ToTarget)
	targets["kinesis"] = MapTargets(c.Kinesis, MapKinesisToTarget)
	targets["securityHub"] = MapTargets(c.SecurityHub, MapSecurityHubToTarget)
	targets["gcs"] = MapTargets(c.GCS, MapGCSToTarget)
	targets["alertManager"] = MapTargets(c.AlertManager, MapAlertManagerToTarget)
	targets["splunk"] = MapTargets(c.Splunk, MapSplunkToTarget)

	for k, v := range targets {
		if len(v) == 0 {
			delete(targets, k)
		}
	}

	return targets
}

type ResultProperty struct {
	Namespace string `json:"namespace"`
	Property  string `json:"property"`
}

func MapResultPropertyList(results []db.ResultProperty) []ResultProperty {
	return helper.Map(results, func(res db.ResultProperty) ResultProperty {
		return ResultProperty{
			Namespace: res.Namespace,
			Property:  res.Property,
		}
	})
}
