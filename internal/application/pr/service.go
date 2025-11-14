package pr

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
	uow "avito-test-pr-service/internal/domain/ports/output/uow"
	"avito-test-pr-service/internal/domain/services"
	"avito-test-pr-service/internal/utils"
	"context"
	"time"

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
	if authorID == uuid.Nil || title == "" {
		return nil, utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("CreatePR begin tx failed", "err", err, "author_id", authorID)
		return nil, err
	}

	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	userRepo := tx.UserRepository()
	if _, err := userRepo.GetUserByID(ctx, authorID); err != nil {
		s.log.Error("CreatePR author fetch failed", "err", err, "author_id", authorID)
		return nil, err
	}

	teamID, err := userRepo.GetTeamIDByUserID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	candidates, err := userRepo.ListActiveMembersByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	filtered := utils.FilterUUIDs(candidates, map[uuid.UUID]struct{}{authorID: {}})

	selected := s.selector.Select(filtered, 2)

	prRepo := tx.PRRepository()
	pr := &models.PullRequest{
		ID:          uuid.New(),
		Title:       title,
		AuthorID:    authorID,
		Status:      models.PRStatusOPEN,
		ReviewerIDs: selected,
	}
	if err := prRepo.CreatePR(ctx, pr); err != nil {
		s.log.Error("CreatePR repo failed", "err", err, "author_id", authorID)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("CreatePR commit failed", "err", err, "pr_id", pr.ID)
		return nil, err
	}

	commit = true
	s.log.Info("CreatePR success", "pr_id", pr.ID, "author_id", authorID)
	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID uuid.UUID, oldReviewerID uuid.UUID) (*models.PullRequest, error) {
	if prID == uuid.Nil || oldReviewerID == uuid.Nil {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("Reassign begin tx failed", "err", err, "pr_id", prID)
		return nil, err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	prRepo := tx.PRRepository()
	pr, err := prRepo.LockPRByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.Status == models.PRStatusMERGED {
		return nil, utils.ErrAlreadyMerged
	}
	if !utils.ContainsUUID(pr.ReviewerIDs, oldReviewerID) {
		return nil, utils.ErrReviewerNotAssigned
	}

	userRepo := tx.UserRepository()
	teamID, err := userRepo.GetTeamIDByUserID(ctx, pr.AuthorID)
	if err != nil {
		return nil, err
	}
	members, err := userRepo.ListActiveMembersByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	ex := make(map[uuid.UUID]struct{}, len(pr.ReviewerIDs)+1)
	ex[pr.AuthorID] = struct{}{}
	for _, id := range pr.ReviewerIDs {
		ex[id] = struct{}{}
	}
	pool := utils.FilterUUIDs(members, ex)
	if len(pool) == 0 {
		return nil, utils.ErrNoReplacementCandidates
	}
	picked := s.selector.Select(pool, 1)
	if len(picked) == 0 {
		return nil, utils.ErrNoReplacementCandidates
	}
	newReviewerID := picked[0]
	if err := prRepo.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		return nil, err
	}
	if err := prRepo.AddReviewer(ctx, prID, newReviewerID); err != nil {
		return nil, err
	}
	for i, id := range pr.ReviewerIDs {
		if id == oldReviewerID {
			pr.ReviewerIDs[i] = newReviewerID
			break
		}
	}
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("Reassign commit failed", "err", err, "pr_id", prID)
		return nil, err
	}
	commit = true
	return pr, nil
}

func (s *Service) MergePR(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error) {
	if prID == uuid.Nil {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("MergePR begin tx failed", "err", err, "pr_id", prID)
		return nil, err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	prRepo := tx.PRRepository()
	pr, err := prRepo.LockPRByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.Status == models.PRStatusMERGED {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		commit = true
		return pr, nil
	}

	now := time.Now().UTC()
	if err := prRepo.UpdateStatus(ctx, prID, models.PRStatusMERGED, &now); err != nil {
		return nil, err
	}
	pr.Status = models.PRStatusMERGED
	pr.MergedAt = &now

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("MergePR commit failed", "err", err, "pr_id", prID)
		return nil, err
	}
	commit = true
	return pr, nil
}

func (s *Service) GetPR(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error) {
	if prID == uuid.Nil {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("GetPR begin tx failed", "err", err, "pr_id", prID)
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	prRepo := tx.PRRepository()
	pr, err := prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) ListPRsByAssignee(ctx context.Context, reviewerID uuid.UUID, status *models.PRStatus) ([]*models.PullRequest, error) {
	if reviewerID == uuid.Nil {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("ListPRs begin tx failed", "err", err, "reviewer_id", reviewerID)
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	prRepo := tx.PRRepository()
	res, err := prRepo.ListPRsByReviewer(ctx, reviewerID, status)
	if err != nil {
		return nil, err
	}

	return res, nil
}
