package pr

import (
	"avito-test-pr-service/internal/domain/models"
	"context"
	"time"

	"github.com/google/uuid"
)

//go:generate mockery --name PRRepository --dir . --output ../../../../../mocks --outpkg mocks --with-expecter --filename PRRepository.go

type PRRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPRByID(ctx context.Context, id string) (*models.PullRequest, error)
	LockPRByID(ctx context.Context, id string) (*models.PullRequest, error)
	AddReviewer(ctx context.Context, prID string, reviewerID uuid.UUID) error
	RemoveReviewer(ctx context.Context, prID string, reviewerID uuid.UUID) error
	UpdateStatus(ctx context.Context, prID string, status models.PRStatus, mergedAt *time.Time) error
	ListPRsByReviewer(ctx context.Context, reviewerID uuid.UUID, status *models.PRStatus) ([]*models.PullRequest, error)
	CountReviewersByPRID(ctx context.Context, prID string) (int, error)
}
