package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	api "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/report"
	_ "github.com/mattn/go-sqlite3"
)

const (
	reportSQL = `CREATE TABLE policy_report (
		"id" TEXT NOT NULL PRIMARY KEY,		
		"type" TEXT,
		"namespace" TEXT,
		"name" TEXT NOT NULL,
		"skip" INTEGER DEFAULT 0,
		"pass" INTEGER DEFAULT 0,
		"warn" INTEGER DEFAULT 0,
		"fail" INTEGER DEFAULT 0,
		"error" INTEGER DEFAULT 0,
		"created" INTEGER
	);`

	resultSQL = `CREATE TABLE policy_report_result (
		"policy_report_id" TEXT NOT NULL,
		"id" TEXT NOT NULL PRIMARY KEY,
		"policy" TEXT,
		"rule" TEXT,
		"message" TEXT,
		"scored" INTEGER,
		"priority" TEXT,
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
		FOREIGN KEY (policy_report_id) REFERENCES policy_report(id) ON DELETE CASCADE
	);`
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

	_, err = s.db.Exec(resultSQL)

	return err
}

// Get an PolicyReport by Type and ID
func (s *policyReportStore) Get(id string) (*report.PolicyReport, bool) {
	var created int64
	r := &report.PolicyReport{Summary: &report.Summary{}}

	row := s.db.QueryRow("SELECT namespace, name, pass, skip, warn, fail, error, created FROM policy_report WHERE id=$1", id)
	err := row.Scan(&r.Namespace, &r.Name, &r.Summary.Pass, &r.Summary.Skip, &r.Summary.Warn, &r.Summary.Fail, &r.Summary.Error, &created)
	if err == sql.ErrNoRows {
		return r, false
	} else if err != nil {
		log.Printf("[ERROR] Failed to select PolicyReport: %s", err)
		return r, false
	}

	r.CreationTimestamp = time.Unix(created, 0)

	results, err := s.fetchResults(id)
	if err != nil {
		log.Printf("Failed to fetch Reports: %s\n", err)
		return r, false
	}

	r.Results = results

	return r, true
}

