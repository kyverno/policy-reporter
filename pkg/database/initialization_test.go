package database

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattn/go-sqlite3"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func TestPersistenceErrors(t *testing.T) {
	ctx := context.Background()
	db, err := NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewStore(db, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PrepareDatabase(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_report BEFORE INSERT ON policy_report BEGIN SELECT RAISE(FAIL, 'report insert rejected'); END`); err != nil {
		t.Fatal(err)
	}
	rep := result.NewReconditioner(nil).Prepare(fixtures.DefaultPolicyReport)
	for name, write := range map[string]func() error{
		"Add":    func() error { return store.Add(ctx, rep) },
		"Update": func() error { return store.Update(ctx, rep) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := write(); err == nil || !strings.Contains(err.Error(), "report insert rejected") {
				t.Fatalf("expected original insert failure, got %v", err)
			}
		})
	}
}

func TestPrepareDatabasePreservesEarlySchemaFailure(t *testing.T) {
	ctx := context.Background()
	db, err := NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewStore(db, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PrepareDatabase(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE config_reference (config_id INTEGER REFERENCES policy_report_config(id)); INSERT INTO config_reference VALUES (1)`); err != nil {
		t.Fatal(err)
	}
	if err := store.PrepareDatabase(ctx); err == nil {
		t.Fatal("expected failure dropping referenced config table")
	}
}

func TestCreateSchemasPreservesEarlyFailure(t *testing.T) {
	ctx := context.Background()
	db, err := NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, err := db.DB.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = conn.Raw(func(driver any) error {
		driver.(*sqlite3.SQLiteConn).RegisterAuthorizer(func(action int, table, _, _ string) int {
			if action == sqlite3.SQLITE_CREATE_TABLE && table == "policy_report_config" {
				return sqlite3.SQLITE_DENY
			}
			return sqlite3.SQLITE_OK
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(db, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSchemas(ctx); err == nil {
		t.Fatal("expected the first schema creation failure")
	}
}
