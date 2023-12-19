package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
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

	jsonExtractLayout string
}

var (
	withNamespace    = true
	withClusterscope = false
)

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
		_, err = s.db.NewInsert().Model(&list).Exec(ctx)
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

func (s *Store) FetchPolicyReports(ctx context.Context, filter api.Filter, pagination api.Pagination) ([]*api.PolicyReport, error) {
	list := []*api.PolicyReport{}
	query := s.db.NewSelect().Model((*PolicyReport)(nil))

	s.addFilter(query, filter)
	addPolicyReportFilter(query, filter)
	query.Where(`pr.type = ?`, report.PolicyReportType)

	addPagination(query, pagination)

	err := query.Scan(ctx, &list)
	if err != nil {
		zap.L().Error("failed to select policy report results", zap.Error(err), zap.Any("filter", filter), zap.Any("pagination", pagination))
	}

	return list, err
}

func (s *Store) CountPolicyReports(ctx context.Context, filter api.Filter) (int, error) {
	query := s.db.NewSelect().Model((*PolicyReport)(nil))

	s.addFilter(query, filter)
	addPolicyReportFilter(query, filter)
	query.Where(`pr.type = ?`, report.PolicyReportType)

	count, err := query.Count(ctx)
	if err != nil {
		zap.L().Error("failed to select policy report results", zap.Error(err), zap.Any("filter", filter))
	}

	return count, err
}

func (s *Store) FetchNamespacedReportLabels(ctx context.Context, filter api.Filter) (map[string][]string, error) {
	results := []string{}
	list := make(map[string][]string)

	query := s.db.NewSelect().
		TableExpr("policy_report as pr").
		Distinct().
		Where(`pr.type = ?`, report.PolicyReportType)

	if s.db.Dialect().Name() == dialect.PG {
		query.ColumnExpr("labels::text")
	} else {
		query.Column("labels")
	}

	addPolicyReportFilter(query, filter)

	err := query.Scan(ctx, &results)
	if err != nil {
		return list, err
	}

	for _, labels := range results {
		for key, value := range convertJSONToMap(labels) {
			_, ok := list[key]
			contained := contains(value, list[key])

			if ok && !contained {
				list[key] = append(list[key], value)
				continue
			} else if ok && contained {
				continue
			}

			list[key] = []string{value}
		}
	}

	return list, nil
}

func (s *Store) FetchClusterPolicyReports(ctx context.Context, filter api.Filter, pagination api.Pagination) ([]*api.PolicyReport, error) {
	list := []*api.PolicyReport{}
	query := s.db.NewSelect().Model((*PolicyReport)(nil))

	s.addFilter(query, filter)
	addPolicyReportFilter(query, filter)
	query.Where(`pr.type = ?`, report.ClusterPolicyReportType)

	addPagination(query, pagination)

	err := query.Scan(ctx, &list)
	if err != nil {
		zap.L().Error("failed to select policy report results", zap.Error(err), zap.Any("filter", filter), zap.Any("pagination", pagination))
	}

	return list, err
}

func (s *Store) CountClusterPolicyReports(ctx context.Context, filter api.Filter) (int, error) {
	query := s.db.NewSelect().Model((*PolicyReport)(nil))

	s.addFilter(query, filter)
	addPolicyReportFilter(query, filter)
	query.Where(`pr.type = ?`, report.ClusterPolicyReportType)

	count, err := query.Count(ctx)
	if err != nil {
		zap.L().Error("failed to select policy report results", zap.Error(err), zap.Any("filter", filter))
	}

	return count, err
}

func (s *Store) FetchClusterReportLabels(ctx context.Context, filter api.Filter) (map[string][]string, error) {
	results := []string{}
	list := make(map[string][]string)

	query := s.db.NewSelect().
		TableExpr("policy_report as pr").
		Distinct().
		Where(`pr.type = ?`, report.ClusterPolicyReportType)

	if s.db.Dialect().Name() == dialect.PG {
		query.ColumnExpr("labels::text")
	} else {
		query.Column("labels")
	}

	addPolicyReportFilter(query, filter)

	err := query.Scan(ctx, &results)
	if err != nil {
		return list, err
	}

	for _, labels := range results {
		for key, value := range convertJSONToMap(labels) {
			_, ok := list[key]
			contained := contains(value, list[key])

			if ok && !contained {
				list[key] = append(list[key], value)
				continue
			} else if ok && contained {
				continue
			}

			list[key] = []string{value}
		}
	}

	return list, nil
}

