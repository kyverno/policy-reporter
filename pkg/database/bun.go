package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

type Type = string

const (
	MySQL      Type = "mysql"
	MariaDB    Type = "mariadb"
	PostgreSQL Type = "postgres"
	SQLite     Type = "sqlite"
)

type Store struct {
	db      *bun.DB
	version string
}

///////////////////////////////
///		V1 API Queries		///
///////////////////////////////

func (s *Store) FetchPolicyReports(ctx context.Context, filter Filter, pagination Pagination) ([]PolicyReport, error) {
	list := make([]PolicyReport, 0)

	err := FromQuery(s.db.NewSelect().Model(&list)).
		FilterMap(map[string][]string{
			"pr.source":    filter.Sources,
			"pr.namespace": filter.Namespaces,
		}).
		FilterLabels(filter.ReportLabel).
		FilterValue("pr.type", report.PolicyReportType).
		Pagination(pagination).
		Scan(ctx)

	return list, err
}

func (s *Store) CountPolicyReports(ctx context.Context, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*PolicyReport)(nil))).
		FilterMap(map[string][]string{
			"pr.source":    filter.Sources,
			"pr.namespace": filter.Namespaces,
		}).
		FilterLabels(filter.ReportLabel).
		FilterValue("pr.type", report.PolicyReportType).
		GetQuery().
		Count(ctx)
}

func (s *Store) FetchClusterPolicyReports(ctx context.Context, filter Filter, pagination Pagination) ([]PolicyReport, error) {
	list := make([]PolicyReport, 0)

	err := FromQuery(s.db.NewSelect().Model(&list)).
		FilterMap(map[string][]string{
			"pr.source": filter.Sources,
		}).
		FilterLabels(filter.ReportLabel).
		FilterValue("pr.type", report.ClusterPolicyReportType).
		Pagination(pagination).
		Scan(ctx)

	return list, err
}

func (s *Store) CountClusterPolicyReports(ctx context.Context, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*PolicyReport)(nil))).
		FilterMap(map[string][]string{
			"pr.source": filter.Sources,
		}).
		FilterLabels(filter.ReportLabel).
		FilterValue("pr.type", report.ClusterPolicyReportType).
		GetQuery().
		Count(ctx)
}

func (s *Store) FetchRuleStatusCounts(ctx context.Context, policy, rule string) ([]StatusCount, error) {
	counts := make([]StatusCount, 0)

	err := s.db.NewSelect().
		Table("policy_report_result").
		ColumnExpr("COUNT(id) as count, result as status").
		Where("rule = ?", rule).
		Where("policy = ?", policy).
		Group("status").
		Scan(ctx, &counts)

	return counts, err
}

func (s *Store) FetchNamespaces(ctx context.Context, filter Filter) ([]string, error) {
	return s.FetchNamespacedFilter(ctx, "resource_namespace", filter)
}

