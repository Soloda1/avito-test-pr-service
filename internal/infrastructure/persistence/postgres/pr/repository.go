package pr_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	"avito-test-pr-service/internal/domain/ports/output/pr"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

type PRRepository struct {
	querier postgres.Querier
	log     ports.Logger
}

func NewPRRepository(querier postgres.Querier, log ports.Logger) pr.PRRepository {
	return &PRRepository{querier: querier, log: log}
}

func (r *PRRepository) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	return errors.New("not implemented")
}

func (r *PRRepository) GetPRByID(ctx context.Context, id uuid.UUID) (*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (r *PRRepository) LockPRByID(ctx context.Context, id uuid.UUID) (*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (r *PRRepository) AddReviewer(ctx context.Context, prID uuid.UUID, reviewerID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID uuid.UUID, reviewerID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *PRRepository) UpdateStatus(ctx context.Context, prID uuid.UUID, status models.PRStatus, mergedAt *time.Time) error {
	return errors.New("not implemented")
}

func (r *PRRepository) ListPRsByReviewer(ctx context.Context, reviewerID uuid.UUID, status *models.PRStatus) ([]*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (r *PRRepository) CountReviewersByPRID(ctx context.Context, prID uuid.UUID) (int, error) {
	return 0, errors.New("not implemented")
}
