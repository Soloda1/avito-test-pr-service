package team

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeams(ctx context.Context) ([]*models.Team, error)
	AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
}
