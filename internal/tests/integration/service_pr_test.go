package integration

import (
	prapp "avito-test-pr-service/internal/application/pr"
	"avito-test-pr-service/internal/domain/models"
	pr_port "avito-test-pr-service/internal/domain/ports/output/pr"
	"avito-test-pr-service/internal/infrastructure/logger"
	prrepo "avito-test-pr-service/internal/infrastructure/persistence/postgres/pr"
	pguow "avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/infrastructure/reviewerselector"
	"avito-test-pr-service/internal/utils"
	"errors"
	"testing"
	"time"

	rand "math/rand/v2"
)

func newPRService() *prapp.Service {
	log := logger.New("test")
	u := pguow.NewPostgresUOW(pgC.Pool, log)

	seed := uint64(1)
	r := rand.New(rand.NewPCG(seed, seed<<1|1))
	selector := reviewerselector.NewRandomReviewerSelectorWithRand(r)

	svc := prapp.NewService(u, selector, log)
	return svc.(*prapp.Service)
}

func insertPRrepo(t *testing.T) pr_port.PRRepository {
	t.Helper()
	log := logger.New("test")
	return prrepo.NewPRRepository(pgC.Pool, log)
}

func TestPRService_Integration(t *testing.T) {
	ctx := testCtx

	insertUser := func(t *testing.T, id, name string, active bool) {
		t.Helper()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	insertTeam := func(t *testing.T, name string) string {
		t.Helper()
		row := pgC.Pool.QueryRow(ctx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(),$1,now(),now()) RETURNING id`, name)
		var id string
		if err := row.Scan(&id); err != nil {
			t.Fatalf("insert team: %v", err)
		}
		return id
	}
	addMember := func(t *testing.T, teamID string, userID string) {
		t.Helper()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2)`, teamID, userID)
		if err != nil {
			t.Fatalf("add member: %v", err)
		}
	}
	setActive := func(t *testing.T, id string, active bool) {
		t.Helper()
		_, err := pgC.Pool.Exec(ctx, `UPDATE users SET is_active=$2, updated_at=now() WHERE id=$1`, id, active)
		if err != nil {
			t.Fatalf("set active: %v", err)
		}
	}

	t.Run("CreatePR happy 2 candidates -> 2 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		insertUser(t, "u2", "r1", true)
		insertUser(t, "u3", "r2", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		addMember(t, teamID, "u2")
		addMember(t, teamID, "u3")
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if pr.ID != "pr-1" || pr.AuthorID != "u1" || pr.Title != "title" {
			t.Fatalf("meta mismatch: %+v", pr)
		}
		if len(pr.ReviewerIDs) != 2 {
			t.Fatalf("want 2 got %d", len(pr.ReviewerIDs))
		}
		for _, id := range pr.ReviewerIDs {
			if id == "u1" {
				t.Fatalf("author assigned")
			}
		}
	})

	t.Run("CreatePR one candidate -> 1 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		insertUser(t, "u2", "r1", true)
		insertUser(t, "u3", "r2", false) // inactive
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		addMember(t, teamID, "u2")
		addMember(t, teamID, "u3")
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if len(pr.ReviewerIDs) != 1 {
			t.Fatalf("want 1 got %d", len(pr.ReviewerIDs))
		}
	})

	t.Run("CreatePR zero candidates -> 0 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if len(pr.ReviewerIDs) != 0 {
			t.Fatalf("want 0 got %d", len(pr.ReviewerIDs))
		}
	})

	t.Run("CreatePR author not exists -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, err := svc.CreatePR(ctx, "pr-1", "missing", "title")
		if err == nil || !errors.Is(err, utils.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound got %v", err)
		}
	})

	t.Run("CreatePR author no team -> ErrUserNoTeam", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		_, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err == nil || !errors.Is(err, utils.ErrUserNoTeam) {
			t.Fatalf("want ErrUserNoTeam got %v", err)
		}
	})

	t.Run("ReassignReviewer happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := insertPRrepo(t)
		insertUser(t, "u1", "author", true)
		insertUser(t, "u2", "r1", true)
		insertUser(t, "u3", "r2", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		addMember(t, teamID, "u2")
		addMember(t, teamID, "u3")
		// создаём PR без кандидатов (сначала сделаем обоих неактивными)
		setActive(t, "u2", false)
		setActive(t, "u3", false)
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		// назначим старого ревьюера
		setActive(t, "u2", true)
		if err := repo.AddReviewer(ctx, pr.ID, "u2"); err != nil {
			t.Fatalf("add rev: %v", err)
		}
		setActive(t, "u3", true) // будет единственным кандидатом
		upd, err := svc.ReassignReviewer(ctx, pr.ID, "u2")
		if err != nil {
			t.Fatalf("Reassign: %v", err)
		}
		if len(upd.ReviewerIDs) != 1 || upd.ReviewerIDs[0] != "u3" {
			t.Fatalf("expected [u3], got %+v", upd.ReviewerIDs)
		}
	})

	t.Run("ReassignReviewer no candidates -> ErrNoReplacementCandidates", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := insertPRrepo(t)
		insertUser(t, "u1", "author", true)
		insertUser(t, "u2", "r1", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		addMember(t, teamID, "u2")
		setActive(t, "u2", false)
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if err := repo.AddReviewer(ctx, pr.ID, "u2"); err != nil {
			t.Fatalf("add rev: %v", err)
		}
		_, err = svc.ReassignReviewer(ctx, pr.ID, "u2")
		if err == nil || !errors.Is(err, utils.ErrNoReplacementCandidates) {
			t.Fatalf("want ErrNoReplacementCandidates got %v", err)
		}
	})

	t.Run("ReassignReviewer merged -> ErrAlreadyMerged", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := insertPRrepo(t)
		insertUser(t, "u1", "author", true)
		insertUser(t, "u2", "r1", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		addMember(t, teamID, "u2")
		setActive(t, "u2", false)
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if err := repo.AddReviewer(ctx, pr.ID, "u2"); err != nil {
			t.Fatalf("add rev: %v", err)
		}
		now := time.Now()
		if err := repo.UpdateStatus(ctx, pr.ID, models.PRStatusMERGED, &now); err != nil {
			t.Fatalf("merge: %v", err)
		}
		_, err = svc.ReassignReviewer(ctx, pr.ID, "u2")
		if err == nil || !errors.Is(err, utils.ErrAlreadyMerged) {
			t.Fatalf("want ErrAlreadyMerged got %v", err)
		}
	})

	t.Run("ReassignReviewer old not assigned -> ErrReviewerNotAssigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		_, err = svc.ReassignReviewer(ctx, pr.ID, "u2")
		if err == nil || !errors.Is(err, utils.ErrReviewerNotAssigned) {
			t.Fatalf("want ErrReviewerNotAssigned got %v", err)
		}
	})

	t.Run("ReassignReviewer author no team -> ErrUserNoTeam", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		repo := insertPRrepo(t)
		if err := repo.CreatePR(ctx, &models.PullRequest{ID: "pr-1", Title: "title", AuthorID: "u1"}); err != nil {
			t.Fatalf("repo create: %v", err)
		}
		insertUser(t, "u2", "r1", true)
		if err := repo.AddReviewer(ctx, "pr-1", "u2"); err != nil {
			t.Fatalf("add reviewer: %v", err)
		}
		_, err := svc.ReassignReviewer(ctx, "pr-1", "u2")
		if err == nil || !errors.Is(err, utils.ErrUserNoTeam) {
			t.Fatalf("want ErrUserNoTeam got %v", err)
		}
	})

	t.Run("MergePR happy and idempotent", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		insertUser(t, "u1", "author", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		m1, err := svc.MergePR(ctx, pr.ID)
		if err != nil {
			t.Fatalf("merge1: %v", err)
		}
		if m1.Status != models.PRStatusMERGED || m1.MergedAt == nil {
			t.Fatalf("not merged: %+v", m1)
		}
		m2, err := svc.MergePR(ctx, pr.ID)
		if err != nil {
			t.Fatalf("merge2: %v", err)
		}
		if m2.Status != models.PRStatusMERGED {
			t.Fatalf("idempotent failed: %+v", m2)
		}
	})

	t.Run("MergePR not found -> ErrPRNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, err := svc.MergePR(ctx, "missing")
		if err == nil || !errors.Is(err, utils.ErrPRNotFound) {
			t.Fatalf("want ErrPRNotFound got %v", err)
		}
	})

	t.Run("MergePR already merged -> idempotent no error", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := insertPRrepo(t)
		insertUser(t, "u1", "author", true)
		teamID := insertTeam(t, "core")
		addMember(t, teamID, "u1")
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		now := time.Now()
		if err := repo.UpdateStatus(ctx, pr.ID, models.PRStatusMERGED, &now); err != nil {
			t.Fatalf("merge: %v", err)
		}
		m, err := svc.MergePR(ctx, pr.ID)
		if err != nil {
			t.Fatalf("idempotent merge returned error: %v", err)
		}
		if m.Status != models.PRStatusMERGED {
			t.Fatalf("expected MERGED, got %v", m.Status)
		}
	})

	t.Run("CreatePR duplicate id -> ErrPRExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, _ = pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u1", "author", true)
		teamID := func() string {
			row := pgC.Pool.QueryRow(ctx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(),$1,now(),now()) RETURNING id`, "core")
			var id string
			if err := row.Scan(&id); err != nil {
				t.Fatalf("team: %v", err)
			}
			return id
		}()
		_, _ = pgC.Pool.Exec(ctx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2)`, teamID, "u1")
		if _, err := svc.CreatePR(ctx, "pr-dup", "u1", "t"); err != nil {
			t.Fatalf("first create: %v", err)
		}
		_, err := svc.CreatePR(ctx, "pr-dup", "u1", "t")
		if err == nil || !errors.Is(err, utils.ErrPRExists) {
			t.Fatalf("want ErrPRExists got %v", err)
		}
	})

	t.Run("CreatePR author inactive allowed", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, _ = pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u1", "author", false)
		_, _ = pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u2", "r1", true)
		_, _ = pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u3", "r2", true)
		teamID := func() string {
			row := pgC.Pool.QueryRow(ctx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(),$1,now(),now()) RETURNING id`, "core")
			var id string
			if err := row.Scan(&id); err != nil {
				t.Fatalf("team: %v", err)
			}
			return id
		}()
		_, _ = pgC.Pool.Exec(ctx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2),($1,$3),($1,$4)`, teamID, "u1", "u2", "u3")
		pr, err := svc.CreatePR(ctx, "pr-inactive-author", "u1", "t")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if pr.AuthorID != "u1" || len(pr.ReviewerIDs) == 0 {
			t.Fatalf("expected reviewers despite inactive author: %+v", pr)
		}
	})

}
