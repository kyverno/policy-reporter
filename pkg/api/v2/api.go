package v2

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/config"
	db "github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/namespaces"
)

var defaultOrder = []string{"resource_namespace", "resource_name", "resource_uid", "policy", "rule", "message"}

type APIHandler struct {
	store    *db.Store
	nsClient namespaces.Client
	targets  map[string][]*Target
}

func (h *APIHandler) Register(engine *gin.RouterGroup) error {
	engine.GET("resource/:id/status-counts", h.GetResourceStatusCounts)
	engine.GET("resource/:id/resource-results", h.ListResourceResults)
	engine.GET("resource/:id/results", h.ListResourcePolilcyResults)
	engine.GET("resource/:id", h.GetResource)
	engine.GET("resource/:id/source-categories", h.ListResourceCategories)

	engine.POST("namespaces/resolve-selector", h.ResolveNamespaceSelector)
	engine.GET("namespaces", h.ListNamespaces)
	engine.GET("sources", h.ListSources)
	engine.GET("sources/:source/use-resources", h.UseResources)
	engine.GET("sources/categories", h.ListSourceWithCategories)
	engine.GET("policies", h.ListPolicies)
	engine.GET("findings", h.ListFindings)
	engine.GET("results-without-resources", h.ListResultsWithoutResource)
	engine.GET("targets", h.ListTargets)

	ns := engine.Group("namespace-scoped")
	ns.GET("resource-results", h.ListNamespaceResourceResults)
	ns.GET(":source/status-counts", h.GetNamespaceStatusCounts)
	ns.GET("kinds", h.ListNamespaceKinds)
	ns.GET("results", h.ListPolicyResults(true))

	cluster := engine.Group("cluster-scoped")
	cluster.GET("resource-results", h.ListClusterResourceResults)
	cluster.GET(":source/status-counts", h.GetClusterStatusCounts)
	cluster.GET("kinds", h.ListClusterKinds)
	cluster.GET("results", h.ListPolicyResults(false))

	return nil
}