func (s *Store) FetchNamespacedFilter(ctx context.Context, column string, filter Filter) ([]string, error) {
	list := make([]string, 0)

	err := NewFilterQuery(s.db, "f."+column).
		FilterMap(map[string][]string{
			"f.result":             filter.Status,
			"f.source":             filter.Sources,
			"f.category":           filter.Categories,
			"f.policy":             filter.Policies,
			"f.resource_kind":      filter.Kinds,
			"f.resource_namespace": filter.Namespaces,
		}).
		FilterReportLabels(filter.ReportLabel).
		NamespaceScope().
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchClusterFilter(ctx context.Context, column string, filter Filter) ([]string, error) {
	list := make([]string, 0)

	err := NewFilterQuery(s.db, "f."+column).
		FilterMap(map[string][]string{
			"f.source":        filter.Sources,
			"f.category":      filter.Categories,
			"f.policy":        filter.Policies,
			"f.resource_kind": filter.Kinds,
		}).
		FilterReportLabels(filter.ReportLabel).
		ClusterScope().
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchNamespacedResources(ctx context.Context, filter Filter) ([]ResourceResult, error) {
	list := make([]ResourceResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&list).Distinct()).
		Columns("resource_name", "resource_kind", "resource_namespace").
		FilterMap(map[string][]string{
			"res.source":             filter.Sources,
			"res.category":           filter.Categories,
			"res.policy":             filter.Policies,
			"res.resource_kind":      filter.Kinds,
			"res.resource_namespace": filter.Namespaces,
		}).
		FilterReportLabels(filter.ReportLabel).
		NamespaceScope().
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchClusterResources(ctx context.Context, filter Filter) ([]ResourceResult, error) {
	list := make([]ResourceResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&list).Distinct()).
		Columns("resource_name", "resource_kind").
		FilterMap(map[string][]string{
			"f.source":             filter.Sources,
			"f.category":           filter.Categories,
			"f.policy":             filter.Policies,
			"f.resource_kind":      filter.Kinds,
			"f.resource_namespace": filter.Namespaces,
		}).
		FilterReportLabels(filter.ReportLabel).
		ClusterScope().
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchClusterScopedStatusCounts(ctx context.Context, filter Filter) ([]StatusCount, error) {
	list := make([]StatusCount, 0)

	err := FromQuery(s.db.NewSelect().Model(&list).ColumnExpr("SUM(f.count) as count, f.result as status")).
		FilterMap(map[string][]string{
			"f.source":        filter.Sources,
			"f.category":      filter.Categories,
			"f.policy":        filter.Policies,
			"f.resource_kind": filter.Kinds,
			"f.result":        filter.Status,
			"f.severity":      filter.Severities,
		}).
		FilterReportLabels(filter.ReportLabel).
		ClusterScope().
		Group("status").
		Scan(ctx)

	return list, err
}

func (s *Store) FetchNamespaceScopedStatusCounts(ctx context.Context, filter Filter) ([]StatusCount, error) {
	list := make([]StatusCount, 0)

	err := FromQuery(s.db.NewSelect().Model(&list).ColumnExpr("SUM(f.count) as count, f.result as status, f.resource_namespace")).
		FilterMap(map[string][]string{
			"f.source":             filter.Sources,
			"f.category":           filter.Categories,
			"f.policy":             filter.Policies,
			"f.resource_kind":      filter.Kinds,
			"f.resource_namespace": filter.Namespaces,
			"f.result":             filter.Status,
			"f.severity":           filter.Severities,
		}).
		FilterReportLabels(filter.ReportLabel).
		NamespaceScope().
		Group("status", "f.resource_namespace").
		Scan(ctx)

	return list, err
}

///////////////////////////////
///		V2 API Queries		///
///////////////////////////////

func (s *Store) FetchSources(ctx context.Context, filter Filter) ([]string, error) {
	list := make([]string, 0)

	err := NewFilterQuery(s.db, "f.source").
		FilterMap(map[string][]string{
			"f.resource_kind": filter.Kinds,
		}).
		FilterValue("id", filter.ResourceID).
		FilterReportLabels(filter.ReportLabel).
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchCategories(ctx context.Context, filter Filter) ([]Category, error) {
	list := make([]Category, 0)

	err := FromQuery(s.db.NewSelect().Model(&list).ColumnExpr("SUM(f.count) as count")).
		Columns("f.source", "f.category", "f.result", "f.severity").
		FilterMap(map[string][]string{
			"f.source":             filter.Sources,
			"f.category":           filter.Categories,
			"f.resource_kind":      filter.Kinds,
			"f.resource_namespace": filter.Namespaces,
		}).
		Exclude(filter, "f").
		FilterReportLabels(filter.ReportLabel).
		Group("f.source", "f.category", "f.result", "f.severity").
		Order("f.source ASC", "f.category ASC").
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchResource(ctx context.Context, id string) (ResourceResult, error) {
	result := ResourceResult{}

	err := FromQuery(s.db.NewSelect().Model(&result)).
		Columns("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name", "res.source", "res.category").
		SelectStatusSummaries().
		FilterValue("res.id", id).
		Group("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name", "res.source", "res.category").
		GetQuery().
		Limit(1).
		Scan(ctx)

	return result, err
}

func (s *Store) FetchResourceCategories(ctx context.Context, resource string, filter Filter) ([]ResourceCategory, error) {
	list := make([]ResourceCategory, 0)

	err := FromQuery(s.db.NewSelect().Model(&list)).
		Columns("res.source", "res.category").
		SelectStatusSummaries().
		FilterMap(map[string][]string{
			"res.source":   filter.Sources,
			"res.category": filter.Categories,
		}).
		Exclude(filter, "res").
		FilterValue("id", resource).
		FilterReportLabels(filter.ReportLabel).
		Group("res.source", "res.category").
		Order("res.source ASC", "res.category ASC").
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchProperty(ctx context.Context, property string, filter Filter) ([]ResultProperty, error) {
	result := make([]ResultProperty, 0)

	err := FromQuery(
		s.db.NewSelect().
			Model(&result).
			Distinct().
			ColumnExpr(fmt.Sprintf("resource_namespace, properties->>'%s' as property", property))).
		FilterMap(map[string][]string{
			"category":           filter.Categories,
			"source":             filter.Sources,
			"policy":             filter.Policies,
			"rules":              filter.Rules,
			"status":             filter.Status,
			"resource_namespace": filter.Namespaces,
		}).
		GetQuery().
		Where(fmt.Sprintf("properties->'%s' IS NOT NULL", property)).
		Scan(ctx)

	return result, err
}

func (s *Store) FetchResourceStatusCounts(ctx context.Context, id string, filter Filter) ([]ResourceStatusCount, error) {
	result := []ResourceStatusCount{}

	err := FromQuery(s.db.NewSelect().Model(&result)).
		Columns("res.source").
		SelectStatusSummaries().
		FilterMap(map[string][]string{
			"res.category": filter.Categories,
			"res.source":   filter.Sources,
			"policy":       filter.Policies,
		}).
		FilterValue("res.id", id).
		FilterReportLabels(filter.ReportLabel).
		Group("res.source").
		Scan(ctx)

	return result, err
}

func (s *Store) FetchResourceSeverityCounts(ctx context.Context, id string, filter Filter) ([]ResourceSeverityCount, error) {
	result := []ResourceSeverityCount{}

	err := FromQuery(s.db.NewSelect().Model(&result)).
		Columns("res.source").
		SelectSeveritySummaries().
		FilterMap(map[string][]string{
			"res.category": filter.Categories,
			"res.source":   filter.Sources,
			"policy":       filter.Policies,
		}).
		FilterValue("res.id", id).
		FilterReportLabels(filter.ReportLabel).
		Group("res.source").
		Scan(ctx)

	return result, err
}

func (s *Store) FetchNamespaceResourceResults(ctx context.Context, filter Filter, pagination Pagination) ([]ResourceResult, error) {
	results := make([]ResourceResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&results)).
		Columns("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name").
		SelectStatusSummaries().
		SelectSeveritySummaries().
		Group("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name").
		FilterMap(map[string][]string{
			"source":             filter.Sources,
			"category":           filter.Categories,
			"resource_namespace": filter.Namespaces,
			"resource_kind":      filter.Kinds,
		}).
		FilterValue("res.id", filter.ResourceID).
		ResourceSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "res").
		NamespaceScope().
		Pagination(pagination).
		Scan(ctx)

	return results, err
}

func (s *Store) CountNamespaceResourceResults(ctx context.Context, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*ResourceResult)(nil))).
		Columns("res.id").
		Distinct().
		FilterMap(map[string][]string{
			"source":             filter.Sources,
			"category":           filter.Categories,
			"resource_namespace": filter.Namespaces,
			"resource_kind":      filter.Kinds,
		}).
		FilterValue("res.id", filter.ResourceID).
		ResourceSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "res").
		NamespaceScope().
		GetQuery().
		Count(ctx)
}

