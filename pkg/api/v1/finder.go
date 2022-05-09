package v1

type Filter struct {
	Kinds      []string
	Categories []string
	Namespaces []string
	Sources    []string
	Policies   []string
	Rules      []string
	Severities []string
	Status     []string
	Resources  []string
}

type PolicyReportFinder interface {
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
	FetchNamespacedResults(Filter) ([]*ListResult, error)
	// FetchClusterResults from current PolicyReportResults
	FetchClusterResults(Filter) ([]*ListResult, error)
	// FetchRuleStatusCounts from current PolicyReportResults
	FetchRuleStatusCounts(policy, rule string) ([]StatusCount, error)
}
