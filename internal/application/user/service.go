package user

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
	uow "avito-test-pr-service/internal/domain/ports/output/uow"
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	uow uow.UnitOfWork
	log ports.Logger
}

func NewService(uow uow.UnitOfWork, log ports.Logger) input.UserInputPort {
	return &Service{uow: uow, log: log}
}

func (s *Service) CreateUser(ctx context.Context, name string, isActive bool) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	return errors.New("not implemented")
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) ListUsers(ctx context.Context) ([]*models.User, error) {
	return nil, errors.New("not implemented")
}