func (s *Store) FetchClusterRules(ctx context.Context, filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_result as r").
		Column("rule").
		Distinct().
		Order("rule ASC").
		Where(`r.resource_namespace = ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	query.Scan(ctx, &list)

	return list, nil
}

func (s *Store) FetchClusterResources(ctx context.Context, filter api.Filter) ([]*api.Resource, error) {
	list := make([]*api.Resource, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_result as r").
		ColumnExpr("resource_name as name, resource_kind as kind").
		Distinct().
		Order("kind ASC", "name ASC").
		Where(`r.resource_namespace = ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	query.Scan(ctx, &list)

	return list, nil
}

func (s *Store) FetchClusterPolicies(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "policy", filter, &withClusterscope)
}

func (s *Store) FetchClusterKinds(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "kind", filter, &withClusterscope)
}

func (s *Store) FetchClusterCategories(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "category", filter, &withClusterscope)
}

func (s *Store) FetchClusterSources(ctx context.Context) ([]string, error) {
	return s.fetchFilterOptions(ctx, "source", api.Filter{}, &withClusterscope)
}

func (s *Store) FetchSources(ctx context.Context, resource string) ([]*api.Source, error) {
	query := s.db.
		NewSelect().
		TableExpr("policy_report_resource as r").
		Distinct().
		ColumnExpr("source, category, SUM(pass) as pass, SUM(warn) as warn, SUM(fail) as fail, SUM(error) as error, SUM(skip) as skip").
		Group("source", "category").
		Order("source ASC", "category ASC")

	if resource != "" {
		query.Where("id = ?", resource)
	}

	results := make([]*ResourceResult, 0)

	err := query.Scan(ctx, &results)
	if err != nil {
		zap.L().Error("failed to load cluster status counts", zap.Error(err))
		return nil, err
	}

	list := make(map[string]*api.Source, 0)
	for _, r := range results {
		if s, ok := list[r.Source]; ok {
			s.Categories = append(s.Categories, api.Category{
				Name:  r.Category,
				Pass:  r.Pass,
				Fail:  r.Fail,
				Warn:  r.Warn,
				Error: r.Error,
				Skip:  r.Skip,
			})
			continue
		}

		list[r.Source] = &api.Source{
			Name: r.Source,
			Categories: []api.Category{{
				Name:  r.Category,
				Pass:  r.Pass,
				Fail:  r.Fail,
				Warn:  r.Warn,
				Error: r.Error,
				Skip:  r.Skip,
			}},
		}
	}

	return helper.ToList(list), nil
}

func (s *Store) FetchFindingCounts(ctx context.Context, filter api.Filter) (*api.Findings, error) {
	query := s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.result as status, source").
		Where("status IN (?)", bun.In([]string{v1alpha2.StatusPass, v1alpha2.StatusFail, v1alpha2.StatusWarn, v1alpha2.StatusError})).
		Group("status", "source")

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = f.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportFilterFilter(query, filter)

	results := make([]api.StatusCount, 0)

	err := query.Scan(ctx, &results)
	if err != nil {
		zap.L().Error("failed to load cluster status counts", zap.Error(err))
		return nil, err
	}

	findings := make(map[string]*api.FindingCounts, 0)
	total := 0
	for _, count := range results {
		if finding, ok := findings[count.Source]; ok {
			finding.Counts[count.Status] = count.Count
			finding.Total = finding.Total + count.Count
		} else {
			findings[count.Source] = &api.FindingCounts{
				Source: count.Source,
				Total:  count.Count,
				Counts: map[string]int{
					count.Status: count.Count,
				},
			}
		}

		total += count.Count
	}

	return &api.Findings{Counts: helper.ToList(findings), Total: total}, nil
}

