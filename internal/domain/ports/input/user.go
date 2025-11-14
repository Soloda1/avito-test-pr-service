package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name UserInputPort --dir . --output ../../../../mocks --outpkg mocks --with-expecter --filename UserInputPort.go

type UserInputPort interface {
	CreateUser(ctx context.Context, name string, isActive bool) (*models.User, error)
	UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListUsers(ctx context.Context) ([]*models.User, error)
	GetUserTeamName(ctx context.Context, id uuid.UUID) (string, error)
}
