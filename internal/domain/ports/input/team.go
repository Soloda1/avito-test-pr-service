package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name TeamInputPort --dir . --output ../../../../mocks --outpkg mocks --with-expecter --filename TeamInputPort.go

type TeamInputPort interface {
	CreateTeam(ctx context.Context, name string) (*models.Team, error)
	CreateTeamWithMembers(ctx context.Context, name string, members []*models.User) (*models.Team, []*models.User, error)
	AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeams(ctx context.Context) ([]*models.Team, error)
}
