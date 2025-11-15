package team

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name TeamRepository --dir . --output ../../../../../mocks --outpkg mocks --with-expecter --filename TeamRepository.go

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
	ListTeams(ctx context.Context) ([]*models.Team, error)
	AddMember(ctx context.Context, teamID uuid.UUID, userID string) error
	RemoveMember(ctx context.Context, teamID uuid.UUID, userID string) error
}
