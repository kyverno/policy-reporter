package v1

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

type PolicyReportFinder interface {
	// FetchClusterPolicyReports by filter and pagination
	FetchClusterPolicyReports(Filter, Pagination) ([]*PolicyReport, error)
	// FetchPolicyReports by filter and pagination
	FetchPolicyReports(Filter, Pagination) ([]*PolicyReport, error)
	// CountClusterPolicyReports by filter
	CountClusterPolicyReports(Filter) (int, error)
	// CountPolicyReports by filter
	CountPolicyReports(Filter) (int, error)
	// FetchClusterPolicies from current PolicyReportResults
	FetchClusterPolicies(Filter) ([]string, error)
	// FetchClusterRules from current PolicyReportResults
	FetchClusterRules(Filter) ([]string, error)
	// FetchNamespacedPolicies from current PolicyReportResults with a Namespace
	FetchNamespacedPolicies(Filter) ([]string, error)
	// FetchNamespacedRules from current PolicyReportResults with a Namespace
	FetchNamespacedRules(Filter) ([]string, error)
	// FetchCategories from current PolicyReportResults
	FetchCategories(Filter) ([]string, error)
	// FetchClusterSources from current PolicyReportResults
	FetchClusterSources() ([]string, error)
	// FetchNamespacedSources from current PolicyReportResults with a Namespace
	FetchNamespacedSources() ([]string, error)
	// FetchNamespacedKinds from current PolicyReportResults with a Namespace
	FetchNamespacedKinds(Filter) ([]string, error)
	// FetchNamespacedResources from current PolicyReportResults with a Namespace
	FetchNamespacedResources(Filter) ([]*Resource, error)
	// FetchClusterResources from current PolicyReportResults
	FetchClusterResources(Filter) ([]*Resource, error)
	// FetchClusterKinds from current PolicyReportResults
	FetchClusterKinds(Filter) ([]string, error)
	// FetchNamespaces from current PolicyReports
	FetchNamespaces(Filter) ([]string, error)
	// FetchNamespacedStatusCounts from current PolicyReportResults with a Namespace
	FetchNamespacedStatusCounts(Filter) ([]NamespacedStatusCount, error)
	// FetchStatusCounts from current PolicyReportResults
	FetchStatusCounts(Filter) ([]StatusCount, error)
	// FetchNamespacedResults from current PolicyReportResults with a Namespace
	FetchNamespacedResults(Filter, Pagination) ([]*ListResult, error)
	// FetchClusterResults from current PolicyReportResults
	FetchClusterResults(Filter, Pagination) ([]*ListResult, error)
	// CountNamespacedResults from current PolicyReportResults with a Namespace
	CountNamespacedResults(Filter) (int, error)
	// CountClusterResults from current PolicyReportResults
	CountClusterResults(Filter) (int, error)
	// FetchRuleStatusCounts from current PolicyReportResults
	FetchRuleStatusCounts(policy, rule string) ([]StatusCount, error)
	// FetchClusterReportLabels from ClusterPolicyReports
	FetchClusterReportLabels(Filter) (map[string][]string, error)
	// FetchNamespacedReportLabels from PolicyReports
	FetchNamespacedReportLabels(Filter) (map[string][]string, error)
}
