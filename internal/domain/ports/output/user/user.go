package user

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name UserRepository --dir . --output ../../../../../mocks --outpkg mocks --with-expecter --filename UserRepository.go

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error
	ListUsers(ctx context.Context) ([]*models.User, error)
	UpdateUserName(ctx context.Context, id uuid.UUID, name string) error
	GetTeamIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	ListActiveMembersByTeamID(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error)
}
