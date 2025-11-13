package uow

import (
	ports "avito-test-pr-service/internal/domain/ports/output"
	pr_port "avito-test-pr-service/internal/domain/ports/output/pr"
	team_port "avito-test-pr-service/internal/domain/ports/output/team"
	user_port "avito-test-pr-service/internal/domain/ports/output/user"

	"avito-test-pr-service/internal/domain/ports/output/uow"
	pr_repo "avito-test-pr-service/internal/infrastructure/persistence/postgres/pr"
	team_repo "avito-test-pr-service/internal/infrastructure/persistence/postgres/team"
	user_repo "avito-test-pr-service/internal/infrastructure/persistence/postgres/user"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUnitOfWork struct {
	pool *pgxpool.Pool
	log  ports.Logger
}

func NewPostgresUOW(pool *pgxpool.Pool, log ports.Logger) uow.UnitOfWork {
	return &PostgresUnitOfWork{pool: pool, log: log}
}

func (uow *PostgresUnitOfWork) Begin(ctx context.Context) (uow.Transaction, error) {
	tx, err := uow.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	return &PostgresTransaction{tx: tx, log: uow.log}, nil
}

type PostgresTransaction struct {
	tx  pgx.Tx
	log ports.Logger
}

func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *PostgresTransaction) UserRepository() user_port.UserRepository {
	return user_repo.NewUserRepository(t.tx, t.log)
}

func (t *PostgresTransaction) TeamRepository() team_port.TeamRepository {
	return team_repo.NewTeamRepository(t.tx, t.log)
}

func (t *PostgresTransaction) PRRepository() pr_port.PRRepository {
	return pr_repo.NewPRRepository(t.tx, t.log)
}
