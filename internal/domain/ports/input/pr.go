package input

import (
	"avito-test-pr-service/internal/domain/models"
	"context"
)

//go:generate mockery --name PRInputPort --dir . --output ../../../../mocks --outpkg mocks --with-expecter --filename PRInputPort.go

type PRInputPort interface {
	CreatePR(ctx context.Context, prID string, authorID string, title string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	GetPR(ctx context.Context, prID string) (*models.PullRequest, error)
	ListPRsByAssignee(ctx context.Context, reviewerID string, status *models.PRStatus) ([]*models.PullRequest, error)
}
