package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type QueryBuilder struct {
	query *bun.SelectQuery
}

func (q *QueryBuilder) Filter(column string, values []string) *QueryBuilder {
	if len(values) > 1 {
		q.query.Where(column+" IN (?)", bun.In(values))
	} else if len(values) == 1 {
		q.query.Where(column+" = ?", values[0])
	}

	return q
}

func (q *QueryBuilder) Scoped(scoped bool) *QueryBuilder {
	if scoped {
		return q.NamespaceScope()
	}

	return q.ClusterScope()
}

func (q *QueryBuilder) FilterValue(column string, value string) *QueryBuilder {
	if value != "" {
		q.query.Where(column+" = ?", value)
	}

	return q
}

func (q *QueryBuilder) WithEmpty(column string) *QueryBuilder {
	q.query.Where(column + " = ''")

	return q
}

func (q *QueryBuilder) WithNotEmpty(column string) *QueryBuilder {
	q.query.Where(column + " != ''")

	return q
}

func (q *QueryBuilder) Exclude(filter Filter, prefix string) *QueryBuilder {
	if filter.ResourceID == "" && len(filter.Kinds) == 0 && len(filter.Exclude) > 0 {
		for source, kind := range filter.Exclude {
			q.query.Where(fmt.Sprintf("(%s.source != ? OR (%s.source = ? AND %s.resource_kind NOT IN (?)))", prefix, prefix, prefix), source, source, bun.In(kind))
		}
	}

	return q
}

func (q *QueryBuilder) ResourceSearch(value string) *QueryBuilder {
	if value != "" {
		q.query.Where(`(resource_name LIKE ?0 OR LOWER(resource_kind) = LOWER(?1))`, "%"+value+"%", value)
	}

	return q
}

func (q *QueryBuilder) PolicySearch(value string) *QueryBuilder {
	if value != "" {
		q.query.Where(`(f.policy LIKE ?0 OR f.severity LIKE ?0 OR LOWER(f.resource_kind) = LOWER(?1))`, "%"+value+"%", value)
	}

	return q
}

func (q *QueryBuilder) ResultSearch(value string) *QueryBuilder {
	if value != "" {
		q.query.Where(`(resource_namespace LIKE ?0 OR resource_name LIKE ?0 OR policy LIKE ?0 OR rule LIKE ?0 OR severity = ?1 OR result = ?1 OR LOWER(resource_kind) = LOWER(?1))`, "%"+value+"%", value)
	}

	return q
}

func (q *QueryBuilder) FilterMap(columns map[string][]string) *QueryBuilder {
	for column, values := range columns {
		if len(values) > 1 {
			q.query.Where(column+" IN (?)", bun.In(values))
		} else if len(values) == 1 {
			q.query.Where(column+" = ?", values[0])
		}
	}

	return q
}

func (q *QueryBuilder) FilterOptionalNamespaces(values []string) *QueryBuilder {
	if len(values) > 1 {
		q.query.Where("(resource_namespace IN (?) OR resource_namespace = '')", bun.In(values))
	} else if len(values) == 1 {
		q.query.Where("(resource_namespace = ? OR resource_namespace = '')", values[0])
	}

	return q
}

func (q *QueryBuilder) NamespaceScope() *QueryBuilder {
	q.query.Where("resource_namespace != ''")

	return q
}

func (q *QueryBuilder) ClusterScope() *QueryBuilder {
	q.query.Where("resource_namespace = ''")

	return q
}

func (q *QueryBuilder) Scan(ctx context.Context, dest ...any) error {
	return q.query.Scan(ctx, dest...)
}

func (q *QueryBuilder) Columns(columns ...string) *QueryBuilder {
	q.query.Column(columns...)

	return q
}

func (q *QueryBuilder) Group(columns ...string) *QueryBuilder {
	q.query.Group(columns...)

	return q
}

func (q *QueryBuilder) Order(orders ...string) *QueryBuilder {
	q.query.Order(orders...)

	return q
}

func (q *QueryBuilder) SelectStatusSummaries() *QueryBuilder {
	q.query.ColumnExpr("SUM(res.pass) as pass, SUM(res.warn) as warn, SUM(res.fail) as fail, SUM(res.error) as error, SUM(res.skip) as skip")

	return q
}

func (q *QueryBuilder) Pagination(pagination Pagination) *QueryBuilder {
	q.query.OrderExpr(fmt.Sprintf(
		"%s %s",
		strings.Join(pagination.SortBy, ","),
		pagination.Direction,
	))

	if pagination.Page == 0 || pagination.Offset == 0 {
		return q
	}

	q.query.Limit(pagination.Offset)
	q.query.Offset((pagination.Page - 1) * pagination.Offset)

	return q
}

func (q *QueryBuilder) FilterReportLabels(labels map[string]string) *QueryBuilder {
	if len(labels) > 0 {
		q.query.Join("JOIN policy_report AS pr ON pr.id = policy_report_id")

		for key, value := range labels {
			q.query.Where(fmt.Sprintf(q.jsonExtractLayout(), key), value)
		}
	}

	return q
}

func (q *QueryBuilder) FilterLabels(labels map[string]string) *QueryBuilder {
	if len(labels) > 0 {
		q.query.Join("JOIN policy_report AS pr ON pr.id = policy_report_id")

		for key, value := range labels {
			q.query.Where(fmt.Sprintf(q.jsonExtractLayout(), key), value)
		}
	}

	return q
}

func (q *QueryBuilder) jsonExtractLayout() string {
	if q.query.Dialect().Name() == dialect.PG {
		return "(pr.labels->>'%s') = ?"
	}

	return "json_extract(pr.labels, '$.\"%s\"') = ?"
}

func (q *QueryBuilder) GetQuery() *bun.SelectQuery {
	return q.query
}

func FromQuery(query *bun.SelectQuery) *QueryBuilder {
	return &QueryBuilder{query: query}
}

func NewFilterQuery(db *bun.DB, column string) *QueryBuilder {
	return &QueryBuilder{
		query: db.
			NewSelect().
			TableExpr("policy_report_filter as f").
			Column(column).
			Distinct().
			Order(column + " ASC").
			Where(column + " != ''"),
	}
}

func NewResourceQuery(db *bun.DB) *QueryBuilder {
	return &QueryBuilder{
		query: db.NewSelect().TableExpr("policy_report_resource as res").Distinct(),
	}
}
