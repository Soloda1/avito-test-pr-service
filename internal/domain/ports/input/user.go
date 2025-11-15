package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"
)

//go:generate mockery --name UserInputPort --dir . --output ../../../../mocks --outpkg mocks --with-expecter --filename UserInputPort.go

type UserInputPort interface {
	CreateUser(ctx context.Context, id string, name string, isActive bool) (*models.User, error)
	UpdateUserActive(ctx context.Context, id string, isActive bool) error
	UpdateUserName(ctx context.Context, id string, name string) error
	GetUser(ctx context.Context, id string) (*models.User, error)
	ListUsers(ctx context.Context) ([]*models.User, error)
	GetUserTeamName(ctx context.Context, id string) (string, error)
	ListMembersByTeamID(ctx context.Context, teamID string) ([]*models.User, error)
}