func (h *APIHandler) ResolveNamespaceSelector(ctx *gin.Context) {
	selector := make(map[string]string)
	if err := ctx.BindJSON(&selector); err != nil {
		zap.L().Error("resolve namespace selector: failed to convert request body", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid selector content"))
	}

	list, err := h.nsClient.List(ctx, selector)

	api.SendResponse(ctx, list, "failed to get namespaces for the provided selector", err)
}

func (h *APIHandler) ListNamespaces(ctx *gin.Context) {
	list, err := h.store.FetchNamespaces(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, list, "failed to load namespaces", err)
}

func (h *APIHandler) ListSources(ctx *gin.Context) {
	sources, err := h.store.FetchSources(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, sources, "failed to load sources", err)
}

func (h *APIHandler) ListPolicies(ctx *gin.Context) {
	policies, err := h.store.FetchPolicies(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, MapPolicies(policies), "failed to load policies", err)
}

func (h *APIHandler) ListSourceWithCategories(ctx *gin.Context) {
	categories, err := h.store.FetchCategories(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, MapToSourceDetails(categories), "failed to load source details", err)
}

func (h *APIHandler) ListResourceCategories(ctx *gin.Context) {
	categories, err := h.store.FetchResourceCategories(ctx, ctx.Param("id"), api.BuildFilter(ctx))

	api.SendResponse(ctx, MapResourceCategoryToSourceDetails(categories), "failed to load source details", err)
}

func (h *APIHandler) GetResource(ctx *gin.Context) {
	resource, err := h.store.FetchResource(ctx, ctx.Param("id"))

	api.SendResponse(ctx, MapResource(resource), "failed to load source details", err)
}

func (h *APIHandler) GetResourceStatusCounts(ctx *gin.Context) {
	counts, err := h.store.FetchResourceStatusCounts(ctx, ctx.Param("id"), api.BuildFilter(ctx))

	api.SendResponse(ctx, MapResourceStatusCounts(counts), "failed to load resource status counts", err)
}

func (h *APIHandler) ListNamespaceResourceResults(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)
	list, err := h.store.FetchNamespaceResourceResults(ctx, filter, api.BuildPagination(ctx, []string{"resource_namespace", "resource_name", "resource_uid"}))
	if err != nil {
		zap.L().Error("failed to load resource results", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	count, err := h.store.CountNamespaceResourceResults(ctx, filter)

	api.SendResponse(ctx, Paginated[ResourceResult]{Count: count, Items: MapResourceResults(list)}, "failed to load resource result list", err)
}

func (h *APIHandler) ListClusterResourceResults(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)
	list, err := h.store.FetchClusterResourceResults(ctx, filter, api.BuildPagination(ctx, []string{"resource_namespace", "resource_name", "resource_uid"}))
	if err != nil {
		zap.L().Error("failed to load resource results", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	count, err := h.store.CountClusterResourceResults(ctx, filter)

	api.SendResponse(ctx, Paginated[ResourceResult]{Count: count, Items: MapResourceResults(list)}, "failed to load resource result list", err)
}

func (h *APIHandler) GetClusterStatusCounts(ctx *gin.Context) {
	results, err := h.store.FetchClusterStatusCounts(ctx, ctx.Param("source"), api.BuildFilter(ctx))

	api.SendResponse(ctx, MapClusterStatusCounts(results), "failed to calculate cluster status counts", err)
}

func (h *APIHandler) GetNamespaceStatusCounts(ctx *gin.Context) {
	results, err := h.store.FetchNamespaceStatusCounts(ctx, ctx.Param("source"), api.BuildFilter(ctx))

	api.SendResponse(ctx, MapNamespaceStatusCounts(results), "failed to calculate namespace status counts", err)
}

func (h *APIHandler) ListClusterKinds(ctx *gin.Context) {
	kinds, err := h.store.FetchClusterKinds(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, kinds, "failed to load cluster kinds", err)
}

func (h *APIHandler) ListNamespaceKinds(ctx *gin.Context) {
	kinds, err := h.store.FetchNamespaceKinds(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, kinds, "failed to load namespaced kinds", err)
}

func (h *APIHandler) ListResourceResults(ctx *gin.Context) {
	list, err := h.store.FetchResourceResults(ctx, ctx.Param("id"), api.BuildFilter(ctx))

	api.SendResponse(ctx, MapResourceResults(list), "failed to load resource result list", err)
}

func (h *APIHandler) ListResourcePolilcyResults(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)
	list, err := h.store.FetchResourcePolicyResults(ctx, ctx.Param("id"), filter, api.BuildPagination(ctx, defaultOrder))
	if err != nil {
		zap.L().Error("failed to load resource results", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	count, err := h.store.CountResourcePolicyResults(ctx, ctx.Param("id"), filter)

	api.SendResponse(ctx, Paginated[PolicyResult]{Count: count, Items: MapPolicyResults(list)}, "failed to load resource result list", err)
}

func (h *APIHandler) ListPolicyResults(namespaced bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filter := api.BuildFilter(ctx)

		list, err := h.store.FetchResults(ctx, namespaced, filter, api.BuildPagination(ctx, defaultOrder))
		if err != nil {
			zap.L().Error("failed to load results", zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		count, err := h.store.CountResults(ctx, namespaced, filter)

		api.SendResponse(ctx, Paginated[PolicyResult]{Count: count, Items: MapPolicyResults(list)}, "failed to load resource result list", err)
	}
}

func (h *APIHandler) ListResultsWithoutResource(ctx *gin.Context) {
	filter := api.BuildFilter(ctx)

	list, err := h.store.FetchResultsWithoutResource(ctx, filter, api.BuildPagination(ctx, defaultOrder))
	if err != nil {
		zap.L().Error("failed to load results without resources", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	count, err := h.store.CountResultsWithoutResource(ctx, filter)

	api.SendResponse(ctx, Paginated[PolicyResult]{Count: count, Items: MapPolicyResults(list)}, "failed to load result list without resources", err)
}

func (h *APIHandler) UseResources(ctx *gin.Context) {
	resources, err := h.store.UseResources(ctx, ctx.Param("source"), api.BuildFilter(ctx))

	api.SendResponse(ctx, gin.H{"resources": resources}, "failed to check if resources are used", err)
}

func (h *APIHandler) ListFindings(ctx *gin.Context) {
	results, err := h.store.FetchFindingCounts(ctx, api.BuildFilter(ctx))

	api.SendResponse(ctx, MapFindings(results), "failed to load findings", err)
}

func (h *APIHandler) ListTargets(ctx *gin.Context) {
	api.SendResponse(ctx, h.targets, "failed to load findings", nil)
}

func NewAPIHandler(store *db.Store, client namespaces.Client, targets map[string][]*Target) *APIHandler {
	return &APIHandler{
		store:    store,
		nsClient: client,
		targets:  targets,
	}
}

func WithAPI(store *db.Store, client namespaces.Client, targets config.Targets) api.ServerOption {
	return func(s *api.Server) error {
		return s.Register("v2", NewAPIHandler(store, client, MapConfigTagrgets(targets)))
	}
}
