package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

const (
	reportSQL = `CREATE TABLE policy_report (
    "id" TEXT NOT NULL PRIMARY KEY,   
    "type" TEXT,
    "namespace" TEXT,
    "name" TEXT NOT NULL,
    "source" TEXT,
    "labels" JSON DEFAULT '{}',
    "skip" INTEGER DEFAULT 0,
    "pass" INTEGER DEFAULT 0,
    "warn" INTEGER DEFAULT 0,
    "fail" INTEGER DEFAULT 0,
    "error" INTEGER DEFAULT 0,
    "created" INTEGER
  );`

	resultSQL = `CREATE TABLE policy_report_result (
    "policy_report_id" TEXT NOT NULL,
    "id" TEXT NOT NULL,
    "policy" TEXT,
    "rule" TEXT,
    "message" TEXT,
    "scored" INTEGER,
    "status" TEXT,
    "severity" TEXT,
    "category" TEXT,
    "source" TEXT,
    "resource_api_version" TEXT,
    "resource_kind" TEXT,
    "resource_name" TEXT,
    "resource_namespace" TEXT,
    "resource_uid" TEXT,
    "properties" TEXT,
    "timestamp" INTEGER,
	PRIMARY KEY (policy_report_id, id),
    FOREIGN KEY (policy_report_id) REFERENCES policy_report(id) ON DELETE CASCADE
  );`

	policySQL = `CREATE TABLE policy_report_filter (
	"policy_report_id" TEXT NOT NULL,
	"namespace" TEXT,
	"policy" TEXT,
	"kind" TEXT,
	"category" TEXT,
	"severity" TEXT,
	"status" TEXT,
	"source" TEXT,
	"count" INTEGER
  );`

	resultInsertBaseSQL = "INSERT OR IGNORE INTO policy_report_result(policy_report_id, id, policy, rule, message, scored, status, severity, category, source, resource_api_version, resource_kind, resource_name, resource_namespace, resource_uid, properties, timestamp) VALUES "
	filterInsertBaseSQL = "INSERT OR IGNORE INTO policy_report_filter(policy_report_id, namespace, policy, status, severity, category, source, kind, count) VALUES "
)

type PolicyReportStore interface {
	report.PolicyReportStore
	api.PolicyReportFinder
}

// policyReportStore caches the latest version of an PolicyReport
type policyReportStore struct {
	db *sql.DB
}

func (s *policyReportStore) CreateSchemas() error {
	_, err := s.db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	_, err = s.db.Exec(reportSQL)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(policySQL)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(resultSQL)

	return err
}

// Get an PolicyReport by Type and ID
func (s *policyReportStore) Get(id string) (v1alpha2.ReportInterface, bool) {
	var created int64
	var labels string

	r := &v1alpha2.PolicyReport{
		Summary: v1alpha2.PolicyReportSummary{},
	}

	row := s.db.QueryRow("SELECT namespace, name, labels, pass, skip, warn, fail, error, created FROM policy_report WHERE id=$1", id)
	err := row.Scan(&r.Namespace, &r.Name, &labels, &r.Summary.Pass, &r.Summary.Skip, &r.Summary.Warn, &r.Summary.Fail, &r.Summary.Error, &created)
	if err == sql.ErrNoRows {
		return r, false
	} else if err != nil {
		log.Printf("[ERROR] failed to select PolicyReport: %s", err)
		return r, false
	}

	r.CreationTimestamp = v1.NewTime(time.Unix(created, 0))
	r.Labels = convertJSONToMap(labels)

	results, err := s.fetchResults(id)
	if err != nil {
		log.Printf("[ERROR] failed to fetch Reports: %s\n", err)
		return r, false
	}

	r.Results = results

	return r, true
}

