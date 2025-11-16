package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DSN       string
	Pool      *pgxpool.Pool
}

func StartPostgres(ctx context.Context) (*PostgresContainer, error) {
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("prservice"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("admin"),
		testcontainers.WithWaitStrategy(
			wait.ForSQL(nat.Port("5432/tcp"), "pgx", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", "postgres", "admin", host, port.Port(), "prservice")
			}).WithQuery("SELECT 1").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	dsn, err := pg.ConnectionString(ctx)
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, err
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, err
	}
	cfg.MaxConns = 5
	cfg.MinConns = 1
	cfg.HealthCheckPeriod = 2 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, err
	}

	return &PostgresContainer{Container: pg, DSN: dsn, Pool: pool}, nil
}

func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}
