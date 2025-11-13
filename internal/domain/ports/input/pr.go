package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name PRInputPort --dir . --output ../../../../mocks --outpkg mocks --with-expecter --filename PRInputPort.go

type PRInputPort interface {
	CreatePR(ctx context.Context, authorID uuid.UUID, title string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID uuid.UUID, oldReviewerID uuid.UUID) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error)
	GetPR(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error)
	ListPRsByAssignee(ctx context.Context, reviewerID uuid.UUID, status *models.PRStatus) ([]*models.PullRequest, error)
}