func (s *Store) FetchClusterResourceResults(ctx context.Context, filter Filter, pagination Pagination) ([]ResourceResult, error) {
	results := make([]ResourceResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&results)).
		Columns("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name").
		SelectStatusSummaries().
		SelectSeveritySummaries().
		Group("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name").
		FilterMap(map[string][]string{
			"source":        filter.Sources,
			"category":      filter.Categories,
			"resource_kind": filter.Kinds,
		}).
		FilterValue("res.id", filter.ResourceID).
		ResourceSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "res").
		ClusterScope().
		Pagination(pagination).
		Scan(ctx)

	return results, err
}

func (s *Store) CountClusterResourceResults(ctx context.Context, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*ResourceResult)(nil))).
		Columns("res.id").
		Distinct().
		FilterMap(map[string][]string{
			"source":        filter.Sources,
			"category":      filter.Categories,
			"resource_kind": filter.Kinds,
		}).
		FilterValue("res.id", filter.ResourceID).
		ResourceSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "res").
		ClusterScope().
		GetQuery().
		Count(ctx)
}

func (s *Store) FetchResourceResults(ctx context.Context, id string, filter Filter) ([]ResourceResult, error) {
	results := make([]ResourceResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&results)).
		Columns("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name", "res.source").
		SelectStatusSummaries().
		FilterValue(`res.id`, id).
		FilterMap(map[string][]string{
			"source":        filter.Sources,
			"category":      filter.Categories,
			"resource_kind": filter.Kinds,
		}).
		FilterReportLabels(filter.ReportLabel).
		ResourceSearch(filter.Search).
		Order("res.source ASC").
		Group("res.id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name", "res.source").
		Scan(ctx)

	return results, err
}

