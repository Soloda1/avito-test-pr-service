package integration

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	prrepo "avito-test-pr-service/internal/infrastructure/persistence/postgres/pr"
	"avito-test-pr-service/internal/utils"
	"testing"
	"time"
)

func TestPRRepository_Integration(t *testing.T) {
	ctx := TestCtx
	log := logger.New("test")
	repo := prrepo.NewPRRepository(PGC.Pool, log)

	insertUser := func(t *testing.T, id, name string, active bool) {
		_, err := PGC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}

	t.Run("CreatePR success with reviewers", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		insertUser(t, "u-r2", "r2", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author", ReviewerIDs: []string{"u-r1", "u-r2"}}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if len(pr.ReviewerIDs) != 2 {
			t.Fatalf("expected 2 reviewers, got %d", len(pr.ReviewerIDs))
		}
		if pr.Status != models.PRStatusOPEN {
			t.Fatalf("expected status OPEN, got %s", pr.Status)
		}
	})

	t.Run("CreatePR duplicate id -> ErrPRExists", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		pr1 := &models.PullRequest{ID: "pr-1", Title: "feat", AuthorID: "u-author"}
		if err := repo.CreatePR(ctx, pr1); err != nil {
			t.Fatalf("create first: %v", err)
		}
		pr2 := &models.PullRequest{ID: "pr-1", Title: "feat2", AuthorID: "u-author"}
		err := repo.CreatePR(ctx, pr2)
		if err == nil || err != utils.ErrPRExists {
			t.Fatalf("expected ErrPRExists, got %v", err)
		}
	})

	t.Run("CreatePR author FK violation -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "missing-author"}
		err := repo.CreatePR(ctx, pr)
		if err == nil || err != utils.ErrUserNotFound {
			t.Fatalf("expected ErrUserNotFound got %v", err)
		}
	})

	t.Run("CreatePR invalid arg (missing fields)", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		pr := &models.PullRequest{ID: "", Title: "x", AuthorID: "u-author"}
		err := repo.CreatePR(ctx, pr)
		if err == nil || err != utils.ErrInvalidArgument {
			t.Fatalf("expected ErrInvalidArgument got %v", err)
		}
	})

	t.Run("GetPRByID success", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author", ReviewerIDs: []string{"u-r1"}}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		got, err := repo.GetPRByID(ctx, "pr-1")
		if err != nil {
			t.Fatalf("GetPRByID: %v", err)
		}
		if got.ID != pr.ID || len(got.ReviewerIDs) != 1 {
			t.Fatalf("unexpected pr: %+v", got)
		}
	})

	t.Run("GetPRByID not found -> ErrPRNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		_, err := repo.GetPRByID(ctx, "pr-x")
		if err == nil || err != utils.ErrPRNotFound {
			t.Fatalf("expected ErrPRNotFound got %v", err)
		}
	})

	t.Run("AddReviewer success", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		insertUser(t, "u-r2", "r2", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author", ReviewerIDs: []string{"u-r1"}}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if err := repo.AddReviewer(ctx, "pr-1", "u-r2"); err != nil {
			t.Fatalf("AddReviewer: %v", err)
		}
		got, err := repo.GetPRByID(ctx, "pr-1")
		if err != nil {
			t.Fatalf("GetPRByID: %v", err)
		}
		if len(got.ReviewerIDs) != 2 {
			t.Fatalf("expected 2 reviewers, got %d", len(got.ReviewerIDs))
		}
	})

	t.Run("AddReviewer too many -> ErrTooManyReviewers", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		insertUser(t, "u-r2", "r2", true)
		insertUser(t, "u-r3", "r3", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author", ReviewerIDs: []string{"u-r1", "u-r2"}}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		err := repo.AddReviewer(ctx, "pr-1", "u-r3")
		if err == nil || err != utils.ErrTooManyReviewers {
			t.Fatalf("expected ErrTooManyReviewers got %v", err)
		}
	})

	t.Run("AddReviewer duplicate -> ErrReviewerAlreadyAssigned", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author", ReviewerIDs: []string{"u-r1"}}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		err := repo.AddReviewer(ctx, "pr-1", "u-r1")
		if err == nil || err != utils.ErrReviewerAlreadyAssigned {
			t.Fatalf("expected ErrReviewerAlreadyAssigned got %v", err)
		}
	})

	t.Run("AddReviewer user not found -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author"}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		err := repo.AddReviewer(ctx, "pr-1", "missing-user")
		if err == nil || err != utils.ErrUserNotFound {
			t.Fatalf("expected ErrUserNotFound got %v", err)
		}
	})

	t.Run("AddReviewer PR not found -> ErrPRNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-r1", "r1", true)
		err := repo.AddReviewer(ctx, "pr-no", "u-r1")
		if err == nil || err != utils.ErrPRNotFound {
			t.Fatalf("expected ErrPRNotFound got %v", err)
		}
	})

	t.Run("RemoveReviewer success", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author", ReviewerIDs: []string{"u-r1"}}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if err := repo.RemoveReviewer(ctx, "pr-1", "u-r1"); err != nil {
			t.Fatalf("RemoveReviewer: %v", err)
		}
		got, err := repo.GetPRByID(ctx, "pr-1")
		if err != nil {
			t.Fatalf("get PR: %v", err)
		}
		if len(got.ReviewerIDs) != 0 {
			t.Fatalf("expected 0 reviewers got %d", len(got.ReviewerIDs))
		}
	})

	t.Run("RemoveReviewer not assigned -> ErrReviewerNotAssigned", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author"}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		err := repo.RemoveReviewer(ctx, "pr-1", "u-rX")
		if err == nil || err != utils.ErrReviewerNotAssigned {
			t.Fatalf("expected ErrReviewerNotAssigned got %v", err)
		}
	})

	t.Run("UpdateStatus merge success", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author"}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		mergedAt := time.Now()
		if err := repo.UpdateStatus(ctx, "pr-1", models.PRStatusMERGED, &mergedAt); err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
		got, err := repo.GetPRByID(ctx, "pr-1")
		if err != nil {
			t.Fatalf("GetPRByID: %v", err)
		}
		if got.Status != models.PRStatusMERGED || got.MergedAt == nil {
			t.Fatalf("expected merged, got %+v", got)
		}
	})

	t.Run("UpdateStatus already merged -> ErrAlreadyMerged", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		pr := &models.PullRequest{ID: "pr-1", Title: "feature", AuthorID: "u-author"}
		if err := repo.CreatePR(ctx, pr); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		mergedAt := time.Now()
		if err := repo.UpdateStatus(ctx, "pr-1", models.PRStatusMERGED, &mergedAt); err != nil {
			t.Fatalf("first merge: %v", err)
		}
		err := repo.UpdateStatus(ctx, "pr-1", models.PRStatusMERGED, &mergedAt)
		if err == nil || err != utils.ErrAlreadyMerged {
			t.Fatalf("expected ErrAlreadyMerged got %v", err)
		}
	})

	t.Run("UpdateStatus invalid status -> ErrInvalidStatus", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		err := repo.UpdateStatus(ctx, "pr-1", "WRONG", nil)
		if err == nil || err != utils.ErrInvalidStatus {
			t.Fatalf("expected ErrInvalidStatus got %v", err)
		}
	})

	t.Run("UpdateStatus PR not found -> ErrPRNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		err := repo.UpdateStatus(ctx, "pr-no", models.PRStatusMERGED, nil)
		if err == nil || err != utils.ErrPRNotFound {
			t.Fatalf("expected ErrPRNotFound got %v", err)
		}
	})

	t.Run("ListPRsByReviewer success and filter", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-author", "author", true)
		insertUser(t, "u-r1", "r1", true)
		pr1 := &models.PullRequest{ID: "pr-1", Title: "f1", AuthorID: "u-author", ReviewerIDs: []string{"u-r1"}}
		pr2 := &models.PullRequest{ID: "pr-2", Title: "f2", AuthorID: "u-author", ReviewerIDs: []string{"u-r1"}}
		if err := repo.CreatePR(ctx, pr1); err != nil {
			t.Fatalf("CreatePR1: %v", err)
		}
		if err := repo.CreatePR(ctx, pr2); err != nil {
			t.Fatalf("CreatePR2: %v", err)
		}
		list, err := repo.ListPRsByReviewer(ctx, "u-r1", nil)
		if err != nil {
			t.Fatalf("ListPRsByReviewer: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 got %d", len(list))
		}
		mergedAt := time.Now()
		if err := repo.UpdateStatus(ctx, "pr-1", models.PRStatusMERGED, &mergedAt); err != nil {
			t.Fatalf("merge: %v", err)
		}
		status := models.PRStatusMERGED
		filtered, err := repo.ListPRsByReviewer(ctx, "u-r1", &status)
		if err != nil {
			t.Fatalf("ListPRsByReviewer filtered: %v", err)
		}
		if len(filtered) != 1 || filtered[0].ID != "pr-1" {
			t.Fatalf("filter mismatch: %+v", filtered)
		}
	})

	t.Run("ListPRsByReviewer empty result", func(t *testing.T) {
		if err := TruncateAll(ctx, PGC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u-r1", "r1", true)
		list, err := repo.ListPRsByReviewer(ctx, "u-r1", nil)
		if err != nil {
			t.Fatalf("ListPRsByReviewer: %v", err)
		}
		if len(list) != 0 {
			t.Fatalf("expected 0 got %d", len(list))
		}
	})
}
