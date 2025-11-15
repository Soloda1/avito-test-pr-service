package user

import (
	"avito-test-pr-service/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

//go:generate mockery --name UserRepository --dir . --output ../../../../../mocks --outpkg mocks --with-expecter --filename UserRepository.go

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUserActive(ctx context.Context, id string, isActive bool) error
	ListUsers(ctx context.Context) ([]*models.User, error)
	UpdateUserName(ctx context.Context, id string, name string) error
	GetTeamIDByUserID(ctx context.Context, userID string) (uuid.UUID, error)
	ListActiveMembersByTeamID(ctx context.Context, teamID uuid.UUID) ([]string, error)
	ListMembersByTeamID(ctx context.Context, teamID uuid.UUID) ([]*models.User, error)
}
