package v1

import (
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var defaultOrder = []string{"resource_namespace", "resource_name", "resource_uid", "policy", "rule", "message"}

type Handler struct {
	finder PolicyReportFinder
}

func (h *Handler) logError(err error) {
	if err != nil {
		zap.L().Error("failed to load data", zap.Error(err))
	}
}

// TargetsHandler for the Targets REST API
func (h *Handler) TargetsHandler(targets []target.Client) http.HandlerFunc {
	apiTargets := make([]Target, 0, len(targets))
	for _, t := range targets {
		apiTargets = append(apiTargets, mapTarget(t))
	}

	return func(w http.ResponseWriter, req *http.Request) {
		helper.SendJSONResponse(w, apiTargets, nil)
	}
}

// PolicyReportListHandler REST API
func (h *Handler) PolicyReportListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		filter := buildFilter(req)
		count, _ := h.finder.CountPolicyReports(filter)
		list, err := h.finder.FetchPolicyReports(filter, buildPaginatiomn(req, []string{"namespace", "name"}))
		h.logError(err)
		helper.SendJSONResponse(w, PolicyReportList{Items: list, Count: count}, err)
	}
}

// PolicyReportListHandler REST API
func (h *Handler) ClusterPolicyReportListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		filter := buildFilter(req)
		count, _ := h.finder.CountClusterPolicyReports(filter)
		list, err := h.finder.FetchClusterPolicyReports(filter, buildPaginatiomn(req, []string{"namespace", "name"}))
		h.logError(err)
		helper.SendJSONResponse(w, PolicyReportList{Items: list, Count: count}, err)
	}
}

// ClusterResourcesPolicyListHandler REST API
func (h *Handler) ClusterResourcesPolicyListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchClusterPolicies(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesRuleListHandler REST API
func (h *Handler) ClusterResourcesRuleListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchClusterRules(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesPolicyListHandler REST API
func (h *Handler) NamespacedResourcesPolicyListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedPolicies(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesRuleListHandler REST API
func (h *Handler) NamespacedResourcesRuleListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedRules(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// CategoryListHandler REST API
func (h *Handler) CategoryListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchCategories(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesKindListHandler REST API
func (h *Handler) ClusterResourcesKindListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchClusterKinds(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesKindListHandler REST API
func (h *Handler) NamespacedResourcesKindListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedKinds(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesListHandler REST API
func (h *Handler) ClusterResourcesListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchClusterResources(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesListHandler REST API
func (h *Handler) NamespacedResourcesListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedResources(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesSourceListHandler REST API
func (h *Handler) ClusterResourcesSourceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchClusterSources()
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedSourceListHandler REST API
func (h *Handler) NamespacedSourceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedSources()
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedReportLabelListHandler REST API
func (h *Handler) NamespacedReportLabelListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedReportLabels(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterReportLabelListHandler REST API
func (h *Handler) ClusterReportLabelListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchClusterReportLabels(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// ClusterResourcesStatusCountHandler REST API
func (h *Handler) ClusterResourcesStatusCountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchStatusCounts(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesStatusCountsHandler REST API
func (h *Handler) NamespacedResourcesStatusCountsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespacedStatusCounts(buildFilter(req))
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// RuleStatusCountHandler REST API
func (h *Handler) RuleStatusCountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchRuleStatusCounts(
			req.URL.Query().Get("policy"),
			req.URL.Query().Get("rule"),
		)
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

// NamespacedResourcesResultHandler REST API
func (h *Handler) NamespacedResourcesResultHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		filter := buildFilter(req)
		count, _ := h.finder.CountNamespacedResults(filter)
		list, err := h.finder.FetchNamespacedResults(filter, buildPaginatiomn(req, defaultOrder))
		h.logError(err)
		helper.SendJSONResponse(w, ResultList{Items: list, Count: count}, err)
	}
}

// ClusterResourcesResultHandler REST API
func (h *Handler) ClusterResourcesResultHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		filter := buildFilter(req)
		count, _ := h.finder.CountClusterResults(filter)
		list, err := h.finder.FetchClusterResults(filter, buildPaginatiomn(req, defaultOrder))
		h.logError(err)
		helper.SendJSONResponse(w, ResultList{Items: list, Count: count}, err)
	}
}

// NamespaceListHandler REST API
func (h *Handler) NamespaceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		list, err := h.finder.FetchNamespaces(Filter{
			Sources:    req.URL.Query()["sources"],
			Categories: req.URL.Query()["categories"],
			Policies:   req.URL.Query()["policies"],
			Rules:      req.URL.Query()["rules"],
		})
		h.logError(err)
		helper.SendJSONResponse(w, list, err)
	}
}

func buildPaginatiomn(req *http.Request, defaultOrder []string) Pagination {
	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 0
	}
	offset, err := strconv.Atoi(req.URL.Query().Get("offset"))
	if err != nil || offset < 1 {
		offset = 0
	}
	direction := "ASC"
	if strings.ToLower(req.URL.Query().Get("direction")) == "desc" {
		direction = "DESC"
	}
	sortBy := req.URL.Query()["sortBy"]
	if len(sortBy) == 0 {
		sortBy = defaultOrder
	}

	return Pagination{
		Page:      page,
		Offset:    offset,
		SortBy:    sortBy,
		Direction: direction,
	}
}

func buildFilter(req *http.Request) Filter {
	labels := map[string]string{}

	for _, label := range req.URL.Query()["labels"] {
		parts := strings.Split(label, ":")
		if len(parts) != 2 {
			continue
		}

		labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return Filter{
		Namespaces:  req.URL.Query()["namespaces"],
		Kinds:       req.URL.Query()["kinds"],
		Resources:   req.URL.Query()["resources"],
		Sources:     req.URL.Query()["sources"],
		Categories:  req.URL.Query()["categories"],
		Severities:  req.URL.Query()["severities"],
		Policies:    req.URL.Query()["policies"],
		Rules:       req.URL.Query()["rules"],
		Status:      req.URL.Query()["status"],
		ReportLabel: labels,
		Search:      req.URL.Query().Get("search"),
	}
}

func NewHandler(finder PolicyReportFinder) *Handler {
	return &Handler{
		finder: finder,
	}
}