func (s *Store) FetchClusterStatusCounts(ctx context.Context, filter api.Filter) ([]api.StatusCount, error) {
	var list map[string]api.StatusCount

	if len(filter.Status) == 0 {
		list = map[string]api.StatusCount{
			v1alpha2.StatusPass:  {Status: v1alpha2.StatusPass},
			v1alpha2.StatusFail:  {Status: v1alpha2.StatusFail},
			v1alpha2.StatusWarn:  {Status: v1alpha2.StatusWarn},
			v1alpha2.StatusError: {Status: v1alpha2.StatusError},
			v1alpha2.StatusSkip:  {Status: v1alpha2.StatusSkip},
		}
	} else {
		list = map[string]api.StatusCount{}

		for _, status := range filter.Status {
			list[status] = api.StatusCount{Status: status}
		}
	}

	counts := make([]api.StatusCount, 0, len(list))
	results := make([]api.StatusCount, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.result as status").
		Where(`f.namespace = ''`).
		Group("status")

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = f.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportFilterFilter(query, filter)

	err := query.Scan(ctx, &results)
	if err != nil {
		zap.L().Error("failed to load cluster status counts", zap.Error(err))
		return nil, err
	}

	for _, count := range results {
		list[count.Status] = count
	}

	for _, count := range list {
		counts = append(counts, count)
	}

	return counts, nil
}

func (s *Store) FetchClusterResults(ctx context.Context, filter api.Filter, pagination api.Pagination) ([]*api.ListResult, error) {
	results := make([]*PolicyReportResult, 0)

	query := s.db.
		NewSelect().
		Model(&results).
		Where(`r.resource_namespace = ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)
	addPagination(query, pagination)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return MapListResult(results), nil
}

func (s *Store) CountClusterResults(ctx context.Context, filter api.Filter) (int, error) {
	query := s.db.
		NewSelect().
		Model((*PolicyReportResult)(nil)).
		Where(`r.resource_namespace = ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	return query.Count(ctx)
}

func (s *Store) FetchNamespacedRules(ctx context.Context, filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_result as r").
		Column("rule").
		Distinct().
		Order("rule ASC").
		Where(`r.resource_namespace != ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	query.Scan(ctx, &list)

	return list, nil
}

func (s *Store) FetchNamespacedResources(ctx context.Context, filter api.Filter) ([]*api.Resource, error) {
	list := make([]*api.Resource, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_result as r").
		ColumnExpr("resource_name as name, resource_kind as kind").
		Distinct().
		Order("kind ASC", "name ASC").
		Where(`r.resource_namespace != ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	query.Scan(ctx, &list)

	return list, nil
}

func (s *Store) FetchNamespacedPolicies(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "policy", filter, &withNamespace)
}

func (s *Store) FetchNamespacedKinds(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "kind", filter, &withNamespace)
}

func (s *Store) FetchNamespacedCategories(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "category", filter, &withNamespace)
}

func (s *Store) FetchNamespacedSources(ctx context.Context) ([]string, error) {
	return s.fetchFilterOptions(ctx, "source", api.Filter{}, &withNamespace)
}

func (s *Store) FetchNamespaces(ctx context.Context, filter api.Filter) ([]string, error) {
	return s.fetchFilterOptions(ctx, "f.namespace", filter, &withNamespace)
}