// Add a PolicyReport to the Store
func (s *policyReportStore) Add(r *report.PolicyReport) error {
	stmt, err := s.db.Prepare("INSERT INTO policy_report(id, type, namespace, name, pass, skip, warn, fail, error, created) values(?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(r.GetIdentifier(), r.GetType(), r.Namespace, r.Name, r.Summary.Pass, r.Summary.Skip, r.Summary.Warn, r.Summary.Fail, r.Summary.Error, r.CreationTimestamp.Unix())
	if err != nil {
		return err
	}

	return s.persistResults(r)
}

func (s *policyReportStore) Update(r *report.PolicyReport) error {
	stmt, err := s.db.Prepare("UPDATE policy_report SET pass=?, skip=?, warn=?, fail=?, error=?, created=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(r.Summary.Pass, r.Summary.Skip, r.Summary.Warn, r.Summary.Fail, r.Summary.Error, r.CreationTimestamp.Unix(), r.GetIdentifier())
	if err != nil {
		return err
	}

	dstmt, err := s.db.Prepare("DELETE FROM policy_report_result WHERE policy_report_id=?")
	if err != nil {
		return err
	}
	defer dstmt.Close()

	_, err = dstmt.Exec(r.GetIdentifier())
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

	dstmt, err := s.db.Prepare("DELETE FROM policy_report_result")
	if err != nil {
		return err
	}
	defer dstmt.Close()

	_, err = dstmt.Exec()
	return err
}

func (s *policyReportStore) FetchClusterPolicies(source string) ([]string, error) {
	list := make([]string, 0)

	where, args := appendSourceWhere(source)
	if where != "" {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT policy FROM policy_report_result WHERE resource_namespace == ""`+where+` ORDER BY policy ASC`, args...)
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

func (s *policyReportStore) FetchNamespacedPolicies(source string) ([]string, error) {
	list := make([]string, 0)

	where, args := appendSourceWhere(source)
	if where != "" {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT policy FROM policy_report_result WHERE resource_namespace != ""`+where+` ORDER BY policy ASC`, args...)
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

func (s *policyReportStore) FetchCategories(source string) ([]string, error) {
	list := make([]string, 0)

	where, args := appendSourceWhere(source)
	if where != "" {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT category FROM policy_report_result WHERE category != ""`+where+` ORDER BY category ASC`, args...)
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

func (s *policyReportStore) FetchNamespacedKinds(source string) ([]string, error) {
	list := make([]string, 0)

	where, args := appendSourceWhere(source)
	if where != "" {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT resource_kind FROM policy_report_result WHERE resource_kind != "" AND resource_namespace != ""`+where+` ORDER BY resource_kind ASC`, args...)
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

func (s *policyReportStore) FetchClusterKinds(source string) ([]string, error) {
	list := make([]string, 0)

	where, args := appendSourceWhere(source)
	if where != "" {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT resource_kind FROM policy_report_result WHERE resource_kind != "" AND resource_namespace == ""`+where+` ORDER BY resource_kind ASC`, args...)
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

func (s *policyReportStore) FetchClusterSources() ([]string, error) {
	list := make([]string, 0)
	rows, err := s.db.Query(`SELECT DISTINCT source FROM policy_report_result WHERE source != "" AND resource_namespace == "" ORDER BY source ASC`)
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
	rows, err := s.db.Query(`SELECT DISTINCT source FROM policy_report_result WHERE source != "" AND resource_namespace != "" ORDER BY source ASC`)
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

func (s *policyReportStore) FetchNamespaces(source string) ([]string, error) {
	list := make([]string, 0)

	where, args := appendSourceWhere(source)
	if where != "" {
		where = " AND " + where
	}

	rows, err := s.db.Query(`SELECT DISTINCT resource_namespace FROM policy_report_result WHERE resource_namespace != ""`+where+` ORDER BY resource_namespace ASC`, args...)
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
			report.Pass:  make([]api.NamespaceCount, 0),
			report.Fail:  make([]api.NamespaceCount, 0),
			report.Warn:  make([]api.NamespaceCount, 0),
			report.Error: make([]api.NamespaceCount, 0),
			report.Skip:  make([]api.NamespaceCount, 0),
		}
	} else {
		list = map[string][]api.NamespaceCount{}

		for _, status := range filter.Status {
			list[status] = make([]api.NamespaceCount, 0)
		}
	}

	statusCounts := make([]api.NamespacedStatusCount, 0, 5)

	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	argCounter, where, args = appendWhere(filter.Policies, "policy", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Kinds, "resource_kind", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Sources, "source", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Categories, "category", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Severities, "severity", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Status, "status", where, args, argCounter)
	_, where, args = appendWhere(filter.Namespaces, "resource_namespace", where, args, argCounter)

	whereClause := ""
	if len(where) > 0 {
		whereClause = " AND " + strings.Join(where, " AND ")
	}

	rows, err := s.db.Query(`
		SELECT COUNT(id) as counter, resource_namespace, status 
		FROM policy_report_result WHERE resource_namespace != ""`+whereClause+`
		GROUP BY resource_namespace, status 
		ORDER BY resource_namespace ASC`, args...)

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
		report.Pass:  {Status: report.Pass},
		report.Fail:  {Status: report.Fail},
		report.Warn:  {Status: report.Warn},
		report.Error: {Status: report.Error},
		report.Skip:  {Status: report.Skip},
	}

	statusCounts := make([]api.StatusCount, 0, len(list))

	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	argCounter, where, args = appendWhere([]string{policy}, "policy", where, args, argCounter)
	argCounter, where, args = appendWhere([]string{rule}, "rule", where, args, argCounter)

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	rows, err := s.db.Query(`
		SELECT COUNT(id) as counter, status 
		FROM policy_report_result`+whereClause+`
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
			report.Pass:  {Status: report.Pass},
			report.Fail:  {Status: report.Fail},
			report.Warn:  {Status: report.Warn},
			report.Error: {Status: report.Error},
			report.Skip:  {Status: report.Skip},
		}
	} else {
		list = map[string]api.StatusCount{}

		for _, status := range filter.Status {
			list[status] = api.StatusCount{Status: status}
		}
	}

	statusCounts := make([]api.StatusCount, 0, len(list))

	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	argCounter, where, args = appendWhere(filter.Policies, "policy", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Kinds, "resource_kind", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Sources, "source", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Categories, "category", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Severities, "severity", where, args, argCounter)
	_, where, args = appendWhere(filter.Status, "status", where, args, argCounter)

	whereClause := ""
	if len(where) > 0 {
		whereClause = " AND " + strings.Join(where, " AND ")
	}

	rows, err := s.db.Query(`
		SELECT COUNT(id) as counter, status 
		FROM policy_report_result WHERE resource_namespace = ""`+whereClause+`
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

func (s *policyReportStore) FetchNamespacedResults(filter api.Filter) ([]*api.ListResult, error) {
	list := []*api.ListResult{}

	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	argCounter, where, args = appendWhere(filter.Policies, "policy", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Kinds, "resource_kind", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Sources, "source", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Categories, "category", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Severities, "severity", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Status, "status", where, args, argCounter)
	_, where, args = appendWhere(filter.Namespaces, "resource_namespace", where, args, argCounter)

	whereClause := ""
	if len(where) > 0 {
		whereClause = " AND " + strings.Join(where, " AND ")
	}

	rows, err := s.db.Query(`
		SELECT id, resource_namespace, resource_kind, resource_name, message, policy, rule, severity, properties, status 
		FROM policy_report_result WHERE resource_namespace != ""`+whereClause+`
		ORDER BY resource_namespace, resource_name, resource_uid ASC`, args...)

	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		result := api.ListResult{}
		var props []byte

		err := rows.Scan(&result.ID, &result.Namespace, &result.Kind, &result.Name, &result.Message, &result.Policy, &result.Rule, &result.Severity, &props, &result.Status)
		if err != nil {
			return list, err
		}

		json.Unmarshal(props, &result.Properties)

		list = append(list, &result)
	}

	return list, nil
}

func (s *policyReportStore) FetchClusterResults(filter api.Filter) ([]*api.ListResult, error) {
	list := []*api.ListResult{}

	where := make([]string, 0)
	args := make([]interface{}, 0)

	var argCounter int

	argCounter, where, args = appendWhere(filter.Policies, "policy", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Kinds, "resource_kind", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Sources, "source", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Categories, "category", where, args, argCounter)
	argCounter, where, args = appendWhere(filter.Severities, "severity", where, args, argCounter)
	_, where, args = appendWhere(filter.Status, "status", where, args, argCounter)

	whereClause := ""
	if len(where) > 0 {
		whereClause = " AND " + strings.Join(where, " AND ")
	}

	rows, err := s.db.Query(`
		SELECT id, resource_namespace, resource_kind, resource_name, message, policy, rule, severity, properties, status 
		FROM policy_report_result WHERE resource_namespace =""`+whereClause+`
		ORDER BY resource_namespace, resource_name, resource_uid ASC`, args...)

	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		result := api.ListResult{}
		var props []byte

		err := rows.Scan(&result.ID, &result.Namespace, &result.Kind, &result.Name, &result.Message, &result.Policy, &result.Rule, &result.Severity, &props, &result.Status)
		if err != nil {
			return list, err
		}

		json.Unmarshal(props, &result.Properties)

		list = append(list, &result)
	}

	return list, nil
}

func (s *policyReportStore) persistResults(report *report.PolicyReport) error {
	for _, result := range report.Results {
		rstmt, err := s.db.Prepare("INSERT INTO policy_report_result(policy_report_id, id, policy, rule, message, scored, priority, status, severity, category, source, resource_api_version, resource_kind, resource_name, resource_namespace, resource_uid, properties, timestamp) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			return err
		}
		defer rstmt.Close()

		var props string

		b, err := json.Marshal(result.Properties)
		if err == nil {
			props = string(b)
		}

		_, err = rstmt.Exec(
			report.GetIdentifier(),
			result.GetIdentifier(),
			result.Policy,
			result.Rule,
			result.Message,
			result.Scored,
			result.Priority,
			result.Status,
			result.Severity,
			result.Category,
			result.Source,
			result.Resource.APIVersion,
			result.Resource.Kind,
			result.Resource.Name,
			result.Resource.Namespace,
			result.Resource.UID,
			props,
			result.Timestamp.Unix(),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *policyReportStore) fetchResults(reportID string) (map[string]*report.Result, error) {
	results := make(map[string]*report.Result)

	rows, err := s.db.Query(`
		SELECT 
			id,
			policy,
			rule,
			message,
			scored,
			priority,
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
		result := &report.Result{
			Resource: &report.Resource{},
		}

		err = rows.Scan(
			&result.ID,
			&result.Policy,
			&result.Rule,
			&result.Message,
			&result.Scored,
			&result.Priority,
			&result.Status,
			&result.Severity,
			&result.Category,
			&result.Source,
			&result.Resource.APIVersion,
			&result.Resource.Kind,
			&result.Resource.Name,
			&result.Resource.Namespace,
			&result.Resource.UID,
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

		result.Timestamp = time.Unix(timestamp, 0)

		results[result.GetIdentifier()] = result
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

func appendSourceWhere(source string) (string, []interface{}) {
	if source == "" {
		return "", make([]interface{}, 0)
	}

	return "LOWER(source)=$1", []interface{}{strings.ToLower(source)}
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

	return sql.Open("sqlite3", dbFile)
}
