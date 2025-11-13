package pr

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
	uow "avito-test-pr-service/internal/domain/ports/output/uow"
	"avito-test-pr-service/internal/domain/services"
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	uow      uow.UnitOfWork
	selector services.ReviewerSelector
	log      ports.Logger
}

func NewService(uow uow.UnitOfWork, selector services.ReviewerSelector, log ports.Logger) input.PRInputPort {
	return &Service{uow: uow, selector: selector, log: log}
}

func (s *Service) CreatePR(ctx context.Context, authorID uuid.UUID, title string) (*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) ReassignReviewer(ctx context.Context, prID uuid.UUID, oldReviewerID uuid.UUID) (*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) MergePR(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) GetPR(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) ListPRsByAssignee(ctx context.Context, reviewerID uuid.UUID, status *models.PRStatus) ([]*models.PullRequest, error) {
	return nil, errors.New("not implemented")
}