func (s *Store) FetchResourcePolicyResults(ctx context.Context, id string, filter Filter, pagination Pagination) ([]PolicyReportResult, error) {
	results := make([]PolicyReportResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&results)).
		FilterValue(`r.resource_id`, id).
		FilterMap(map[string][]string{
			"source":   filter.Sources,
			"category": filter.Categories,
		}).
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Pagination(pagination).
		Scan(ctx)

	return results, err
}

func (s *Store) CountResourcePolicyResults(ctx context.Context, id string, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*PolicyReportResult)(nil))).
		FilterValue(`r.resource_id`, id).
		FilterMap(map[string][]string{
			"source":   filter.Sources,
			"category": filter.Categories,
		}).
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		GetQuery().
		Count(ctx)
}

func (s *Store) FetchResults(ctx context.Context, namespaced bool, filter Filter, pagination Pagination) ([]PolicyReportResult, error) {
	results := make([]PolicyReportResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&results)).
		FilterMap(map[string][]string{
			"source":             filter.Sources,
			"category":           filter.Categories,
			"policy":             filter.Policies,
			"rule":               filter.Rules,
			"resource_namespace": filter.Namespaces,
			"resource_kind":      filter.Kinds,
			"resource_name":      filter.Resources,
			"result":             filter.Status,
			"severity":           filter.Severities,
		}).
		Scoped(namespaced).
		FilterValue(`r.resource_id`, filter.ResourceID).
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "r").
		Pagination(pagination).
		Scan(ctx)

	return results, err
}

func (s *Store) CountResults(ctx context.Context, namespaced bool, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*PolicyReportResult)(nil))).
		FilterMap(map[string][]string{
			"source":             filter.Sources,
			"category":           filter.Categories,
			"policy":             filter.Policies,
			"rule":               filter.Rules,
			"resource_namespace": filter.Namespaces,
			"resource_kind":      filter.Kinds,
			"resource_name":      filter.Resources,
			"result":             filter.Status,
			"severity":           filter.Severities,
		}).
		Scoped(namespaced).
		FilterValue(`r.resource_id`, filter.ResourceID).
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "r").
		GetQuery().
		Count(ctx)
}

func (s *Store) FetchResultsWithoutResource(ctx context.Context, filter Filter, pagination Pagination) ([]PolicyReportResult, error) {
	results := make([]PolicyReportResult, 0)

	err := FromQuery(s.db.NewSelect().Model(&results)).
		FilterMap(map[string][]string{
			"source":   filter.Sources,
			"category": filter.Categories,
			"policy":   filter.Policies,
			"rule":     filter.Rules,
			"result":   filter.Status,
			"severity": filter.Severities,
		}).
		WithEmpty("resource_name").
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "r").
		Pagination(pagination).
		Scan(ctx)

	return results, err
}

