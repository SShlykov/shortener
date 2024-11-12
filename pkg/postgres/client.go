package postgres

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sshlykov/shortener/pkg/logger"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type pgClient struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	db DB
}

// NewPool создает новый пул соединений к базе данных.
// ctx - контекст; dsn - строка подключения к базе данных;
// maxPoolSize - максимальное количество соединений в пуле.
func NewPool(ctx context.Context, dsn string, maxPoolSize int) (Client, error) {
	client := &pgClient{
		maxPoolSize:  maxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("Cant parse dsn", slog.String("dsn", dsn))
		return nil, err
	}

	poolConfig.MaxConns = int32(_defaultMaxPoolSize)

	return client.Connect(ctx, poolConfig)
}

func NewClient(ctx context.Context, dsn string) (Client, error) {
	client := &pgClient{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("Cant parse dsn", slog.String("dsn", dsn))
		return nil, err
	}

	poolConfig.MaxConns = int32(_defaultMaxPoolSize)

	return client.Connect(ctx, poolConfig)
}

func (c *pgClient) Connect(ctx context.Context, poolConfig *pgxpool.Config) (Client, error) {
	logger.Debug(
		ctx,
		"connecting to db",
		slog.Int("attempts", c.connAttempts),
		slog.String("dsn", poolConfig.ConnString()),
		slog.Int("maxPoolSize", _defaultMaxPoolSize),
	)
	for c.connAttempts > 0 {
		var pool *pgxpool.Pool
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			c.db = NewDB(pool)
			return c, nil
		}

		time.Sleep(c.connTimeout)

		c.connAttempts--
	}

	return nil, errors.New("failed to connect to db")
}

func (c *pgClient) Exec(ctx context.Context, query string, attrs ...interface{}) (pgconn.CommandTag, error) {
	q := Query{Name: "exec", Raw: query}
	return c.db.ExecContext(ctx, q, attrs...)
}

func (c *pgClient) Query(ctx context.Context, query string, attrs ...interface{}) (pgx.Rows, error) {
	q := Query{Name: "exec", Raw: query}
	return c.db.QueryRawContextMulti(ctx, q, attrs...)
}

func (c *pgClient) QueryRow(ctx context.Context, query string, attrs ...interface{}) pgx.Row {
	q := Query{Name: "exec", Raw: query}
	return c.db.QueryRowContext(ctx, q, attrs...)
}

func (c *pgClient) DB() DB {
	return c.db
}

func (c *pgClient) Close() error {
	if c.db != nil {
		c.db.Close()
	}

	return nil
}
