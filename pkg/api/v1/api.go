package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyverno/policy-reporter/pkg/api"
	db "github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var defaultOrder = []string{"resource_namespace", "resource_name", "resource_uid", "policy", "rule", "message"}

type APIHandler struct {
	store   *db.Store
	targets []Target
}

func (h *APIHandler) Register(engine *gin.RouterGroup) error {
	engine.GET("targets", h.ListTargets)
	engine.GET("namespaces", h.ListNamespaces)
	engine.GET("policy-reports", h.ListPolicyReports)
	engine.GET("cluster-policy-reports", h.ListClusterPolicyReports)
	engine.GET("rule-status-count", h.RuleStatusCounts)

	ns := engine.Group("namespaced-resources")
	ns.GET("sources", h.ListNamespacedFilter("source"))
	ns.GET("categories", h.ListNamespacedFilter("category"))
	ns.GET("policies", h.ListNamespacedFilter("policy"))
	ns.GET("kinds", h.ListNamespacedFilter("resource_kind"))
	ns.GET("resources", h.ListNamespacedResources)
	ns.GET("status-counts", h.ListNamespacedStatusCounts)
	ns.GET("results", h.ListNamespacedResults)

	cluster := engine.Group("cluster-resources")
	cluster.GET("sources", h.ListClusterFilter("source"))
	cluster.GET("categories", h.ListClusterFilter("category"))
	cluster.GET("policies", h.ListClusterFilter("policy"))
	cluster.GET("kinds", h.ListClusterFilter("resource_kind"))
	cluster.GET("resources", h.ListClusterResources)
	cluster.GET("status-counts", h.ListClusterStatusCounts)
	cluster.GET("results", h.ListClusterResults)

	return nil
}

func (h *APIHandler) ListTargets(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.targets)
}

func (h *APIHandler) ListPolicyReports(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)

	count, err := h.store.CountPolicyReports(ctx, filter)
	list, err := h.store.FetchPolicyReports(ctx, filter, api.BuildPagination(ctx, []string{"namespace", "name"}))

	api.SendResponse(ctx, api.Paginated[PolicyReport]{Count: count, Items: MapPolicyReports(list)}, "failed to load policy reports", err)
}

func (h *APIHandler) ListClusterPolicyReports(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)

	count, err := h.store.CountClusterPolicyReports(ctx, filter)
	list, err := h.store.FetchClusterPolicyReports(ctx, filter, api.BuildPagination(ctx, []string{"name"}))

	api.SendResponse(ctx, api.Paginated[PolicyReport]{Count: count, Items: MapPolicyReports(list)}, "failed to load policy reports", err)
}

func (h *APIHandler) ListNamespaces(ctx *gin.Context) {
	list, err := h.store.FetchNamespaces(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, list, "failed to load namespaces", err)
}

func (h *APIHandler) RuleStatusCounts(ctx *gin.Context) {
	list, err := h.store.FetchRuleStatusCounts(ctx, ctx.Query("policy"), ctx.Query("rule"))

	api.SendResponse(ctx, MapRuleStatusCounts(list), "failed to load namespaces", err)
}

func (h *APIHandler) ListClusterFilter(filter string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		list, err := h.store.FetchClusterFilter(ctx, filter, api.BuildFilter(ctx))

		api.SendResponse(ctx, list, fmt.Sprintf("failed to load cluster scoped %s list", filter), err)
	}
}

func (h *APIHandler) ListNamespacedFilter(filter string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		list, err := h.store.FetchNamespacedFilter(ctx, filter, api.BuildFilter(ctx))

		api.SendResponse(ctx, list, fmt.Sprintf("failed to load namespace scoped %s list", filter), err)
	}
}

func (h *APIHandler) ListClusterResources(ctx *gin.Context) {
	list, err := h.store.FetchClusterResources(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, MapResource(list), "failed to load cluster scoped resource list", err)
}

func (h *APIHandler) ListNamespacedResources(ctx *gin.Context) {
	list, err := h.store.FetchNamespacedResources(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, MapResource(list), "failed to load namespace scoped resource list", err)
}

func (h *APIHandler) ListClusterStatusCounts(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)
	list, err := h.store.FetchClusterScopedStatusCounts(ctx, filter)

	api.SendResponse(ctx, MapClusterStatusCounts(list, filter.Status), "failed to load cluster scoped status counts", err)
}

func (h *APIHandler) ListNamespacedStatusCounts(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)
	list, err := h.store.FetchNamespaceScopedStatusCounts(ctx, filter)

	api.SendResponse(ctx, MapNamespaceStatusCounts(list, filter.Status), "failed to load namespace scoped status counts", err)
}

func (h *APIHandler) ListClusterResults(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)

	count, err := h.store.CountResults(ctx, false, filter)
	list, err := h.store.FetchResults(ctx, false, filter, api.BuildPagination(ctx, defaultOrder))

	api.SendResponse(ctx, api.Paginated[Result]{Count: count, Items: MapResults(list)}, "failed to load results", err)
}

func (h *APIHandler) ListNamespacedResults(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)

	count, err := h.store.CountResults(ctx, true, filter)
	list, err := h.store.FetchResults(ctx, true, filter, api.BuildPagination(ctx, defaultOrder))

	api.SendResponse(ctx, api.Paginated[Result]{Count: count, Items: MapResults(list)}, "failed to load results", err)
}

func NewAPIHandler(store *db.Store, targets []target.Client) *APIHandler {
	return &APIHandler{store, helper.Map(targets, mapTarget)}
}

func WithAPI(store *db.Store, targets []target.Client) api.ServerOption {
	return func(s *api.Server) error {
		return s.Register("v1", NewAPIHandler(store, targets))
	}
}
