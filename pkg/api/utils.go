package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	db "github.com/kyverno/policy-reporter/pkg/database"
)

type Paginated[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}

func SendResponse(ctx *gin.Context, content any, errMsg string, err error) {
	if err != nil {
		zap.L().Error(errMsg, zap.Error(err))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, content)
}

func BuildFilter(ctx *gin.Context) db.Filter {
	labels := map[string]string{}

	for _, label := range ctx.QueryArray("labels") {
		parts := strings.Split(label, ":")
		if len(parts) != 2 {
			continue
		}

		labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	exclude := map[string][]string{}
	for _, sourceKind := range ctx.QueryArray("exclude") {
		parts := strings.Split(sourceKind, ":")
		length := len(parts)
		if length < 2 {
			continue
		}

		if l, ok := exclude[strings.TrimSpace(parts[0])]; ok {
			exclude[strings.TrimSpace(parts[0])] = append(l, strings.TrimSpace(parts[1]))
		} else {
			exclude[strings.TrimSpace(parts[0])] = []string{strings.TrimSpace(parts[1])}
		}
	}

	id := ctx.Query("resource_id")
	if id == "" {
		id = ctx.Query("id")
	}

	return db.Filter{
		Namespaces:  ctx.QueryArray("namespaces"),
		Kinds:       ctx.QueryArray("kinds"),
		Resources:   ctx.QueryArray("resources"),
		Sources:     ctx.QueryArray("sources"),
		Categories:  ctx.QueryArray("categories"),
		Severities:  ctx.QueryArray("severities"),
		Policies:    ctx.QueryArray("policies"),
		Rules:       ctx.QueryArray("rules"),
		Status:      ctx.QueryArray("status"),
		ReportLabel: labels,
		Search:      ctx.Query("search"),
		ResourceID:  id,
		Exclude:     exclude,
		Namespaced:  ctx.Query("namespaced") == "true",
	}
}

func BuildPagination(ctx *gin.Context, defaultOrder []string) db.Pagination {
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 0
	}
	offset, err := strconv.Atoi(ctx.Query("offset"))
	if err != nil || offset < 1 {
		offset = 0
	}
	direction := "ASC"
	if strings.ToLower(ctx.Query("direction")) == "desc" {
		direction = "DESC"
	}
	sortBy := ctx.QueryArray("sortBy")
	if len(sortBy) == 0 {
		sortBy = defaultOrder
	}

	return db.Pagination{
		Page:      page,
		Offset:    offset,
		SortBy:    sortBy,
		Direction: direction,
	}
}
