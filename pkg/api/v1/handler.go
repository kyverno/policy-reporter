package v1

import (
	"net/http"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
)

// TargetsHandler for the Targets REST API
func TargetsHandler(targets []target.Client) http.HandlerFunc {
	apiTargets := make([]Target, 0, len(targets))
	for _, t := range targets {
		apiTargets = append(apiTargets, mapTarget(t))
	}

	return func(w http.ResponseWriter, req *http.Request) {
		helper.SendJSONResponse(w, apiTargets, nil)
	}
}

// ClusterResourcesPolicyListHandler REST API
func ClusterResourcesPolicyListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchClusterPolicies(req.URL.Query().Get("source"))
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesPolicyListHandler REST API
func NamespacedResourcesPolicyListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchNamespacedPolicies(req.URL.Query().Get("source"))
		helper.SendJSONResponse(w, list, err)
	}
}

// CategoryListHandler REST API
func CategoryListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchCategories(req.URL.Query().Get("source"))
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesKindListHandler REST API
func ClusterResourcesKindListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchClusterKinds(req.URL.Query().Get("source"))
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesKindListHandler REST API
func NamespacedResourcesKindListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchNamespacedKinds(req.URL.Query().Get("source"))
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesSourceListHandler REST API
func ClusterResourcesSourceListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchClusterSources()
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedSourceListHandler REST API
func NamespacedSourceListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchNamespacedSources()
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesStatusCountHandler REST API
func ClusterResourcesStatusCountHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchStatusCounts(Filter{
			Kinds:      req.URL.Query()["kinds"],
			Sources:    req.URL.Query()["sources"],
			Categories: req.URL.Query()["categories"],
			Severities: req.URL.Query()["severities"],
			Policies:   req.URL.Query()["policies"],
			Status:     req.URL.Query()["status"],
		})
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesStatusCountsHandler REST API
func NamespacedResourcesStatusCountsHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchNamespacedStatusCounts(Filter{
			Namespaces: req.URL.Query()["namespaces"],
			Kinds:      req.URL.Query()["kinds"],
			Sources:    req.URL.Query()["sources"],
			Categories: req.URL.Query()["categories"],
			Severities: req.URL.Query()["severities"],
			Policies:   req.URL.Query()["policies"],
			Status:     req.URL.Query()["status"],
		})
		helper.SendJSONResponse(w, list, err)
	}
}

// RuleStatusCountHandler REST API
func RuleStatusCountHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchRuleStatusCounts(
			req.URL.Query().Get("policy"),
			req.URL.Query().Get("rule"),
		)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesResultHandler REST API
func NamespacedResourcesResultHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchNamespacedResults(Filter{
			Namespaces: req.URL.Query()["namespaces"],
			Kinds:      req.URL.Query()["kinds"],
			Sources:    req.URL.Query()["sources"],
			Categories: req.URL.Query()["categories"],
			Severities: req.URL.Query()["severities"],
			Policies:   req.URL.Query()["policies"],
			Status:     req.URL.Query()["status"],
		})
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesResultHandler REST API
func ClusterResourcesResultHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchClusterResults(Filter{
			Kinds:      req.URL.Query()["kinds"],
			Sources:    req.URL.Query()["sources"],
			Categories: req.URL.Query()["categories"],
			Severities: req.URL.Query()["severities"],
			Policies:   req.URL.Query()["policies"],
			Status:     req.URL.Query()["status"],
		})
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespaceListHandler REST API
func NamespaceListHandler(finder PolicyReportFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := finder.FetchNamespaces(req.URL.Query().Get("source"))
		helper.SendJSONResponse(w, list, err)
	}
}
