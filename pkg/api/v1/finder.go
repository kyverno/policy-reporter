package v1

type Filter struct {
	Kinds      []string
	Categories []string
	Namespaces []string
	Sources    []string
	Policies   []string
	Severities []string
	Status     []string
}

type PolicyReportFinder interface {
	// FetchClusterPolicies from current PolicyReportResults
	FetchClusterPolicies(source string) ([]string, error)
	// FetchNamespacedPolicies from current PolicyReportResults with a Namespace
	FetchNamespacedPolicies(source string) ([]string, error)
	// FetchCategories from current PolicyReportResults
	FetchCategories(source string) ([]string, error)
	// FetchClusterSources from current PolicyReportResults
	FetchClusterSources() ([]string, error)
	// FetchNamespacedSources from current PolicyReportResults with a Namespace
	FetchNamespacedSources() ([]string, error)
	// FetchNamespacedKinds from current PolicyReportResults with a Namespace
	FetchNamespacedKinds(source string) ([]string, error)
	// FetchClusterKinds from current PolicyReportResults
	FetchClusterKinds(source string) ([]string, error)
	// FetchNamespaces from current PolicyReports
	FetchNamespaces(source string) ([]string, error)
	// FetchNamespacedStatusCounts from current PolicyReportResults with a Namespace
	FetchNamespacedStatusCounts(Filter) ([]NamespacedStatusCount, error)
	// FetchStatusCounts from current PolicyReportResults
	FetchStatusCounts(Filter) ([]StatusCount, error)
	// FetchNamespacedResults from current PolicyReportResults with a Namespace
	FetchNamespacedResults(filter Filter) ([]*ListResult, error)
	// FetchClusterResults from current PolicyReportResults
	FetchClusterResults(filter Filter) ([]*ListResult, error)
	// FetchRuleStatusCounts from current PolicyReportResults
	FetchRuleStatusCounts(policy, rule string) ([]StatusCount, error)
}
