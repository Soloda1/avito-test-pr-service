package integration

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TruncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE pr_reviewers, team_members, prs, users RESTART IDENTITY CASCADE;
	`)
	return err
}
