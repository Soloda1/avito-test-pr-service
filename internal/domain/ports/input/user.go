package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type UserInputPort interface {
	CreateUser(ctx context.Context, name string, isActive bool) (*models.User, error)
	UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListUsers(ctx context.Context) ([]*models.User, error)
}