// Add a PolicyReport to the Store
func (s *policyReportStore) Add(r v1alpha2.ReportInterface) error {
	stmt, err := s.db.Prepare("INSERT INTO policy_report(id, type, namespace, source, name, labels, pass, skip, warn, fail, error, created) values(?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	sum := r.GetSummary()

	_, err = stmt.Exec(
		r.GetID(),
		report.GetType(r),
		r.GetNamespace(),
		r.GetSource(),
		r.GetName(),
		convertMapToJSON(r.GetLabels()),
		sum.Pass,
		sum.Skip,
		sum.Warn,
		sum.Fail,
		sum.Error,
		r.GetCreationTimestamp().Unix(),
	)
	if err != nil {
		return err
	}

	err = s.persistFilterValues(r)
	if err != nil {
		return err
	}

	return s.persistResults(r)
}

func (s *policyReportStore) Update(r v1alpha2.ReportInterface) error {
	stmt, err := s.db.Prepare("UPDATE policy_report SET labels=?, pass=?, skip=?, warn=?, fail=?, error=?, created=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	sum := r.GetSummary()

	_, err = stmt.Exec(
		convertMapToJSON(r.GetLabels()),
		sum.Pass,
		sum.Skip,
		sum.Warn,
		sum.Fail,
		sum.Error,
		r.GetCreationTimestamp().Unix(),
		r.GetID(),
	)
	if err != nil {
		return err
	}

	fstmt, err := s.db.Prepare("DELETE FROM policy_report_filter WHERE policy_report_id=?")
	if err != nil {
		return err
	}
	defer fstmt.Close()

	_, err = fstmt.Exec(r.GetID())
	if err != nil {
		return err
	}

	err = s.persistFilterValues(r)
	if err != nil {
		return err
	}

	dstmt, err := s.db.Prepare("DELETE FROM policy_report_result WHERE policy_report_id=?")
	if err != nil {
		return err
	}
	defer dstmt.Close()

	_, err = dstmt.Exec(r.GetID())
	if err != nil {
		return err
	}

	return s.persistResults(r)
}

// Remove a PolicyReport with the given Type and ID from the Store
func (s *policyReportStore) Remove(id string) error {
	stmt, err := s.db.Prepare("DELETE FROM policy_report WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	fstmt, err := s.db.Prepare("DELETE FROM policy_report_filter WHERE policy_report_id=?")
	if err != nil {
		return err
	}
	defer fstmt.Close()

	_, err = fstmt.Exec(id)
	if err != nil {
		return err
	}

	stmt, err = s.db.Prepare("DELETE FROM policy_report_result WHERE policy_report_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}

func (s *policyReportStore) CleanUp() error {
	stmt, err := s.db.Prepare("DELETE FROM policy_report")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	fstmt, err := s.db.Prepare("DELETE FROM policy_report_filter")
	if err != nil {
		return err
	}
	defer fstmt.Close()

	_, err = fstmt.Exec()
	if err != nil {
		return err
	}

	dstmt, err := s.db.Prepare("DELETE FROM policy_report_result")
	if err != nil {
		return err
	}
	defer dstmt.Close()

	_, err = dstmt.Exec()
	return err
}

// FetchPolicyReports by filter and pagination
func (s *policyReportStore) FetchPolicyReports(filter api.Filter, pagination api.Pagination) ([]*api.PolicyReport, error) {
	whereParts := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int
	var where string

	if len(filter.Namespaces) > 0 {
		argCounter, whereParts, args = appendWhere(filter.Namespaces, "namespace", whereParts, args, argCounter)
	} else {
		whereParts = append(whereParts, `namespace != ""`)
	}

	if len(filter.ReportLabel) > 0 {
		for key, value := range filter.ReportLabel {
			argCounter++

			whereParts = append(whereParts, fmt.Sprintf("json_extract(labels, '$.\"%s\"') = $%d", key, argCounter))
			args = append(args, value)
		}
	}

	paginationString := generatePagination(pagination)
	where = strings.Join(whereParts, " AND ")

	list := make([]*api.PolicyReport, 0)

	rows, err := s.db.Query(`SELECT id, namespace, source, name, labels, pass, skip, warn, fail, error FROM policy_report as report WHERE `+where+" "+paginationString, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var labels string
		r := &api.PolicyReport{}
		err := rows.Scan(&r.ID, &r.Namespace, &r.Source, &r.Name, &labels, &r.Pass, &r.Skip, &r.Warn, &r.Fail, &r.Error)
		if err != nil {
			log.Printf("[ERROR] failed to scan PolicyReport: %s", err)
			return list, err
		}

		r.Labels = convertJSONToMap(labels)

		list = append(list, r)
	}

	return list, nil
}

// CountPolicyReports by filter
func (s *policyReportStore) CountPolicyReports(filter api.Filter) (int, error) {
	whereParts := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int
	var where string

	if len(filter.Namespaces) > 0 {
		argCounter, whereParts, args = appendWhere(filter.Namespaces, "namespace", whereParts, args, argCounter)
	} else {
		whereParts = append(whereParts, `namespace != ""`)
	}

	if len(filter.ReportLabel) > 0 {
		for key, value := range filter.ReportLabel {
			argCounter++

			whereParts = append(whereParts, fmt.Sprintf("json_extract(labels, '$.\"%s\"') = $%d", key, argCounter))
			args = append(args, value)
		}
	}

	where = strings.Join(whereParts, " AND ")
	var count int

	row := s.db.QueryRow(`SELECT count(id) FROM policy_report as report WHERE `+where, args...)
	err := row.Scan(&count)
	if err != nil {
		return count, err
	}

	return count, nil
}

// FetchClusterPolicyReports by filter and pagination
func (s *policyReportStore) FetchClusterPolicyReports(filter api.Filter, pagination api.Pagination) ([]*api.PolicyReport, error) {
	whereParts := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int
	var where string

	whereParts = append(whereParts, `namespace = ""`)

	if len(filter.ReportLabel) > 0 {
		for key, value := range filter.ReportLabel {
			argCounter++

			whereParts = append(whereParts, fmt.Sprintf("json_extract(labels, '$.\"%s\"') = $%d", key, argCounter))
			args = append(args, value)
		}
	}

	paginationString := generatePagination(pagination)
	where = strings.Join(whereParts, " AND ")

	list := make([]*api.PolicyReport, 0)

	rows, err := s.db.Query(`SELECT id, source, name, labels, pass, skip, warn, fail, error FROM policy_report as report WHERE `+where+" "+paginationString, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var labels string
		r := &api.PolicyReport{}
		err := rows.Scan(&r.ID, &r.Source, &r.Name, &labels, &r.Pass, &r.Skip, &r.Warn, &r.Fail, &r.Error)
		if err != nil {
			log.Printf("[ERROR] failed to scan PolicyReport: %s", err)
			return list, err
		}

		r.Labels = convertJSONToMap(labels)

		list = append(list, r)
	}

	return list, nil
}

// CountClusterPolicyReports by filter and pagination
func (s *policyReportStore) CountClusterPolicyReports(filter api.Filter) (int, error) {
	whereParts := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int
	var where string

	whereParts = append(whereParts, `namespace = ""`)

	if len(filter.ReportLabel) > 0 {
		for key, value := range filter.ReportLabel {
			argCounter++

			whereParts = append(whereParts, fmt.Sprintf("json_extract(labels, '$.\"%s\"') = $%d", key, argCounter))
			args = append(args, value)
		}
	}

	where = strings.Join(whereParts, " AND ")

	var count int

	row := s.db.QueryRow(`SELECT count(id) FROM policy_report as report WHERE `+where, args...)
	err := row.Scan(&count)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *policyReportStore) FetchClusterPolicies(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT policy FROM policy_report_filter as result`+join+` WHERE result.namespace == ""`+where+` ORDER BY policy ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchClusterRules(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT rule FROM policy_report_result as result`+join+` WHERE resource_namespace == ""`+where+` ORDER BY rule ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespacedPolicies(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT policy FROM policy_report_filter as result`+join+` WHERE result.namespace != ""`+where+` ORDER BY policy ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespacedRules(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT rule FROM policy_report_result as result`+join+` WHERE resource_namespace != ""`+where+` ORDER BY rule ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchCategories(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT category FROM policy_report_filter as result`+join+` WHERE category != ""`+where+` ORDER BY category ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespacedKinds(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "filter_namespaces"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT kind FROM policy_report_filter as result`+join+` WHERE result.namespace != "" AND kind != ""`+where+` ORDER BY kind ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchClusterKinds(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT kind FROM policy_report_filter as result`+join+` WHERE result.namespace == "" AND kind != ""`+where+` ORDER BY kind ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespacedResources(filter api.Filter) ([]*api.Resource, error) {
	list := make([]*api.Resource, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "rules", "namespaces", "kind"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT resource_kind, resource_name FROM policy_report_result as result`+join+` WHERE resource_name != "" AND resource_namespace != ""`+where+` ORDER BY resource_kind, resource_name ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var resource api.Resource
		err := rows.Scan(&resource.Kind, &resource.Name)
		if err != nil {
			return list, err
		}

		list = append(list, &resource)
	}

	return list, nil
}

func (s *policyReportStore) FetchClusterResources(filter api.Filter) ([]*api.Resource, error) {
	list := make([]*api.Resource, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "rules", "kind"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT resource_kind, resource_name FROM policy_report_result as result`+join+` WHERE resource_name != "" AND resource_namespace == ""`+where+` ORDER BY resource_kind, resource_name ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var resource api.Resource
		err := rows.Scan(&resource.Kind, &resource.Name)
		if err != nil {
			return list, err
		}

		list = append(list, &resource)
	}

	return list, nil
}

func (s *policyReportStore) FetchClusterSources() ([]string, error) {
	list := make([]string, 0)
	rows, err := s.db.Query(`SELECT DISTINCT source FROM policy_report WHERE source != "" AND namespace == "" ORDER BY source ASC`)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespacedSources() ([]string, error) {
	list := make([]string, 0)
	rows, err := s.db.Query(`SELECT DISTINCT source FROM policy_report WHERE source != "" AND namespace != "" ORDER BY source ASC`)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespaces(filter api.Filter) ([]string, error) {
	list := make([]string, 0)

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`SELECT DISTINCT result.namespace FROM policy_report_filter as result`+join+` WHERE result.namespace != ""`+where+` ORDER BY result.namespace ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *policyReportStore) FetchNamespacedStatusCounts(filter api.Filter) ([]api.NamespacedStatusCount, error) {
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

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "kinds", "filter_namespaces", "status", "severities"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`
    SELECT SUM(result.count) as counter, result.namespace, status 
    FROM policy_report_filter as result`+join+` WHERE result.namespace != ""`+where+`
    GROUP BY result.namespace, status 
    ORDER BY result.namespace ASC`, args...)
	if err != nil {
		return statusCounts, err
	}
	defer rows.Close()
	for rows.Next() {
		count := api.NamespaceCount{}
		var status string
		err := rows.Scan(&count.Count, &count.Namespace, &status)
		if err != nil {
			return statusCounts, err
		}

		list[status] = append(list[status], count)
	}

	for status, items := range list {
		statusCounts = append(statusCounts, api.NamespacedStatusCount{
			Status: status,
			Items:  items,
		})
	}

	return statusCounts, nil
}

func (s *policyReportStore) FetchRuleStatusCounts(policy, rule string) ([]api.StatusCount, error) {
	list := map[string]api.StatusCount{
		v1alpha2.StatusPass:  {Status: v1alpha2.StatusPass},
		v1alpha2.StatusFail:  {Status: v1alpha2.StatusFail},
		v1alpha2.StatusWarn:  {Status: v1alpha2.StatusWarn},
		v1alpha2.StatusError: {Status: v1alpha2.StatusError},
		v1alpha2.StatusSkip:  {Status: v1alpha2.StatusSkip},
	}

	statusCounts := make([]api.StatusCount, 0, len(list))

	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	argCounter, where, args = appendWhere([]string{policy}, "policy", where, args, argCounter)
	_, where, args = appendWhere([]string{rule}, "rule", where, args, argCounter)

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	rows, err := s.db.Query(`
    SELECT COUNT(id) as counter, status 
    FROM policy_report_result as result`+whereClause+`
    GROUP BY status`, args...)
	if err != nil {
		return statusCounts, err
	}
	defer rows.Close()
	for rows.Next() {
		count := api.StatusCount{}
		err := rows.Scan(&count.Count, &count.Status)
		if err != nil {
			return statusCounts, err
		}

		list[count.Status] = count
	}

	for _, count := range list {
		statusCounts = append(statusCounts, count)
	}

	return statusCounts, nil
}

func (s *policyReportStore) FetchStatusCounts(filter api.Filter) ([]api.StatusCount, error) {
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

	statusCounts := make([]api.StatusCount, 0, len(list))

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "kinds", "status", "severities"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`
    SELECT SUM(result.count) as counter, status 
    FROM policy_report_filter as result`+join+` WHERE result.namespace = ""`+where+`
    GROUP BY status`, args...)
	if err != nil {
		return statusCounts, err
	}
	defer rows.Close()
	for rows.Next() {
		count := api.StatusCount{}
		err := rows.Scan(&count.Count, &count.Status)
		if err != nil {
			return statusCounts, err
		}

		list[count.Status] = count
	}

	for _, count := range list {
		statusCounts = append(statusCounts, count)
	}

	return statusCounts, nil
}

func (s *policyReportStore) FetchNamespacedResults(filter api.Filter, pagination api.Pagination) ([]*api.ListResult, error) {
	list := []*api.ListResult{}

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "rules", "kinds", "resources", "status", "severities", "namespaces"})
	if len(where) > 0 {
		where = " AND " + where
	}
	paginationString := generatePagination(pagination)

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`
    SELECT result.id, resource_namespace, resource_kind, resource_api_version, resource_name, message, policy, rule, severity, properties, status, category, timestamp
    FROM policy_report_result as result`+join+` WHERE resource_namespace != ""`+where+` `+paginationString, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		result := api.ListResult{}
		var props []byte

		err := rows.Scan(&result.ID, &result.Namespace, &result.Kind, &result.APIVersion, &result.Name, &result.Message, &result.Policy, &result.Rule, &result.Severity, &props, &result.Status, &result.Category, &result.Timestamp)
		if err != nil {
			return list, err
		}

		json.Unmarshal(props, &result.Properties)

		list = append(list, &result)
	}

	return list, nil
}

func (s *policyReportStore) CountNamespacedResults(filter api.Filter) (int, error) {
	var count int

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "rules", "kinds", "resources", "status", "severities", "namespaces"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	row := s.db.QueryRow(`SELECT count(result.id) FROM policy_report_result as result`+join+` WHERE resource_namespace != ""`+where, args...)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *policyReportStore) FetchClusterResults(filter api.Filter, pagination api.Pagination) ([]*api.ListResult, error) {
	list := []*api.ListResult{}

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "rules", "kinds", "resources", "status", "severities"})
	if len(where) > 0 {
		where = " AND " + where
	}
	paginationString := generatePagination(pagination)

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	rows, err := s.db.Query(`
    SELECT result.id, resource_namespace, resource_kind, resource_api_version, resource_name, message, policy, rule, severity, properties, status, category, timestamp
    FROM policy_report_result as result`+join+` WHERE resource_namespace =""`+where+` `+paginationString, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		result := api.ListResult{}
		var props []byte

		err := rows.Scan(&result.ID, &result.Namespace, &result.Kind, &result.APIVersion, &result.Name, &result.Message, &result.Policy, &result.Rule, &result.Severity, &props, &result.Status, &result.Category, &result.Timestamp)
		if err != nil {
			return list, err
		}

		json.Unmarshal(props, &result.Properties)

		list = append(list, &result)
	}

	return list, nil
}

func (s *policyReportStore) CountClusterResults(filter api.Filter) (int, error) {
	var count int

	where, args := generateFilterWhere(filter, []string{"sources", "categories", "policies", "rules", "kinds", "resources", "status", "severities"})
	if len(where) > 0 {
		where = " AND " + where
	}

	join := ""
	if len(filter.ReportLabel) > 0 {
		join = " JOIN policy_report as report ON result.policy_report_id = report.id"
	}

	row := s.db.QueryRow(`SELECT count(result.id) FROM policy_report_result as result`+join+` WHERE resource_namespace =""`+where, args...)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *policyReportStore) FetchNamespacedReportLabels(filter api.Filter) (map[string][]string, error) {
	list := make(map[string][]string)

	where, args := generateFilterWhere(filter, []string{"report_sources", "report_namespaces"})
	if len(where) > 0 {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT report.labels from policy_report as report WHERE report.namespace != ""`+where+` ORDER BY report.namespace ASC`, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		for key, value := range convertJSONToMap(item) {
			if value == "" {
				continue
			}

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

func (s *policyReportStore) FetchClusterReportLabels(filter api.Filter) (map[string][]string, error) {
	list := make(map[string][]string)

	where, args := generateFilterWhere(filter, []string{"report_sources"})
	if len(where) > 0 {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT report.labels from policy_report as report WHERE report.namespace = ""`+where, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return list, err
		}

		for key, value := range convertJSONToMap(item) {
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

func (s *policyReportStore) persistResults(report v1alpha2.ReportInterface) error {
	var vals []interface{}
	var sqlStr string

	bulks := chunkSlice(report.GetResults(), 50)

	for _, list := range bulks {
		sqlStr = resultInsertBaseSQL
		vals = make([]interface{}, 0, len(list)*18)

		for _, result := range list {
			sqlStr += "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),"

			var props string
			b, err := json.Marshal(result.Properties)
			if err == nil {
				props = string(b)
			}

			res := result.GetResource()
			if res == nil && report.GetScope() != nil {
				res = report.GetScope()
			} else if res == nil {
				res = &corev1.ObjectReference{}
			}

			vals = append(
				vals,
				report.GetID(),
				result.GetID(),
				result.Policy,
				result.Rule,
				result.Message,
				result.Scored,
				string(result.Result),
				result.Severity,
				result.Category,
				result.Source,
				res.APIVersion,
				res.Kind,
				res.Name,
				report.GetNamespace(),
				res.UID,
				props,
				result.Timestamp.Seconds,
			)
		}

		sqlStr = sqlStr[0 : len(sqlStr)-1]
		rstmt, err := s.db.Prepare(sqlStr)
		if err != nil {
			return err
		}
		defer rstmt.Close()

		_, err = rstmt.Exec(vals...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *policyReportStore) persistFilterValues(report v1alpha2.ReportInterface) error {
	var vals []interface{}
	var sqlStr string

	values := api.ExtractFilterValues(report)

	bulks := chunkSlice(values, 50)

	for _, list := range bulks {
		sqlStr = filterInsertBaseSQL
		vals = make([]interface{}, 0, len(list)*18)

		for _, result := range list {
			sqlStr += "(?,?,?,?,?,?,?,?,?),"

			vals = append(
				vals,
				result.ReportID,
				result.Namespace,
				result.Policy,
				result.Result,
				result.Severity,
				result.Category,
				result.Source,
				result.Kind,
				result.Count,
			)
		}

		sqlStr = sqlStr[0 : len(sqlStr)-1]
		rstmt, err := s.db.Prepare(sqlStr)
		if err != nil {
			return err
		}
		defer rstmt.Close()

		_, err = rstmt.Exec(vals...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *policyReportStore) fetchResults(reportID string) ([]v1alpha2.PolicyReportResult, error) {
	results := make([]v1alpha2.PolicyReportResult, 0)

	rows, err := s.db.Query(`
    SELECT 
      id,
      policy,
      rule,
      message,
      scored,
      status,
      severity, 
      category, 
      source, 
      resource_api_version,
      resource_kind,
      resource_name,
      resource_namespace,
      resource_uid, 
      properties,
      timestamp
    FROM policy_report_result
    WHERE policy_report_id=$1
  `, reportID)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	var props []byte
	var timestamp int64

	for rows.Next() {
		result := v1alpha2.PolicyReportResult{
			Resources: make([]corev1.ObjectReference, 0, 1),
		}

		resource := corev1.ObjectReference{}

		err = rows.Scan(
			&result.ID,
			&result.Policy,
			&result.Rule,
			&result.Message,
			&result.Scored,
			&result.Result,
			&result.Severity,
			&result.Category,
			&result.Source,
			&resource.APIVersion,
			&resource.Kind,
			&resource.Name,
			&resource.Namespace,
			&resource.UID,
			&props,
			&timestamp,
		)
		if err != nil {
			return results, err
		}

		err = json.Unmarshal(props, &result.Properties)
		if err != nil {
			result.Properties = make(map[string]string)
		}

		result.Timestamp = v1.Timestamp{
			Seconds: timestamp,
		}

		result.Resources = append(result.Resources, resource)

		results = append(results, result)
	}

	return results, nil
}

func appendWhere(options []string, field string, where []string, args []interface{}, argCounter int) (int, []string, []interface{}) {
	length := len(options)

	if length == 0 {
		return argCounter, where, args
	}

	if length == 1 {
		option := options[0]
		argCounter++

		args = append(args, strings.ToLower(option))

		where = append(where, fmt.Sprintf("LOWER(%s)=$%d", field, argCounter))

		return argCounter + length, where, args
	}

	arguments := make([]string, 0, length)

	for _, option := range options {
		argCounter++

		arguments = append(arguments, fmt.Sprintf("$%d", argCounter))
		args = append(args, strings.ToLower(option))
	}

	where = append(where, "LOWER("+field+") IN ("+strings.Join(arguments, ",")+")")

	return argCounter + length, where, args
}

func generateFilterWhere(filter api.Filter, active []string) (string, []interface{}) {
	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	if contains("report_namespaces", active) {
		argCounter, where, args = appendWhere(filter.Namespaces, "report.namespace", where, args, argCounter)
	}
	if contains("report_sources", active) {
		argCounter, where, args = appendWhere(filter.Sources, "report.source", where, args, argCounter)
	}
	if contains("namespaces", active) {
		argCounter, where, args = appendWhere(filter.Namespaces, "result.resource_namespace", where, args, argCounter)
	}
	if contains("filter_namespaces", active) {
		argCounter, where, args = appendWhere(filter.Namespaces, "result.namespace", where, args, argCounter)
	}
	if contains("policies", active) {
		argCounter, where, args = appendWhere(filter.Policies, "result.policy", where, args, argCounter)
	}
	if contains("rules", active) {
		argCounter, where, args = appendWhere(filter.Rules, "result.rule", where, args, argCounter)
	}
	if contains("kinds", active) {
		argCounter, where, args = appendWhere(filter.Kinds, "result.resource_kind", where, args, argCounter)
	}
	if contains("resources", active) {
		argCounter, where, args = appendWhere(filter.Resources, "result.resource_name", where, args, argCounter)
	}
	if contains("sources", active) {
		argCounter, where, args = appendWhere(filter.Sources, "result.source", where, args, argCounter)
	}
	if contains("categories", active) {
		argCounter, where, args = appendWhere(filter.Categories, "result.category", where, args, argCounter)
	}
	if contains("severities", active) {
		argCounter, where, args = appendWhere(filter.Severities, "result.severity", where, args, argCounter)
	}
	if contains("status", active) {
		argCounter, where, args = appendWhere(filter.Status, "result.status", where, args, argCounter)
	}
	if filter.Search != "" {
		likeIndex := argCounter + 1
		equalIndex := argCounter + 2
		argCounter += 2

		where = append(where, fmt.Sprintf(
			`(resource_namespace LIKE $%d OR resource_name LIKE $%d OR policy LIKE $%d OR rule LIKE $%d OR severity = $%d OR status = $%d OR LOWER(resource_kind) = LOWER($%d))`,
			likeIndex,
			likeIndex,
			likeIndex,
			likeIndex,
			equalIndex,
			equalIndex,
			equalIndex,
		))
		args = append(args, filter.Search+"%", filter.Search)
	}
	if len(filter.ReportLabel) > 0 {
		for key, value := range filter.ReportLabel {
			argCounter++

			where = append(where, fmt.Sprintf("json_extract(report.labels, '$.\"%s\"') = $%d", key, argCounter))
			args = append(args, value)
		}
	}

	return strings.Join(where, " AND "), args
}

func generatePagination(pagination api.Pagination) string {
	if pagination.Page == 0 || pagination.Offset == 0 {
		return fmt.Sprintf(
			"ORDER BY %s %s",
			strings.Join(pagination.SortBy, ","),
			pagination.Direction,
		)
	}

	return fmt.Sprintf(
		"ORDER BY %s %s LIMIT %d OFFSET %d",
		strings.Join(pagination.SortBy, ","),
		pagination.Direction,
		pagination.Offset,
		(pagination.Page-1)*pagination.Offset,
	)
}

func contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
}

func convertSliveToJSON(m []string) string {
	str, err := json.Marshal(m)
	if err != nil {
		return "[]"
	}

	return string(str)
}

func convertMapToJSON(m map[string]string) string {
	str, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}

	return string(str)
}

func convertJSONToMap(s string) map[string]string {
	m := make(map[string]string)
	_ = json.Unmarshal([]byte(s), &m)

	return m
}

func convertJSONToSlice(s string) []string {
	m := make([]string, 0)
	_ = json.Unmarshal([]byte(s), &m)

	return m
}

func appendUnique(list, values []string) []string {
	for _, v := range values {
		if contains(v, list) {
			continue
		}

		list = append(list, v)
	}

	return list
}

// NewPolicyReportStore construct a PolicyReportStore
func NewPolicyReportStore(db *sql.DB) (PolicyReportStore, error) {
	var err error

	s := &policyReportStore{db}
	if db != nil {
		err = s.CreateSchemas()
	}

	return s, err
}

func NewDatabase(dbFile string) (*sql.DB, error) {
	os.Remove(dbFile)
	file, err := os.Create(dbFile)
	if err != nil {
		return nil, err
	}
	file.Close()

	return sql.Open("sqlite3", dbFile+"?cache=shared")
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