func (s *Store) FetchNamespacedStatusCounts(ctx context.Context, filter api.Filter) ([]api.NamespacedStatusCount, error) {
	var list map[string][]api.NamespaceCount

	if len(filter.Status) == 0 {
		list = map[string][]api.NamespaceCount{
			v1alpha2.StatusPass:  make([]api.NamespaceCount, 0),
			v1alpha2.StatusFail:  make([]api.NamespaceCount, 0),
			v1alpha2.StatusWarn:  make([]api.NamespaceCount, 0),
			v1alpha2.StatusError: make([]api.NamespaceCount, 0),
			v1alpha2.StatusSkip:  make([]api.NamespaceCount, 0),
		}
	} else {
		list = map[string][]api.NamespaceCount{}

		for _, status := range filter.Status {
			list[status] = make([]api.NamespaceCount, 0)
		}
	}

	statusCounts := make([]api.NamespacedStatusCount, 0, 5)
	counts := make([]api.NamespaceCount, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		ColumnExpr("SUM(f.count) as count, f.namespace, f.result as status").
		Where(`f.namespace != ''`).
		Group("f.namespace", "status").
		Order("f.namespace ASC")

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = f.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportFilterFilter(query, filter)

	err := query.Scan(ctx, &counts)
	if err != nil {
		zap.L().Error("failed to load namespaced status counts", zap.Error(err))
		return nil, err
	}

	for _, count := range counts {
		list[count.Status] = append(list[count.Status], count)
	}

	for status, items := range list {
		statusCounts = append(statusCounts, api.NamespacedStatusCount{
			Status: status,
			Items:  items,
		})
	}

	return statusCounts, nil
}

func (s *Store) FetchRuleStatusCounts(ctx context.Context, policy, rule string) ([]api.StatusCount, error) {
	list := map[string]api.StatusCount{
		v1alpha2.StatusPass:  {Status: v1alpha2.StatusPass},
		v1alpha2.StatusFail:  {Status: v1alpha2.StatusFail},
		v1alpha2.StatusWarn:  {Status: v1alpha2.StatusWarn},
		v1alpha2.StatusError: {Status: v1alpha2.StatusError},
		v1alpha2.StatusSkip:  {Status: v1alpha2.StatusSkip},
	}

	statusCounts := make([]api.StatusCount, 0, len(list))
	counts := make([]api.StatusCount, 0)

	err := s.db.NewSelect().
		Table("policy_report_result").
		ColumnExpr("COUNT(id) as count, result as status").
		Where("rule = ?", rule).
		Where("policy = ?", policy).
		Group("status").
		Scan(ctx, &counts)
	if err != nil {
		return statusCounts, err
	}

	for _, count := range counts {
		list[count.Status] = count
	}

	for _, count := range list {
		statusCounts = append(statusCounts, count)
	}

	return statusCounts, nil
}

func (s *Store) FetchNamespacedResults(ctx context.Context, filter api.Filter, pagination api.Pagination) ([]*api.ListResult, error) {
	results := make([]*PolicyReportResult, 0)

	query := s.db.
		NewSelect().
		Model(&results).
		Where(`r.resource_namespace != ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)
	addPagination(query, pagination)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return MapListResult(results), nil
}

func (s *Store) CountNamespacedResults(ctx context.Context, filter api.Filter) (int, error) {
	query := s.db.
		NewSelect().
		Model((*PolicyReportResult)(nil)).
		Where(`r.resource_namespace != ''`)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	return query.Count(ctx)
}

func (s *Store) FetchResults(ctx context.Context, id string, filter api.Filter, pagination api.Pagination) ([]*api.ListResult, error) {
	results := make([]*PolicyReportResult, 0)

	query := s.db.
		NewSelect().
		Model(&results).
		Where(`r.resource_id = ?`, id)

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)
	addPagination(query, pagination)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return MapListResult(results), nil
}

func (s *Store) CountResults(ctx context.Context, id string, filter api.Filter) (int, error) {
	query := s.db.
		NewSelect().
		Model((*PolicyReportResult)(nil)).
		Where(`r.resource_id = ?`, id)

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = r.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	return query.Count(ctx)
}

func (s *Store) FetchNamespacedResourceResults(ctx context.Context, filter api.Filter, pagination api.Pagination) ([]*api.ResourceResult, error) {
	results := make([]*ResourceResult, 0)

	query := s.db.
		NewSelect().
		Model(&results).
		ColumnExpr("id, resource_uid, resource_kind, resource_api_version, resource_namespace, resource_name, SUM(pass) as pass, SUM(warn) as warn, SUM(fail) as fail, SUM(error) as error, SUM(skip) as skip").
		Group("id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name").
		Where(`resource_namespace != ''`)

	addPolicyReportResourceFilter(query, filter)
	addPagination(query, pagination)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return MapResourceResult(results), nil
}

func (s *Store) CountNamespacedResourceResults(ctx context.Context, filter api.Filter) (int, error) {
	query := s.db.
		NewSelect().
		Model((*ResourceResult)(nil)).
		Group("id").
		Where(`resource_namespace != ''`)

	addPolicyReportResourceFilter(query, filter)

	return query.Count(ctx)
}

func (s *Store) FetchClusterResourceResults(ctx context.Context, filter api.Filter, pagination api.Pagination) ([]*api.ResourceResult, error) {
	results := make([]*ResourceResult, 0)

	query := s.db.
		NewSelect().
		Model(&results).
		ColumnExpr("id, resource_uid, resource_kind, resource_api_version, resource_namespace, resource_name, SUM(pass) as pass, SUM(warn) as warn, SUM(fail) as fail, SUM(error) as error, SUM(skip) as skip").
		Where(`resource_namespace = ''`).
		Group("id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name")

	addPolicyReportResourceFilter(query, filter)
	addPagination(query, pagination)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return MapResourceResult(results), nil
}

func (s *Store) CountClusterResourceResults(ctx context.Context, filter api.Filter) (int, error) {
	query := s.db.
		NewSelect().
		Model((*ResourceResult)(nil)).
		Where(`resource_namespace = ''`).
		Group("id")

	addPolicyReportResourceFilter(query, filter)

	return query.Count(ctx)
}

func (s *Store) FetchResourceResults(ctx context.Context, id string, filter api.Filter) ([]*api.ResourceResult, error) {
	results := make([]*ResourceResult, 0)

	query := s.db.
		NewSelect().
		Model(&results).
		ColumnExpr("id, resource_uid, resource_kind, resource_api_version, resource_namespace, resource_name, source, SUM(pass) as pass, SUM(warn) as warn, SUM(fail) as fail, SUM(error) as error, SUM(skip) as skip").
		Where(`id = ?`, id).
		Order("source ASC").
		Group("id", "resource_uid", "resource_kind", "resource_api_version", "resource_namespace", "resource_name", "source")

	addPolicyReportResourceFilter(query, filter)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return MapResourceResult(results), nil
}

func (s *Store) FetchResource(ctx context.Context, id string) (*api.Resource, error) {
	result := ResourceResult{}

	query := s.db.
		NewSelect().
		Model(&result).
		ColumnExpr("id, resource_uid, resource_kind, resource_api_version, resource_namespace, resource_name").
		Where(`id = ?`, id).
		Limit(1)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &api.Resource{
		ID:         result.ID,
		UID:        result.Resource.UID,
		APIVersion: result.Resource.APIVersion,
		Kind:       result.Resource.Kind,
		Name:       result.Resource.Name,
		Namespace:  result.Resource.Namespace,
	}, nil
}

func (s *Store) FetchResourceStatusCounts(ctx context.Context, resourceID string, filter api.Filter) ([]api.ResourceStatusCount, error) {
	var list map[string][]api.ResourceCount

	if len(filter.Status) == 0 {
		list = map[string][]api.ResourceCount{
			v1alpha2.StatusPass:  make([]api.ResourceCount, 0),
			v1alpha2.StatusFail:  make([]api.ResourceCount, 0),
			v1alpha2.StatusWarn:  make([]api.ResourceCount, 0),
			v1alpha2.StatusError: make([]api.ResourceCount, 0),
			v1alpha2.StatusSkip:  make([]api.ResourceCount, 0),
		}
	} else {
		list = map[string][]api.ResourceCount{}

		for _, status := range filter.Status {
			list[status] = make([]api.ResourceCount, 0)
		}
	}

	statusCounts := make([]api.ResourceStatusCount, 0, 5)
	counts := make([]api.ResourceCount, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_result as r").
		ColumnExpr("COUNT(r.resource_id) as count, r.result as status, r.source").
		Where(`r.resource_id = ?`, resourceID).
		Group("status", "source").
		Order("source ASC")

	s.addFilter(query, filter)
	addPolicyReportResultFilter(query, filter)

	err := query.Scan(ctx, &counts)
	if err != nil {
		zap.L().Error("failed to load resource status counts", zap.Error(err))
		return nil, err
	}

	for _, count := range counts {
		list[count.Status] = append(list[count.Status], count)
	}

	for status, items := range list {
		statusCounts = append(statusCounts, api.ResourceStatusCount{
			Status: status,
			Items:  items,
		})
	}

	return statusCounts, nil
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

	return &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              polr.Name,
			Namespace:         polr.Namespace,
			CreationTimestamp: v1.NewTime(time.Unix(polr.Created, 0)),
			Labels:            polr.Labels,
		},
		Summary: v1alpha2.PolicyReportSummary{
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

func (s *Store) fetchResults(ctx context.Context, id string) ([]v1alpha2.PolicyReportResult, error) {
	polr := []*PolicyReportResult{}

	err := s.db.NewSelect().Model(&polr).Where("policy_report_id = ?", id).Scan(ctx)
	if err != nil {
		zap.L().Error("failed to load policy report results", zap.Error(err))
		return nil, err
	}

	list := make([]v1alpha2.PolicyReportResult, 0, len(polr))
	for _, result := range polr {
		list = append(list, v1alpha2.PolicyReportResult{
			ID:       result.ID,
			Result:   v1alpha2.PolicyResult(result.Result),
			Severity: v1alpha2.PolicySeverity(result.Severity),
			Policy:   result.Policy,
			Rule:     result.Rule,
			Message:  result.Message,
			Source:   result.Source,
			Resources: []corev1.ObjectReference{
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

func (s *Store) fetchFilterOptions(ctx context.Context, option string, filter api.Filter, namespaced *bool) ([]string, error) {
	list := make([]string, 0)

	query := s.db.
		NewSelect().
		TableExpr("policy_report_filter as f").
		Column(option).
		Distinct().
		Order(option+" ASC").
		Where(`? != ''`, bun.Ident(option))

	if *namespaced == true {
		query.Where(`f.namespace != ''`)
	} else if *namespaced == false {
		query.Where(`f.namespace = ''`)
	}

	if len(filter.ReportLabel) > 0 {
		query.Join("JOIN policy_report AS pr ON pr.id = f.policy_report_id")
	}

	s.addFilter(query, filter)
	addPolicyReportFilterFilter(query, filter)

	err := query.Scan(ctx, &list)

	return list, err
}

func (s *Store) Configure() {
	if s.db.Dialect().Name() == dialect.PG {
		s.jsonExtractLayout = "(pr.labels->>'%s') = ?"
		return
	}

	s.jsonExtractLayout = "json_extract(pr.labels, '$.\"%s\"') = ?"
}

func (s *Store) RequireSchemaUpgrade(ctx context.Context) bool {
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

	s.Configure()

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

func addPolicyReportFilterFilter(query *bun.SelectQuery, filter api.Filter) {
	if len(filter.Namespaces) > 0 {
		query.Where("f.namespace IN (?)", bun.In(filter.Namespaces))
	} else if len(filter.Namespaces) == 1 {
		query.Where("f.namespace = ?", filter.Namespaces[0])
	}

	if len(filter.Kinds) > 0 {
		query.Where("f.kind IN (?)", bun.In(filter.Kinds))
	} else if len(filter.Kinds) == 1 {
		query.Where("f.kind = ?", filter.Kinds[0])
	}

	if len(filter.Sources) > 0 {
		query.Where("f.source IN (?)", bun.In(filter.Sources))
	} else if len(filter.Sources) == 1 {
		query.Where("f.source = ?", filter.Sources[0])
	}

	if len(filter.Kinds) == 0 && len(filter.Exclude) > 0 {
		for source, kind := range filter.Exclude {
			query.Where("(f.source != ? OR (f.source = ? AND f.kind NOT IN (?)))", source, source, bun.In(kind))
		}
	}
}

func addPolicyReportResultFilter(query *bun.SelectQuery, filter api.Filter) {
	if len(filter.Namespaces) > 1 {
		query.Where("r.resource_namespace IN (?)", bun.In(filter.Namespaces))
	} else if len(filter.Namespaces) == 1 {
		query.Where("r.resource_namespace = ?", filter.Namespaces[0])
	}

	if len(filter.Rules) > 0 {
		query.Where("r.rule IN (?)", bun.In(filter.Rules))
	} else if len(filter.ReportLabel) == 1 {
		query.Where("r.rule = ?", filter.Rules[0])
	}

	if len(filter.Kinds) > 0 {
		query.Where("r.resource_kind IN (?)", bun.In(filter.Kinds))
	} else if len(filter.Kinds) == 1 {
		query.Where("r.resource_kind = ?", filter.Kinds[0])
	}

	if len(filter.Resources) > 0 {
		query.Where("r.resource_name IN (?)", bun.In(filter.Resources))
	} else if len(filter.Resources) == 1 {
		query.Where("r.resource_name = ?", filter.Resources[0])
	}

	if filter.ResourceID != "" {
		query.Where("r.resource_id = ?", filter.ResourceID)
	}

	if len(filter.Sources) > 0 {
		query.Where("r.source IN (?)", bun.In(filter.Sources))
	} else if len(filter.Sources) == 1 {
		query.Where("r.source = ?", filter.Sources[0])
	}

	if len(filter.Kinds) == 0 && len(filter.Exclude) > 0 {
		for source, kind := range filter.Exclude {
			query.Where("(r.source != ? OR (r.source = ? AND r.resource_kind NOT IN (?)))", source, source, bun.In(kind))
		}
	}

	if filter.Search != "" {
		query.Where(`(resource_namespace LIKE ?0 OR resource_name LIKE ?0 OR policy LIKE ?0 OR rule LIKE ?0 OR severity = ?1 OR result = ?1 OR LOWER(resource_kind) = LOWER(?1))`, "%"+filter.Search+"%", filter.Search)
	}
}

func addPolicyReportResourceFilter(query *bun.SelectQuery, filter api.Filter) {
	if len(filter.Namespaces) > 0 {
		query.Where("res.resource_namespace IN (?)", bun.In(filter.Namespaces))
	} else if len(filter.Namespaces) == 1 {
		query.Where("res.resource_namespace = ?", filter.Namespaces[0])
	}

	if len(filter.Kinds) > 0 {
		query.Where("res.resource_kind IN (?)", bun.In(filter.Kinds))
	} else if len(filter.Kinds) == 1 {
		query.Where("res.resource_kind = ?", filter.Kinds[0])
	}

	if len(filter.Resources) > 0 {
		query.Where("res.resource_name IN (?)", bun.In(filter.Resources))
	} else if len(filter.Resources) == 1 {
		query.Where("res.resource_name = ?", filter.Resources[0])
	}

	if len(filter.Sources) > 0 {
		query.Where("res.source IN (?)", bun.In(filter.Sources))
	} else if len(filter.Sources) == 1 {
		query.Where("res.source = ?", filter.Sources[0])
	}

	if len(filter.Categories) > 0 {
		query.Where("res.category IN (?)", bun.In(filter.Categories))
	} else if len(filter.Categories) == 1 {
		query.Where("res.category = ?", filter.Categories[0])
	}

	if filter.ResourceID != "" {
		query.Where("res.id = ?", filter.ResourceID)
	}

	if len(filter.Kinds) == 0 && len(filter.Exclude) > 0 {
		for source, kind := range filter.Exclude {
			query.Where("(res.source != ? OR (res.source = ? AND res.resource_kind NOT IN (?)))", source, source, bun.In(kind))
		}
	}

	if filter.Search != "" {
		query.Where(`(resource_namespace LIKE ?0 OR resource_name LIKE ?0 OR LOWER(resource_kind) = LOWER(?1))`, "%"+filter.Search+"%", filter.Search)
	}
}

func addPolicyReportFilter(query *bun.SelectQuery, filter api.Filter) {
	if len(filter.Namespaces) > 0 {
		query.Where("pr.namespace IN (?)", bun.In(filter.Namespaces))
	} else if len(filter.Namespaces) == 1 {
		query.Where("pr.namespace = ?", filter.Namespaces[0])
	}

	if len(filter.Sources) > 0 {
		query.Where("pr.source IN (?)", bun.In(filter.Sources))
	} else if len(filter.Sources) == 1 {
		query.Where("pr.source = ?", filter.Sources[0])
	}
}

func (s *Store) addFilter(query *bun.SelectQuery, filter api.Filter) {
	if len(filter.Policies) > 0 {
		query.Where("policy IN (?)", bun.In(filter.Policies))
	} else if len(filter.Policies) == 1 {
		query.Where("policy = ?", filter.Policies[0])
	}

	if len(filter.Categories) > 0 {
		query.Where("category IN (?)", bun.In(filter.Categories))
	} else if len(filter.Categories) == 1 {
		query.Where("category = ?", filter.Categories[0])
	}

	if len(filter.Severities) > 0 {
		query.Where("severity IN (?)", bun.In(filter.Severities))
	} else if len(filter.Severities) == 1 {
		query.Where("severity = ?", filter.Severities[0])
	}

	if len(filter.Status) > 0 {
		query.Where("result IN (?)", bun.In(filter.Status))
	} else if len(filter.Status) == 1 {
		query.Where("result = ?", filter.Status[0])
	}

	if len(filter.ReportLabel) > 0 {
		for key, value := range filter.ReportLabel {
			query.Where(fmt.Sprintf(s.jsonExtractLayout, key), value)
		}
	}
}

func addPagination(query *bun.SelectQuery, pagination api.Pagination) {
	query.OrderExpr(fmt.Sprintf(
		"%s %s",
		strings.Join(pagination.SortBy, ","),
		pagination.Direction,
	))

	if pagination.Page == 0 || pagination.Offset == 0 {
		return
	}

	query.Limit(pagination.Offset)
	query.Offset((pagination.Page - 1) * pagination.Offset)
}

func convertJSONToMap(s string) map[string]string {
	m := make(map[string]string)
	if s == "" {
		return m
	}

	_ = json.Unmarshal([]byte(s), &m)

	return m
}

func contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
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
