package pr

import (
	"avito-test-pr-service/internal/domain/models"
	"context"
	"time"
)

//go:generate mockery --name PRRepository --dir . --output ../../../../../mocks --outpkg mocks --with-expecter --filename PRRepository.go

type PRRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPRByID(ctx context.Context, id string) (*models.PullRequest, error)
	LockPRByID(ctx context.Context, id string) (*models.PullRequest, error)
	AddReviewer(ctx context.Context, prID string, reviewerID string) error
	RemoveReviewer(ctx context.Context, prID string, reviewerID string) error
	UpdateStatus(ctx context.Context, prID string, status models.PRStatus, mergedAt *time.Time) error
	ListPRsByReviewer(ctx context.Context, reviewerID string, status *models.PRStatus) ([]*models.PullRequest, error)
	CountReviewersByPRID(ctx context.Context, prID string) (int, error)
}
