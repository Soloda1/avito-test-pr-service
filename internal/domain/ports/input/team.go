package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type TeamInputPort interface {
	CreateTeam(ctx context.Context, name string) (*models.Team, error)
	AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeams(ctx context.Context) ([]*models.Team, error)
}
