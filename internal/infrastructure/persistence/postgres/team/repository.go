package team_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	"avito-test-pr-service/internal/domain/ports/output/team"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"context"
	"errors"

	"github.com/google/uuid"
)

type TeamRepository struct {
	querier postgres.Querier
	log     ports.Logger
}

func NewTeamRepository(querier postgres.Querier, log ports.Logger) team.TeamRepository {
	return &TeamRepository{querier: querier, log: log}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	return errors.New("not implemented")
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	return nil, errors.New("not implemented")
}

func (r *TeamRepository) ListTeams(ctx context.Context) ([]*models.Team, error) {
	return nil, errors.New("not implemented")
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	return errors.New("not implemented")
}
