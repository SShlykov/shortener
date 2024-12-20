package postgres

import (
	"context"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sshlykov/shortener/pkg/logger"
)

type key string

const (
	TxKey key = "tx"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewDB(dbc *pgxpool.Pool) DB {
	return &Postgres{Pool: dbc}
}

func (p *Postgres) ScanSingleContext(ctx context.Context, q Query, dest interface{}, args ...interface{}) error { //nolint:gofmt
	logQuery(ctx, q, args...)

	row, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (p *Postgres) ScanAllContext(ctx context.Context, q Query, dest interface{}, args ...interface{}) error { //nolint:gofmt
	logQuery(ctx, q, args...)

	rows, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (p *Postgres) ExecContext(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error) { //nolint:gofmt
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.Raw, args...)
	}

	return p.Pool.Exec(ctx, q.Raw, args...)
}

func (p *Postgres) QueryContext(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error) {
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.Raw, args...)
	}

	return p.Pool.Query(ctx, q.Raw, args...)
}

func (p *Postgres) QueryRowContext(ctx context.Context, q Query, args ...interface{}) pgx.Row {
	logQuery(ctx, q, args...)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.Raw, args...)
	}

	res := p.Pool.QueryRow(ctx, q.Raw, args...)

	return res
}

func (p *Postgres) QueryRawContextMulti(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.Raw, args...)
	}

	return p.Pool.Query(ctx, q.Raw, args...)
}

func (p *Postgres) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.Pool.BeginTx(ctx, txOptions)
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func logQuery(ctx context.Context, q Query, args ...interface{}) {
	attrs := []any{
		slog.String("sql", q.Name),
		slog.String("query", q.Raw),
		slog.Any("args", args),
	}

	logger.Debug(ctx, "executing query", attrs...)
}
