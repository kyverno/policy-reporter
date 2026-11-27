package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

const (
	postgresTestDSNEnv  = "POLICY_REPORTER_TEST_POSTGRES_DSN"
	postgresTestSchema  = "policy_reporter_cleanup_"
	postgresTestVersion = "test"
)

var cacheTables = []string{"policy_report", "policy_report_resource", "policy_report_result", "policy_report_filter"}

// newPostgresTestStore connects directly to the PostgreSQL instance from
// POLICY_REPORTER_TEST_POSTGRES_DSN and returns a Store that only sees a fresh
// schema. Use a disposable database and a direct connection, not a
// transaction-mode pooler. pgdriver applies the search_path query parameter
// with SET on every new connection, and the pool is capped at one connection,
// so the schema guard and the cleanup share a session.
func newPostgresTestStore(t *testing.T) (*Store, *bun.DB, string) {
	t.Helper()

	dsn := os.Getenv(postgresTestDSNEnv)
	if dsn == "" {
		t.Skipf("set %s to a PostgreSQL URL DSN to run the PostgreSQL cleanup tests", postgresTestDSNEnv)
	}

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse %s: %v", postgresTestDSNEnv, err)
	}

	schema := postgresTestSchema + strings.ToLower(rand.Text())
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(u.String())))
	sqldb.SetMaxOpenConns(1)
	sqldb.SetMaxIdleConns(1)
	db := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create schema %s: %v", schema, err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE"); err != nil {
			t.Errorf("drop schema %s: %v", schema, err)
		}
	})

	store, err := NewStore(db, postgresTestVersion)
	if err != nil {
		t.Fatal(err)
	}

	return store, db, schema
}

// requireTestSchema fails closed unless the single pooled connection resolves to
// the generated schema only and has no temporary schema, which would be searched
// first. The check is point in time: call it right before a destructive
// statement so connection reuse cannot point the cleanup at another schema.
func requireTestSchema(t *testing.T, ctx context.Context, db *bun.DB, schema string) {
	t.Helper()

	var current, resolved string
	var tempSchema int64
	if err := db.QueryRowContext(ctx, "SELECT current_schema(), array_to_string(current_schemas(false), ','), pg_my_temp_schema()::bigint").Scan(&current, &resolved, &tempSchema); err != nil {
		t.Fatalf("resolve schema: %v", err)
	}
	if current != schema || resolved != schema {
		t.Fatalf("connection resolves to current_schema %q and search path %q, want %q only", current, resolved, schema)
	}
	if tempSchema != 0 {
		t.Fatalf("session has temporary schema %d, which is searched before %q", tempSchema, schema)
	}
}

func seedPostgresCache(t *testing.T, ctx context.Context, store *Store, db *bun.DB) map[string]int {
	t.Helper()

	rep := result.NewReconditioner(nil).Prepare(fixtures.DefaultPolicyReport)
	if err := store.Add(ctx, rep); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	counts := cacheRowCounts(t, ctx, db)
	for _, table := range cacheTables {
		if counts[table] == 0 {
			t.Fatalf("table %s is empty after seeding", table)
		}
	}

	return counts
}

func cacheRowCounts(t *testing.T, ctx context.Context, db *bun.DB) map[string]int {
	t.Helper()

	counts := make(map[string]int, len(cacheTables))
	for _, table := range cacheTables {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		counts[table] = count
	}

	return counts
}

func cacheVersion(t *testing.T, ctx context.Context, db *bun.DB) string {
	t.Helper()

	var version string
	if err := db.QueryRowContext(ctx, "SELECT version FROM policy_report_config WHERE id = 1").Scan(&version); err != nil {
		t.Fatalf("read config version: %v", err)
	}

	return version
}

func TestPostgresCleanUpTruncatesCacheTables(t *testing.T) {
	ctx := context.Background()
	store, db, schema := newPostgresTestStore(t)

	requireTestSchema(t, ctx, db, schema)
	if err := store.PrepareDatabase(ctx); err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	seedPostgresCache(t, ctx, store, db)

	if _, err := db.ExecContext(ctx, "CREATE INDEX cache_index_survives ON policy_report_result (policy)"); err != nil {
		t.Fatalf("create index: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TABLE unrelated_table (id integer PRIMARY KEY)"); err != nil {
		t.Fatalf("create unrelated table: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO unrelated_table VALUES (1)"); err != nil {
		t.Fatalf("seed unrelated table: %v", err)
	}

	// The persisted version matches, so this call only runs the cleanup.
	requireTestSchema(t, ctx, db, schema)
	if err := store.PrepareDatabase(ctx); err != nil {
		t.Fatalf("prepare database with matching version: %v", err)
	}

	after := cacheRowCounts(t, ctx, db)
	for _, table := range cacheTables {
		if after[table] != 0 {
			t.Errorf("%s rows after cleanup = %d, want 0", table, after[table])
		}
	}
	if version := cacheVersion(t, ctx, db); version != postgresTestVersion {
		t.Errorf("config version after cleanup = %q, want %q", version, postgresTestVersion)
	}

	var indexes, unrelated int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = current_schema() AND indexname = 'cache_index_survives'").Scan(&indexes); err != nil {
		t.Fatalf("read index: %v", err)
	}
	if indexes != 1 {
		t.Errorf("cache_index_survives exists %d times after cleanup, want 1", indexes)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM unrelated_table").Scan(&unrelated); err != nil {
		t.Fatalf("count unrelated table: %v", err)
	}
	if unrelated != 1 {
		t.Errorf("unrelated_table rows after cleanup = %d, want 1", unrelated)
	}
}

func TestPostgresCleanUpFailsAtomicallyWithInboundForeignKey(t *testing.T) {
	ctx := context.Background()
	store, db, schema := newPostgresTestStore(t)

	requireTestSchema(t, ctx, db, schema)
	if err := store.PrepareDatabase(ctx); err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	before := seedPostgresCache(t, ctx, store, db)

	if _, err := db.ExecContext(ctx, "CREATE TABLE external_report_reference (report_id text REFERENCES policy_report (id))"); err != nil {
		t.Fatalf("create inbound foreign key: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO external_report_reference SELECT id FROM policy_report LIMIT 1"); err != nil {
		t.Fatalf("seed inbound foreign key: %v", err)
	}

	requireTestSchema(t, ctx, db, schema)
	err := store.PrepareDatabase(ctx)
	if err == nil {
		t.Fatal("expected cleanup to fail on an inbound foreign key")
	}
	if !strings.Contains(err.Error(), "cannot truncate a table referenced in a foreign key constraint") {
		t.Fatalf("cleanup error = %v, want a PostgreSQL truncate foreign key constraint error", err)
	}

	after := cacheRowCounts(t, ctx, db)
	for _, table := range cacheTables {
		if after[table] != before[table] {
			t.Errorf("%s rows after failed cleanup = %d, want %d", table, after[table], before[table])
		}
	}
	if version := cacheVersion(t, ctx, db); version != postgresTestVersion {
		t.Errorf("config version after failed cleanup = %q, want %q", version, postgresTestVersion)
	}
}
