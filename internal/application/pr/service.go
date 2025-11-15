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

func (s *Service) CreatePR(ctx context.Context, prID string, authorID uuid.UUID, title string) (*models.PullRequest, error) {
	if prID == "" || authorID == uuid.Nil || title == "" {
		return nil, utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("CreatePR begin tx failed", "err", err, "author_id", authorID, "pr_id", prID)
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
		s.log.Error("CreatePR get team failed", "err", err, "author_id", authorID)
		return nil, err
	}

	candidates, err := userRepo.ListActiveMembersByTeamID(ctx, teamID)
	if err != nil {
		s.log.Error("CreatePR list candidates failed", "err", err, "author_id", authorID, "team_id", teamID)
		return nil, err
	}

	filtered := utils.FilterUUIDs(candidates, map[uuid.UUID]struct{}{authorID: {}})

	selected := s.selector.Select(filtered, 2)
	if selected == nil {
		s.log.Info("CreatePR no reviewers available", "author_id", authorID)
		selected = []uuid.UUID{}
	}

	prRepo := tx.PRRepository()
	pr := &models.PullRequest{
		ID:          prID,
		Title:       title,
		AuthorID:    authorID,
		Status:      models.PRStatusOPEN,
		ReviewerIDs: selected,
	}
	if err := prRepo.CreatePR(ctx, pr); err != nil {
		s.log.Error("CreatePR repo failed", "err", err, "author_id", authorID, "pr_id", prID)
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

func (s *Service) ReassignReviewer(ctx context.Context, prID string, oldReviewerID uuid.UUID) (*models.PullRequest, error) {
	if prID == "" || oldReviewerID == uuid.Nil {
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
		s.log.Error("Reassign lock failed", "err", err, "pr_id", prID)
		return nil, err
	}
	if pr.Status == models.PRStatusMERGED {
		s.log.Error("Reassign on merged PR", "err", utils.ErrAlreadyMerged, "pr_id", prID)
		return nil, utils.ErrAlreadyMerged
	}
	if !utils.ContainsUUID(pr.ReviewerIDs, oldReviewerID) {
		s.log.Error("Reassign reviewer not assigned", "err", utils.ErrReviewerNotAssigned, "pr_id", prID, "old_reviewer_id", oldReviewerID)
		return nil, utils.ErrReviewerNotAssigned
	}

	userRepo := tx.UserRepository()
	teamID, err := userRepo.GetTeamIDByUserID(ctx, pr.AuthorID)
	if err != nil {
		s.log.Error("Reassign get author team failed", "err", err, "pr_id", prID, "author_id", pr.AuthorID)
		return nil, err
	}
	members, err := userRepo.ListActiveMembersByTeamID(ctx, teamID)
	if err != nil {
		s.log.Error("Reassign list team members failed", "err", err, "pr_id", prID, "team_id", teamID)
		return nil, err
	}
	ex := make(map[uuid.UUID]struct{}, len(pr.ReviewerIDs)+1)
	ex[pr.AuthorID] = struct{}{}
	for _, id := range pr.ReviewerIDs {
		ex[id] = struct{}{}
	}
	pool := utils.FilterUUIDs(members, ex)
	if len(pool) == 0 {
		s.log.Error("Reassign no replacement candidates", "err", utils.ErrNoReplacementCandidates, "pr_id", prID)
		return nil, utils.ErrNoReplacementCandidates
	}
	picked := s.selector.Select(pool, 1)
	if len(picked) == 0 {
		s.log.Error("Reassign selector returned no candidate", "err", utils.ErrNoReplacementCandidates, "pr_id", prID)
		return nil, utils.ErrNoReplacementCandidates
	}
	newReviewerID := picked[0]
	if err := prRepo.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		s.log.Error("Reassign remove reviewer failed", "err", err, "pr_id", prID, "old_reviewer_id", oldReviewerID)
		return nil, err
	}
	if err := prRepo.AddReviewer(ctx, prID, newReviewerID); err != nil {
		s.log.Error("Reassign add reviewer failed", "err", err, "pr_id", prID, "new_reviewer_id", newReviewerID)
		return nil, err
	}
	updatedPR, err := prRepo.GetPRByID(ctx, prID)
	if err != nil {
		s.log.Error("Reassign refetch failed", "err", err, "pr_id", prID)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("Reassign commit failed", "err", err, "pr_id", prID)
		return nil, err
	}
	commit = true
	s.log.Info("Reassign success", "pr_id", prID, "old_reviewer_id", oldReviewerID, "new_reviewer_id", newReviewerID)
	return updatedPR, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
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
		s.log.Error("MergePR lock failed", "err", err, "pr_id", prID)
		return nil, err
	}
	if pr.Status == models.PRStatusMERGED {
		s.log.Info("MergePR idempotent already merged", "pr_id", prID)
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		commit = true
		return pr, nil
	}

	now := time.Now().UTC()
	if err := prRepo.UpdateStatus(ctx, prID, models.PRStatusMERGED, &now); err != nil {
		s.log.Error("MergePR update status failed", "err", err, "pr_id", prID)
		return nil, err
	}
	pr.Status = models.PRStatusMERGED
	pr.MergedAt = &now

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("MergePR commit failed", "err", err, "pr_id", prID)
		return nil, err
	}
	commit = true
	s.log.Info("MergePR success", "pr_id", prID)
	return pr, nil
}

func (s *Service) GetPR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
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
		s.log.Error("GetPR repo failed", "err", err, "pr_id", prID)
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
		s.log.Error("ListPRs repo failed", "err", err, "reviewer_id", reviewerID)
		return nil, err
	}

	return res, nil
}
