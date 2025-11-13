package user_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	"avito-test-pr-service/internal/domain/ports/output/user"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"context"
	"errors"

	"github.com/google/uuid"
)

type UserRepository struct {
	querier postgres.Querier
	log     ports.Logger
}

func NewUserRepository(querier postgres.Querier, log ports.Logger) user.UserRepository {
	return &UserRepository{querier: querier, log: log}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return errors.New("not implemented")
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepository) UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	return errors.New("not implemented")
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepository) GetTeamIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, errors.New("not implemented")
}

func (r *UserRepository) ListActiveMembersByTeamID(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error) {
	return nil, errors.New("not implemented")
}
