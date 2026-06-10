package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
)

var ErrNoConfig = errors.New("no configuration for the provider found")

// DatabaseFactory manages database connection and creation
type DatabaseFactory struct {
	secretClient secrets.Client
}

func (f *DatabaseFactory) NewPostgres(config Database) *bun.DB {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(&config, config.SecretRef, config.MountedSecret)
	}

	if config.Host == "" && config.DSN == "" {
		return nil
	}

	dsn := config.DSN
	if config.DSN == "" {
		sslMode := "disable"
		if config.EnableSSL {
			sslMode = "verify-full"
		}

		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", config.Username, config.Password, config.Host, config.Database, sslMode)
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithTimeout(time.Duration(WithDefault(config.Timeout, 5))*time.Second),
	))

	sqldb.SetMaxOpenConns(config.MaxOpenConns)
	sqldb.SetMaxIdleConns(config.MaxIdleConns)
	sqldb.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)
	sqldb.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Minute)

	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook())
	if config.Metrics {
		db.AddQueryHook(database.NewQueryHook(database.WithDBName(config.Database)))
	}

	return db
}

func (f *DatabaseFactory) NewMySQL(config Database) *bun.DB {
	if (config.SecretRef != "" && f.secretClient != nil) || config.MountedSecret != "" {
		f.mapSecretValues(&config, config.SecretRef, config.MountedSecret)
	}

	if config.Host == "" && config.DSN == "" {
		return nil
	}

	dsn := config.DSN
	if config.DSN == "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=%v", config.Username, config.Password, config.Host, config.Database, config.EnableSSL)
	}
	if config.Timeout > 0 {
		dsn = fmt.Sprintf("%s&timeout=%ds", dsn, config.Timeout)
	}

	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		zap.L().Error("failed to create mysql connection", zap.Error(err))
		return nil
	}

	sqldb.SetMaxOpenConns(config.MaxOpenConns)
	sqldb.SetMaxIdleConns(config.MaxIdleConns)
	sqldb.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)
	sqldb.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Minute)

	db := bun.NewDB(sqldb, mysqldialect.New())
	if config.Metrics {
		db.AddQueryHook(database.NewQueryHook(database.WithDBName(config.Database)))
	}

	return db
}

func (f *DatabaseFactory) NewSQLite(file string, metrics bool) *bun.DB {
	sqldb, err := database.NewSQLiteDB(file)
	if err != nil {
		zap.L().Error("failed to create sqlite connection", zap.Error(err))
		return nil
	}

	sqldb.AddQueryHook(bundebug.NewQueryHook())
	if metrics {
		sqldb.AddQueryHook(database.NewQueryHook(database.WithDBName("sqlite")))
	}

	return sqldb
}

func (f *DatabaseFactory) mapSecretValues(config any, ref, mountedSecret string) {
	values := secrets.Values{}

	if ref != "" {
		secretValues, err := f.secretClient.Get(context.Background(), ref)
		values = secretValues
		if err != nil {
			zap.L().Warn("failed to get secret reference", zap.Error(err))
			return
		}
	}

	if mountedSecret != "" {
		file, err := os.ReadFile(mountedSecret)
		if err != nil {
			zap.L().Warn("failed to get mounted secret", zap.Error(err))
			return
		}
		err = json.Unmarshal(file, &values)
		if err != nil {
			zap.L().Warn("failed to unmarshal mounted secret", zap.Error(err))
			return
		}
	}

	if c, ok := config.(*Database); ok {
		if values.Host != "" {
			c.Host = values.Host
		}
		if values.Username != "" {
			c.Username = values.Username
		}
		if values.Password != "" {
			c.Password = values.Password
		}
		if values.Database != "" {
			c.Database = values.Database
		}
		if values.DSN != "" {
			c.DSN = values.DSN
		}
	}
}

func NewDatabaseFactory(client secrets.Client) *DatabaseFactory {
	return &DatabaseFactory{
		secretClient: client,
	}
}

func WithDefault[T comparable](value, fallback T) T {
	var init T
	if value == init {
		return fallback
	}

	return value
}
