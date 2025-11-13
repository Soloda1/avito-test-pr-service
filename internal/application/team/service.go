package team

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

func NewService(uow uow.UnitOfWork, log ports.Logger) input.TeamInputPort {
	return &Service{uow: uow, log: log}
}

func (s *Service) CreateTeam(ctx context.Context, name string) (*models.Team, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (s *Service) RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (s *Service) GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) ListTeams(ctx context.Context) ([]*models.Team, error) {
	return nil, errors.New("not implemented")
}
