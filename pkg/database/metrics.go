package database

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func RegisterDBStats(name string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "DB Status",
	}, []string{"database", "system"})
}

func RegisterHistogram(name string) *prometheus.HistogramVec {
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: name,
		Help: "Timing of processed queries in milliseconds",
	}, []string{"database", "system", "operation", "table"})
}

type Option func(h *QueryHook)

func WithDBName(name string) Option {
	return func(h *QueryHook) {
		h.dbName = name
	}
}

type QueryHook struct {
	dbName             string
	formatQueries      bool
	connections        *prometheus.GaugeVec
	maxOpenConnections *prometheus.GaugeVec
	maxIdleTimeClosed  *prometheus.GaugeVec
	maxIdleClosed      *prometheus.GaugeVec
	idleConnections    *prometheus.GaugeVec
	maxLifetimeClosed  *prometheus.GaugeVec
	inUse              *prometheus.GaugeVec
	waitCount          *prometheus.GaugeVec
	waitDuration       *prometheus.GaugeVec
	queryHistogram     *prometheus.HistogramVec
	spanNameQueryGen   func(*bun.QueryEvent) string
}

var _ bun.QueryHook = (*QueryHook)(nil)

func NewQueryHook(opts ...Option) *QueryHook {
	h := new(QueryHook)

	for _, opt := range opts {
		opt(h)
	}

	h.connections = RegisterDBStats("database_connections")
	h.maxOpenConnections = RegisterDBStats("database_max_open_connections")
	h.maxIdleTimeClosed = RegisterDBStats("database_max_idle_time_closed")
	h.maxIdleClosed = RegisterDBStats("database_max_idle_closed")
	h.idleConnections = RegisterDBStats("database_idle_connections")
	h.maxLifetimeClosed = RegisterDBStats("database_max_lifetime_closed")
	h.inUse = RegisterDBStats("database_in_use")
	h.waitCount = RegisterDBStats("database_wait_count")
	h.waitDuration = RegisterDBStats("database_wait_duration")

	h.queryHistogram = RegisterHistogram("database_query_timing")

	return h
}

func (h *QueryHook) Init(db *bun.DB) {
	ticker := time.NewTicker(2 * time.Second)

	labels := make(map[string]string, 0)
	if sys := dbSystem(db); sys != "" {
		labels["system"] = sys
	}
	labels["database"] = h.dbName

	go func() {
		for range ticker.C {

			stats := db.DB.Stats()
			h.connections.With(labels).Set(float64(stats.OpenConnections))
			h.maxOpenConnections.With(labels).Set(float64(stats.MaxOpenConnections))
			h.idleConnections.With(labels).Set(float64(stats.Idle))
			h.maxIdleTimeClosed.With(labels).Set(float64(stats.MaxIdleTimeClosed))
			h.maxIdleClosed.With(labels).Set(float64(stats.MaxIdleClosed))
			h.maxLifetimeClosed.With(labels).Set(float64(stats.MaxLifetimeClosed))
			h.inUse.With(labels).Set(float64(stats.InUse))
			h.waitCount.With(labels).Set(float64(stats.WaitCount))
			h.waitDuration.With(labels).Set(float64(stats.WaitDuration))
		}
	}()
}

func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	labels := map[string]string{
		"system":    dbSystem(event.DB),
		"database":  h.dbName,
		"operation": event.Operation(),
	}

	if event.IQuery != nil {
		if tableName := event.IQuery.GetTableName(); tableName != "" {
			labels["table"] = tableName
		}
	}

	dur := time.Since(event.StartTime)
	h.queryHistogram.With(labels).Observe(float64(dur.Milliseconds()))
}

func dbSystem(db *bun.DB) string {
	switch db.Dialect().Name() {
	case dialect.PG:
		return "postgresql"
	case dialect.MySQL:
		return "mysql"
	case dialect.SQLite:
		return "sqlite"
	case dialect.MSSQL:
		return "mssql"
	default:
		return ""
	}
}