func (s *Store) CountResultsWithoutResource(ctx context.Context, filter Filter) (int, error) {
	return FromQuery(s.db.NewSelect().Model((*PolicyReportResult)(nil))).
		FilterMap(map[string][]string{
			"source":   filter.Sources,
			"category": filter.Categories,
			"policy":   filter.Policies,
			"rule":     filter.Rules,
			"result":   filter.Status,
			"severity": filter.Severities,
		}).
		WithEmpty("resource_name").
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "r").
		GetQuery().
		Count(ctx)
}

func (s *Store) UseResources(ctx context.Context, source string, filter Filter) (bool, error) {
	return FromQuery(s.db.NewSelect().Model((*PolicyReportResult)(nil))).
		FilterValue("source", source).
		FilterMap(map[string][]string{
			"category": filter.Categories,
			"policy":   filter.Policies,
			"rule":     filter.Rules,
		}).
		WithNotEmpty("resource_name").
		ResultSearch(filter.Search).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "r").
		GetQuery().
		Exists(ctx)
}

func (s *Store) FetchClusterStatusCounts(ctx context.Context, source string, filter Filter) ([]StatusCount, error) {
	results := make([]StatusCount, 0)

	err := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.result as status")).
		FilterMap(map[string][]string{
			"category":      filter.Categories,
			"policy":        filter.Policies,
			"resource_kind": filter.Kinds,
		}).
		FilterValue("f.source", source).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		ClusterScope().
		Group("status").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchClusterSeverityCounts(ctx context.Context, source string, filter Filter) ([]SeverityCount, error) {
	results := make([]SeverityCount, 0)

	err := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.severity")).
		FilterMap(map[string][]string{
			"category":      filter.Categories,
			"policy":        filter.Policies,
			"resource_kind": filter.Kinds,
		}).
		FilterValue("f.source", source).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		ClusterScope().
		Group("f.severity").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchNamespaceStatusCounts(ctx context.Context, source string, filter Filter) ([]StatusCount, error) {
	results := make([]StatusCount, 0)

	err := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("f.resource_namespace, SUM(f.count) as count, f.result as status")).
		FilterMap(map[string][]string{
			"f.category":           filter.Categories,
			"f.resource_kind":      filter.Kinds,
			"f.resource_namespace": filter.Namespaces,
			"f.policy":             filter.Policies,
			"f.result":             filter.Status,
		}).
		FilterValue("f.source", source).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		NamespaceScope().
		Group("f.resource_namespace", "f.result").
		Order("f.resource_namespace ASC", "f.result ASC").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchNamespaceSeverityCounts(ctx context.Context, source string, filter Filter) ([]SeverityCount, error) {
	results := make([]SeverityCount, 0)

	err := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("f.resource_namespace, SUM(f.count) as count, f.severity")).
		FilterMap(map[string][]string{
			"f.category":           filter.Categories,
			"f.resource_kind":      filter.Kinds,
			"f.resource_namespace": filter.Namespaces,
			"f.policy":             filter.Policies,
			"f.severity":           filter.Severities,
		}).
		FilterValue("f.source", source).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		NamespaceScope().
		Group("f.resource_namespace", "f.severity").
		Order("f.resource_namespace ASC", "f.severity ASC").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchTotalStatusCounts(ctx context.Context, source string, filter Filter) ([]StatusCount, error) {
	results := make([]StatusCount, 0)

	err := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.result AS status")).
		FilterMap(map[string][]string{
			"category":      filter.Categories,
			"policy":        filter.Policies,
			"resource_kind": filter.Kinds,
		}).
		FilterValue("f.source", source).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		Group("f.result").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchTotalSeverityCounts(ctx context.Context, source string, filter Filter) ([]SeverityCount, error) {
	results := make([]SeverityCount, 0)

	err := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.severity")).
		FilterMap(map[string][]string{
			"category":      filter.Categories,
			"policy":        filter.Policies,
			"resource_kind": filter.Kinds,
		}).
		FilterValue("f.source", source).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		Group("severity").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchNamespaceKinds(ctx context.Context, filter Filter) ([]string, error) {
	list := make([]string, 0)

	err := NewFilterQuery(s.db, "resource_kind").
		FilterMap(map[string][]string{
			"f.source":             filter.Sources,
			"f.category":           filter.Categories,
			"f.resource_namespace": filter.Namespaces,
		}).
		Exclude(filter, "f").
		FilterReportLabels(filter.ReportLabel).
		NamespaceScope().
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchClusterKinds(ctx context.Context, filter Filter) ([]string, error) {
	list := make([]string, 0)

	err := NewFilterQuery(s.db, "resource_kind").
		FilterMap(map[string][]string{
			"f.source":   filter.Sources,
			"f.category": filter.Categories,
		}).
		Exclude(filter, "f").
		FilterReportLabels(filter.ReportLabel).
		ClusterScope().
		Scan(ctx, &list)

	return list, err
}

func (s *Store) FetchPolicies(ctx context.Context, filter Filter) ([]PolicyReportFilter, error) {
	results := make([]PolicyReportFilter, 0)

	err := FromQuery(s.db.NewSelect().Model(&results).ColumnExpr("f.severity, f.category, f.policy, f.source, f.result, SUM(f.count) as count")).
		FilterMap(map[string][]string{
			"f.source":        filter.Sources,
			"f.category":      filter.Categories,
			"f.resource_kind": filter.Kinds,
		}).
		PolicySearch(filter.Search).
		Exclude(filter, "f").
		FilterReportLabels(filter.ReportLabel).
		Order("f.source ASC", "f.category ASC").
		Group("f.category", "f.policy", "f.source", "f.result", "f.severity").
		Scan(ctx)

	return results, err
}

func (s *Store) FetchFindingCounts(ctx context.Context, filter Filter) ([]StatusCount, error) {
	results := make([]StatusCount, 0)

	query := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.result as status, f.source"))

	if filter.Namespaced {
		query.
			NamespaceScope().
			Filter("resource_namespace", filter.Namespaces)
	} else {
		query.FilterOptionalNamespaces(filter.Namespaces)
	}

	err := query.
		FilterMap(map[string][]string{
			"source":        filter.Sources,
			"category":      filter.Categories,
			"resource_kind": filter.Kinds,
			"policy":        filter.Policies,
			"f.result":      filter.Status,
		}).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		Group("f.source", "f.result").
		Order("f.source").
		Scan(ctx, &results)

	return results, err
}

func (s *Store) FetchSeverityFindingCounts(ctx context.Context, filter Filter) ([]SeverityCount, error) {
	results := make([]SeverityCount, 0)

	query := FromQuery(s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.severity, f.source"))

	if filter.Namespaced {
		query.
			NamespaceScope().
			Filter("resource_namespace", filter.Namespaces)
	} else {
		query.FilterOptionalNamespaces(filter.Namespaces)
	}

	err := query.
		FilterMap(map[string][]string{
			"source":        filter.Sources,
			"category":      filter.Categories,
			"resource_kind": filter.Kinds,
			"policy":        filter.Policies,
			"severity":      filter.Severities,
		}).
		FilterReportLabels(filter.ReportLabel).
		Exclude(filter, "f").
		Group("f.source", "f.severity").
		Order("f.source").
		Scan(ctx, &results)

	return results, err
}

/////////////////////////
/// Lifecycle Methods ///
/////////////////////////

func (s *Store) CreateSchemas(ctx context.Context) error {
	if s.db.Dialect().Name() == dialect.SQLite {
		if _, err := s.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return err
		}
	}

	_, err := s.db.
		NewCreateTable().
		IfNotExists().
		Model((*Config)(nil)).
		Exec(ctx)
	logOnError("create policy_report table", err)

	_, err = s.db.
		NewCreateTable().
		IfNotExists().
		Model((*PolicyReport)(nil)).
		Exec(ctx)
	logOnError("create policy_report table", err)

	_, err = s.db.
		NewCreateTable().
		IfNotExists().
		Model((*PolicyReportResult)(nil)).
		ForeignKey(`(policy_report_id) REFERENCES policy_report(id) ON DELETE CASCADE`).
		Exec(ctx)
	logOnError("create policy_report_result table", err)

	_, err = s.db.
		NewCreateTable().
		IfNotExists().
		Model((*PolicyReportFilter)(nil)).
		ForeignKey(`(policy_report_id) REFERENCES policy_report(id) ON DELETE CASCADE`).
		Exec(ctx)
	logOnError("create policy_report_filter table", err)

	_, err = s.db.
		NewCreateTable().
		IfNotExists().
		Model((*ResourceResult)(nil)).
		ForeignKey(`(policy_report_id) REFERENCES policy_report(id) ON DELETE CASCADE`).
		Exec(ctx)
	logOnError("create policy_report_resource table", err)

	return err
}

func (s *Store) DropSchema(ctx context.Context) error {
	_, err := s.db.NewDropTable().
		IfExists().
		Model((*Config)(nil)).
		Exec(ctx)
	logOnError("drop policy_report_config table", err)

	_, err = s.db.NewDropTable().
		IfExists().
		Model((*PolicyReportFilter)(nil)).
		Exec(ctx)
	logOnError("drop policy_report_filter table", err)

	_, err = s.db.NewDropTable().
		IfExists().
		Model((*PolicyReportResult)(nil)).
		Exec(ctx)
	logOnError("drop policy_report_result table", err)

	_, err = s.db.NewDropTable().
		IfExists().
		Model((*PolicyReport)(nil)).
		Exec(ctx)
	logOnError("drop policy_report table", err)

	_, err = s.db.NewDropTable().
		IfExists().
		Model((*ResourceResult)(nil)).
		Exec(ctx)
	logOnError("drop policy_report_resource table", err)

	return err
}

func (s *Store) Add(ctx context.Context, report v1alpha2.ReportInterface) error {
	_, err := s.db.NewInsert().Model(MapPolicyReport(report)).Exec(ctx)
	if err != nil {
		zap.L().Error("failed to persist policy report", zap.Error(err))
	}

	filters := chunkSlice(MapPolicyReportFilter(report), 50)
	for _, list := range filters {
		_, err = s.db.NewInsert().Ignore().Model(&list).Exec(ctx)
		if err != nil {
			zap.L().Error("failed to bulk import policy report filter", zap.Error(err))
			return err
		}
	}

	resources := chunkSlice(MapPolicyReportResource(report), 50)
	for _, list := range resources {
		_, err = s.db.NewInsert().Model(&list).Exec(ctx)
		if err != nil {
			zap.L().Error("failed to bulk import policy report resources", zap.Error(err))
			return err
		}
	}

	results := chunkSlice(MapPolicyReportResults(report), 50)
	for _, list := range results {
		_, err = s.db.NewInsert().Ignore().Model(&list).Exec(ctx)
		if err != nil {
			zap.L().Error("failed to bulk import policy report results", zap.Error(err))
			return err
		}
	}

	return err
}

func (s *Store) Update(ctx context.Context, report v1alpha2.ReportInterface) error {
	err := s.Remove(ctx, report.GetID())
	if err != nil {
		return err
	}

	s.Add(ctx, report)

	return err
}

func (s *Store) Remove(ctx context.Context, id string) error {
	_, err := s.db.NewDelete().Model((*PolicyReport)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		zap.L().Error("failed to remove previews policy report", zap.Error(err))
	}

	return err
}

func (s *Store) CleanUp(ctx context.Context) error {
	_, err := s.db.NewDelete().Model((*PolicyReport)(nil)).Where("id is not null").Exec(ctx)
	if err != nil {
		zap.L().Error("failed to remove policy reports", zap.Error(err))
	}

	return err
}

func (s *Store) Get(ctx context.Context, id string) (v1alpha2.ReportInterface, error) {
	polr := &PolicyReport{}

	err := s.db.NewSelect().Model(polr).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		zap.L().Error("failed to load policy report", zap.Error(err))
		return nil, err
	}

	results, err := s.fetchResults(ctx, id)
	if err != nil {
		zap.L().Error("failed to load policy report results", zap.Error(err))
		return nil, err
	}

	return &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              polr.Name,
			Namespace:         polr.Namespace,
			CreationTimestamp: v1.NewTime(time.Unix(polr.Created, 0)),
			Labels:            polr.Labels,
		},
		Summary: v1alpha1.ReportSummary{
			Skip:  polr.Skip,
			Pass:  polr.Pass,
			Warn:  polr.Warn,
			Fail:  polr.Fail,
			Error: polr.Error,
		},
		Results: results,
	}, nil
}

func (s *Store) SQLDialect() dialect.Name {
	return s.db.Dialect().Name()
}

func (s *Store) IsSQLite() bool {
	return s.db.Dialect().Name() == dialect.SQLite
}

func (s *Store) fetchResults(ctx context.Context, id string) ([]v1alpha1.ReportResult, error) {
	polr := []*PolicyReportResult{}

	err := s.db.NewSelect().Model(&polr).Where("policy_report_id = ?", id).Scan(ctx)
	if err != nil {
		zap.L().Error("failed to load policy report results", zap.Error(err))
		return nil, err
	}

	list := make([]v1alpha1.ReportResult, 0, len(polr))
	for _, result := range polr {
		list = append(list, v1alpha1.ReportResult{
			ID:          result.ID,
			Result:      v1alpha1.Result(result.Result),
			Severity:    v1alpha1.ResultSeverity(result.Severity),
			Policy:      result.Policy,
			Rule:        result.Rule,
			Description: result.Message,
			Source:      result.Source,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: result.Resource.APIVersion,
					Kind:       result.Resource.Kind,
					Namespace:  result.Resource.Namespace,
					Name:       result.Resource.Name,
					UID:        types.UID(result.Resource.UID),
				},
			},
			Scored:     result.Scored,
			Properties: result.Properties,
			Category:   result.Category,
			Timestamp: v1.Timestamp{
				Seconds: result.Created,
			},
		})
	}

	return list, nil
}

