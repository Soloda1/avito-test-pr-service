package uow

import (
	pr "avito-test-pr-service/internal/domain/ports/output/pr"
	team "avito-test-pr-service/internal/domain/ports/output/team"
	user "avito-test-pr-service/internal/domain/ports/output/user"
	"context"
)

type UnitOfWork interface {
	Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	UserRepository() user.UserRepository
	TeamRepository() team.TeamRepository
	PRRepository() pr.PRRepository
}
