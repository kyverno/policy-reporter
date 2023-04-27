package v1

import (
	"context"
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

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
	ReportLabel map[string]string
	Search      string
}

type Pagination struct {
	Page      int
	Offset    int
	SortBy    []string
	Direction string
}

type ResultFilterValues struct {
	ReportID  string
	Namespace string
	Source    string
	Kind      string
	Category  string
	Policy    string
	Severity  string
	Result    string
	Count     int
}

func (r ResultFilterValues) Hash() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.ReportID)
	h1 = fnv1a.AddString64(h1, r.Namespace)
	h1 = fnv1a.AddString64(h1, r.Source)
	h1 = fnv1a.AddString64(h1, r.Kind)
	h1 = fnv1a.AddString64(h1, r.Category)
	h1 = fnv1a.AddString64(h1, r.Policy)
	h1 = fnv1a.AddString64(h1, r.Severity)
	h1 = fnv1a.AddString64(h1, r.Result)

	return strconv.FormatUint(h1, 10)
}

func ExtractFilterValues(polr v1alpha2.ReportInterface) []*ResultFilterValues {
	mapping := make(map[string]*ResultFilterValues)
	for _, res := range polr.GetResults() {
		kind := res.GetKind()
		if kind == "" && polr.GetScope() != nil {
			kind = polr.GetScope().Namespace
		}

		value := &ResultFilterValues{
			ReportID:  polr.GetID(),
			Namespace: polr.GetNamespace(),
			Source:    res.Source,
			Kind:      kind,
			Category:  res.Category,
			Policy:    res.Policy,
			Severity:  string(res.Severity),
			Result:    string(res.Result),
			Count:     1,
		}

		if item, ok := mapping[value.Hash()]; ok {
			item.Count = item.Count + 1
		} else {
			mapping[value.Hash()] = value
		}
	}
	list := make([]*ResultFilterValues, 0, len(mapping))
	for _, v := range mapping {
		list = append(list, v)
	}

	return list
}

type PolicyReportFinder interface {
	// FetchClusterPolicyReports by filter and pagination
	FetchClusterPolicyReports(context.Context, Filter, Pagination) ([]*PolicyReport, error)
	// FetchPolicyReports by filter and pagination
	FetchPolicyReports(context.Context, Filter, Pagination) ([]*PolicyReport, error)
	// CountClusterPolicyReports by filter
	CountClusterPolicyReports(context.Context, Filter) (int, error)
	// CountPolicyReports by filter
	CountPolicyReports(context.Context, Filter) (int, error)
	// FetchClusterPolicies from current PolicyReportResults
	FetchClusterPolicies(context.Context, Filter) ([]string, error)
	// FetchClusterRules from current PolicyReportResults
	FetchClusterRules(context.Context, Filter) ([]string, error)
	// FetchNamespacedPolicies from current PolicyReportResults with a Namespace
	FetchNamespacedPolicies(context.Context, Filter) ([]string, error)
	// FetchNamespacedRules from current PolicyReportResults with a Namespace
	FetchNamespacedRules(context.Context, Filter) ([]string, error)
	// FetchClusterCategories from current PolicyReportResults
	FetchClusterCategories(context.Context, Filter) ([]string, error)
	// FetchNamespacedCategories from current PolicyReportResults
	FetchNamespacedCategories(context.Context, Filter) ([]string, error)
	// FetchClusterSources from current PolicyReportResults
	FetchClusterSources(context.Context) ([]string, error)
	// FetchNamespacedSources from current PolicyReportResults with a Namespace
	FetchNamespacedSources(context.Context) ([]string, error)
	// FetchNamespacedKinds from current PolicyReportResults with a Namespace
	FetchNamespacedKinds(context.Context, Filter) ([]string, error)
	// FetchNamespacedResources from current PolicyReportResults with a Namespace
	FetchNamespacedResources(context.Context, Filter) ([]*Resource, error)
	// FetchClusterResources from current PolicyReportResults
	FetchClusterResources(context.Context, Filter) ([]*Resource, error)
	// FetchClusterKinds from current PolicyReportResults
	FetchClusterKinds(context.Context, Filter) ([]string, error)
	// FetchNamespaces from current PolicyReports
	FetchNamespaces(context.Context, Filter) ([]string, error)
	// FetchNamespacedStatusCounts from current PolicyReportResults with a Namespace
	FetchNamespacedStatusCounts(context.Context, Filter) ([]NamespacedStatusCount, error)
	// FetchClusterStatusCounts from current PolicyReportResults
	FetchClusterStatusCounts(context.Context, Filter) ([]StatusCount, error)
	// FetchNamespacedResults from current PolicyReportResults with a Namespace
	FetchNamespacedResults(context.Context, Filter, Pagination) ([]*ListResult, error)
	// FetchClusterResults from current PolicyReportResults
	FetchClusterResults(context.Context, Filter, Pagination) ([]*ListResult, error)
	// CountNamespacedResults from current PolicyReportResults with a Namespace
	CountNamespacedResults(context.Context, Filter) (int, error)
	// CountClusterResults from current PolicyReportResults
	CountClusterResults(context.Context, Filter) (int, error)
	// FetchRuleStatusCounts from current PolicyReportResults
	FetchRuleStatusCounts(context.Context, string, string) ([]StatusCount, error)
	// FetchClusterReportLabels from ClusterPolicyReports
	FetchClusterReportLabels(context.Context, Filter) (map[string][]string, error)
	// FetchNamespacedReportLabels from PolicyReports
	FetchNamespacedReportLabels(context.Context, Filter) (map[string][]string, error)
}