func (s *Store) RequireSchemaUpgrade(ctx context.Context) bool {
	if s.IsSQLite() {
		return true
	}

	config := Config{}

	err := s.db.NewSelect().Model(&config).Where("id = ?", 1).Scan(ctx)
	if err != nil {
		zap.L().Debug("failed to load config", zap.Error(err))
		return true
	}

	return config.Version != s.version
}

func (s *Store) PersistSchemaVersion(ctx context.Context) error {
	config := Config{
		Version: s.version,
	}

	_, err := s.db.NewInsert().Model(&config).Exec(ctx)
	if err != nil {
		zap.L().Error("failed to persist database version", zap.Error(err))
		return err
	}

	return nil
}

func (s *Store) PrepareDatabase(ctx context.Context) error {
	zap.L().Debug("preparing database")
	if s.RequireSchemaUpgrade(ctx) {
		zap.L().Debug("database schema upgrade started")
		if err := s.DropSchema(ctx); err != nil {
			return err
		}

		if err := s.CreateSchemas(ctx); err != nil {
			return err
		}

		if err := s.PersistSchemaVersion(ctx); err != nil {
			return err
		}
	}

	return s.CleanUp(ctx)
}

func NewStore(db *bun.DB, version string) (*Store, error) {
	if db == nil {
		return nil, errors.New("missing database connection")
	}

	s := &Store{
		db:      db,
		version: version,
	}

	return s, nil
}

func NewSQLiteDB(dbFile string) (*bun.DB, error) {
	sqldb, err := createSQLiteDB(dbFile)
	if err != nil {
		return nil, err
	}

	return bun.NewDB(sqldb, sqlitedialect.New()), nil
}

func createSQLiteDB(dbFile string) (*sql.DB, error) {
	os.Remove(dbFile)
	file, err := os.Create(dbFile)
	if err != nil {
		return nil, err
	}
	file.Close()

	db, err := sql.Open("sqlite3", dbFile+"?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	return db, nil
}

func chunkSlice[K interface{}](slice []K, chunkSize int) [][]K {
	var chunks [][]K
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func logOnError(operation string, err error) {
	if err == nil {
		return
	}

	zap.L().Error("failed to execute db operatopn", zap.String("operatopm", operation), zap.Error(err))
}
