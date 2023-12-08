package v1

import (
	"context"
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
	ResourceID  string
	ReportLabel map[string]string
	Search      string
}

type Pagination struct {
	Page      int
	Offset    int
	SortBy    []string
	Direction string
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

	FetchFindingCounts(context.Context, Filter) (*Findings, error)
	FetchSources(context.Context, string) ([]*Source, error)

	FetchNamespacedResourceResults(context.Context, Filter, Pagination) ([]*ResourceResult, error)
	FetchClusterResourceResults(context.Context, Filter, Pagination) ([]*ResourceResult, error)
	CountNamespacedResourceResults(context.Context, Filter) (int, error)
	CountClusterResourceResults(context.Context, Filter) (int, error)
	FetchResourceResults(context.Context, string, Filter) ([]*ResourceResult, error)
	FetchResourceStatusCounts(context.Context, string, Filter) ([]ResourceStatusCount, error)
	FetchResource(ctx context.Context, id string) (*Resource, error)

	FetchResults(context.Context, string, Filter, Pagination) ([]*ListResult, error)
	CountResults(context.Context, string, Filter) (int, error)
}
