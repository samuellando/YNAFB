package querycount

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"

	"samuellando.com/YNAFB/data"
)

// Counter tracks how many queries ran for a single request and how long
// they took in total. Use NewContext per request, then read Value() and
// QueryTime() after the handler runs.
type Counter struct {
	n     atomic.Int64
	nanos atomic.Int64
}

func (c *Counter) Inc() {
	c.n.Add(1)
}

func (c *Counter) Value() int64 {
	return c.n.Load()
}

// Add accumulates query latency.
func (c *Counter) Add(d time.Duration) {
	c.nanos.Add(int64(d))
}

// QueryTime returns the total time spent waiting for queries.
func (c *Counter) QueryTime() time.Duration {
	return time.Duration(c.nanos.Load())
}

type contextKey struct{}

// NewContext returns a child context carrying a fresh Counter.
func NewContext(ctx context.Context) (context.Context, *Counter) {
	c := &Counter{}
	return context.WithValue(ctx, contextKey{}, c), c
}

// FromContext returns the Counter stored in ctx, if any.
func FromContext(ctx context.Context) (*Counter, bool) {
	c, ok := ctx.Value(contextKey{}).(*Counter)
	return c, ok && c != nil
}

func record(ctx context.Context, d time.Duration) {
	if c, ok := FromContext(ctx); ok {
		c.Inc()
		c.Add(d)
	}
}

// CountingDB wraps *sql.DB and counts every query made through it.
// It implements data.DBTX. Queries inside a *sql.Tx bypass it
// (WithTx takes a concrete *sql.Tx), so Tx-inner queries are not counted.
type CountingDB struct {
	DB *sql.DB
}

var _ data.DBTX = (*CountingDB)(nil)

func New(db *sql.DB) *CountingDB {
	return &CountingDB{DB: db}
}

func (d *CountingDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := d.DB.ExecContext(ctx, query, args...)
	record(ctx, time.Since(start))
	return res, err
}

func (d *CountingDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	start := time.Now()
	stmt, err := d.DB.PrepareContext(ctx, query)
	record(ctx, time.Since(start))
	return stmt, err
}

func (d *CountingDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.DB.QueryContext(ctx, query, args...)
	record(ctx, time.Since(start))
	return rows, err
}

func (d *CountingDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Note: database/sql defers QueryRow execution until Scan, which we
	// can't intercept while returning *sql.Row, so this only times row
	// creation, not execution.
	start := time.Now()
	row := d.DB.QueryRowContext(ctx, query, args...)
	record(ctx, time.Since(start))
	return row
}
