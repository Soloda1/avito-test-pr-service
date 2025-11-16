package integration

import (
	prapp "avito-test-pr-service/internal/application/pr"
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	prrepo "avito-test-pr-service/internal/infrastructure/persistence/postgres/pr"
	pguow "avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/infrastructure/reviewerselector"
	"avito-test-pr-service/internal/utils"
	"errors"
	"fmt"
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

func TestPRService_Integration(t *testing.T) {
	ctx := testCtx

	if err := TruncateAll(ctx, pgC.Pool); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// CreatePR happy 2 candidates
	t.Run("CreatePR happy 2 candidates -> 2 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "r1", true); err != nil {
			t.Fatalf("insert u2: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u3", "r2", true); err != nil {
			t.Fatalf("insert u3: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for _, u := range []string{"u1", "u2", "u3"} {
			if err := AddTeamMember(ctx, pgC.Pool, teamID, u); err != nil {
				t.Fatalf("add member %s: %v", u, err)
			}
		}
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

	// CreatePR one candidate
	t.Run("CreatePR one candidate -> 1 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("u1: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "r1", true); err != nil {
			t.Fatalf("u2: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u3", "r2", false); err != nil {
			t.Fatalf("u3: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for _, u := range []string{"u1", "u2", "u3"} {
			if err := AddTeamMember(ctx, pgC.Pool, teamID, u); err != nil {
				t.Fatalf("add member %s: %v", u, err)
			}
		}
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if len(pr.ReviewerIDs) != 1 {
			t.Fatalf("want 1 got %d", len(pr.ReviewerIDs))
		}
	})

	// CreatePR zero candidates
	t.Run("CreatePR zero candidates -> 0 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("u1: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if len(pr.ReviewerIDs) != 0 {
			t.Fatalf("want 0 got %d", len(pr.ReviewerIDs))
		}
	})

	// Author not exists
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

	// Author no team
	t.Run("CreatePR author no team -> ErrUserNoTeam", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("u1: %v", err)
		}
		_, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err == nil || !errors.Is(err, utils.ErrUserNoTeam) {
			t.Fatalf("want ErrUserNoTeam got %v", err)
		}
	})

	// ReassignReviewer happy
	t.Run("ReassignReviewer happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := prrepo.NewPRRepository(pgC.Pool, logger.New("test"))
		for _, u := range []struct{ id, name string }{{"u1", "author"}, {"u2", "r1"}, {"u3", "r2"}} {
			if err := InsertUser(ctx, pgC.Pool, u.id, u.name, true); err != nil {
				t.Fatalf("user %s: %v", u.id, err)
			}
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for _, u := range []string{"u1", "u2", "u3"} {
			if err := AddTeamMember(ctx, pgC.Pool, teamID, u); err != nil {
				t.Fatalf("member %s: %v", u, err)
			}
		}
		// deactivate candidates
		if err := UpdateUsersActive(ctx, pgC.Pool, []string{"u2", "u3"}, false); err != nil {
			t.Fatalf("deactivate: %v", err)
		}
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if err := UpdateUserActive(ctx, pgC.Pool, "u2", true); err != nil {
			t.Fatalf("activate u2: %v", err)
		}
		if err := repo.AddReviewer(ctx, pr.ID, "u2"); err != nil {
			t.Fatalf("add rev: %v", err)
		}
		if err := UpdateUserActive(ctx, pgC.Pool, "u3", true); err != nil {
			t.Fatalf("activate u3: %v", err)
		}
		upd, err := svc.ReassignReviewer(ctx, pr.ID, "u2")
		if err != nil {
			t.Fatalf("Reassign: %v", err)
		}
		if len(upd.ReviewerIDs) != 1 || upd.ReviewerIDs[0] != "u3" {
			t.Fatalf("expected [u3] got %+v", upd.ReviewerIDs)
		}
	})

	// ReassignReviewer no candidates
	t.Run("ReassignReviewer no candidates -> ErrNoReplacementCandidates", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		for _, u := range []struct{ id, name string }{{"u1", "author"}, {"u2", "r1"}} {
			if err := InsertUser(ctx, pgC.Pool, u.id, u.name, true); err != nil {
				t.Fatalf("user %s: %v", u.id, err)
			}
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for _, u := range []string{"u1", "u2"} {
			if err := AddTeamMember(ctx, pgC.Pool, teamID, u); err != nil {
				t.Fatalf("member %s: %v", u, err)
			}
		}
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		_, err = svc.ReassignReviewer(ctx, pr.ID, "u2")
		if err == nil || !errors.Is(err, utils.ErrNoReplacementCandidates) {
			t.Fatalf("want ErrNoReplacementCandidates got %v", err)
		}
	})

	// ReassignReviewer merged
	t.Run("ReassignReviewer merged -> ErrAlreadyMerged", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := prrepo.NewPRRepository(pgC.Pool, logger.New("test"))
		for _, u := range []struct{ id, name string }{{"u1", "author"}, {"u2", "r1"}} {
			if err := InsertUser(ctx, pgC.Pool, u.id, u.name, true); err != nil {
				t.Fatalf("user %s: %v", u.id, err)
			}
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for _, u := range []string{"u1", "u2"} {
			if err := AddTeamMember(ctx, pgC.Pool, teamID, u); err != nil {
				t.Fatalf("member %s: %v", u, err)
			}
		}
		// Явно создаём PR через сервис
		pr, err := svc.CreatePR(ctx, "pr-1", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
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

	// ReassignReviewer old not assigned
	t.Run("ReassignReviewer old not assigned -> ErrReviewerNotAssigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := prrepo.NewPRRepository(pgC.Pool, logger.New("test"))
		for _, u := range []struct{ id, name string }{{"u1", "author"}, {"u2", "r1"}} {
			if err := InsertUser(ctx, pgC.Pool, u.id, u.name, true); err != nil {
				t.Fatalf("user %s: %v", u.id, err)
			}
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for _, u := range []string{"u1", "u2"} {
			if err := AddTeamMember(ctx, pgC.Pool, teamID, u); err != nil {
				t.Fatalf("member %s: %v", u, err)
			}
		}
		if err := repo.CreatePR(ctx, &models.PullRequest{ID: "pr-1", Title: "title", AuthorID: "u1"}); err != nil {
			t.Fatalf("repo create: %v", err)
		}
		_, err = svc.ReassignReviewer(ctx, "pr-1", "u2")
		if err == nil || !errors.Is(err, utils.ErrReviewerNotAssigned) {
			t.Fatalf("want ErrReviewerNotAssigned got %v", err)
		}
	})

	// ReassignReviewer author no team
	t.Run("ReassignReviewer author no team -> ErrUserNoTeam", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := prrepo.NewPRRepository(pgC.Pool, logger.New("test"))
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("u1: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "r1", true); err != nil {
			t.Fatalf("u2: %v", err)
		}
		if err := repo.CreatePR(ctx, &models.PullRequest{ID: "pr-1", Title: "title", AuthorID: "u1"}); err != nil {
			t.Fatalf("repo create: %v", err)
		}
		if err := repo.AddReviewer(ctx, "pr-1", "u2"); err != nil {
			t.Fatalf("add reviewer: %v", err)
		}
		_, err := svc.ReassignReviewer(ctx, "pr-1", "u2")
		if err == nil || !errors.Is(err, utils.ErrUserNoTeam) {
			t.Fatalf("want ErrUserNoTeam got %v", err)
		}
	})

	// MergePR
	t.Run("MergePR happy and idempotent", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
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
		repo := prrepo.NewPRRepository(pgC.Pool, logger.New("test"))
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
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
			t.Fatalf("idempotent merge error: %v", err)
		}
		if m.Status != models.PRStatusMERGED {
			t.Fatalf("expected MERGED got %v", m.Status)
		}
	})

	// CreatePR duplicate id
	t.Run("CreatePR duplicate id -> ErrPRExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
		if _, err := svc.CreatePR(ctx, "pr-dup", "u1", "t"); err != nil {
			t.Fatalf("first create: %v", err)
		}
		_, err = svc.CreatePR(ctx, "pr-dup", "u1", "t")
		if err == nil || !errors.Is(err, utils.ErrPRExists) {
			t.Fatalf("want ErrPRExists got %v", err)
		}
	})

	// CreatePR author inactive allowed
	t.Run("CreatePR author inactive allowed", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", false); err != nil {
			t.Fatalf("u1: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "r1", true); err != nil {
			t.Fatalf("u2: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u3", "r2", true); err != nil {
			t.Fatalf("u3: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u2"); err != nil {
			t.Fatalf("member: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u3"); err != nil {
			t.Fatalf("member: %v", err)
		}
		pr, err := svc.CreatePR(ctx, "pr-inactive-author", "u1", "t")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if pr.AuthorID != "u1" || len(pr.ReviewerIDs) == 0 {
			t.Fatalf("expected reviewers despite inactive author: %+v", pr)
		}
	})

	// CreatePR empty cases errors
	t.Run("CreatePR empty cases errors", func(t *testing.T) {
		cases := []struct{ name, prID, authorID, title string }{
			{"empty prID", "", "u1", "title"},
			{"empty authorID", "pr-x", "", "title"},
			{"empty title", "pr-x", "u1", ""},
		}
		for _, tc := range cases {
			if err := TruncateAll(ctx, pgC.Pool); err != nil {
				t.Fatalf("truncate: %v", err)
			}
			svc := newPRService()
			_, err := svc.CreatePR(ctx, tc.prID, tc.authorID, tc.title)
			if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
				t.Fatalf("case %s want ErrInvalidArgument got %v", tc.name, err)
			}
		}
	})

	// CreatePR >2 candidates
	t.Run("CreatePR >2 candidates -> exactly 2 assigned", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		for i := 1; i <= 5; i++ {
			if err := InsertUser(ctx, pgC.Pool, fmt.Sprintf("u%d", i), fmt.Sprintf("user%d", i), true); err != nil {
				t.Fatalf("insert u%d: %v", i, err)
			}
			if err := AddTeamMember(ctx, pgC.Pool, teamID, fmt.Sprintf("u%d", i)); err != nil {
				t.Fatalf("add member u%d: %v", i, err)
			}
		}
		pr, err := svc.CreatePR(ctx, "pr-many", "u1", "title")
		if err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		if len(pr.ReviewerIDs) != 2 {
			t.Fatalf("expected 2 reviewers got %d (%+v)", len(pr.ReviewerIDs), pr.ReviewerIDs)
		}
		for _, r := range pr.ReviewerIDs {
			if r == "u1" {
				t.Fatalf("author assigned")
			}
		}
	})

	// ReassignReviewer empty args
	t.Run("ReassignReviewer empty args -> ErrInvalidArgument", func(t *testing.T) {
		cases := []struct{ prID, old string }{{"", "u2"}, {"pr-1", ""}}
		for i, c := range cases {
			if err := TruncateAll(ctx, pgC.Pool); err != nil {
				t.Fatalf("truncate: %v", err)
			}
			svc := newPRService()
			_, err := svc.ReassignReviewer(ctx, c.prID, c.old)
			if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
				t.Fatalf("case %d want ErrInvalidArgument got %v", i, err)
			}
		}
	})

	// MergePR empty id
	t.Run("MergePR empty id -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, err := svc.MergePR(ctx, "")
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	// GetPR happy
	t.Run("GetPR happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
		if _, err := svc.CreatePR(ctx, "pr-get", "u1", "title"); err != nil {
			t.Fatalf("CreatePR: %v", err)
		}
		got, err := svc.GetPR(ctx, "pr-get")
		if err != nil {
			t.Fatalf("GetPR: %v", err)
		}
		if got.ID != "pr-get" || got.AuthorID != "u1" {
			t.Fatalf("mismatch: %+v", got)
		}
	})

	// GetPR invalid id
	t.Run("GetPR invalid id -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, err := svc.GetPR(ctx, "")
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	// GetPR not found
	t.Run("GetPR not found -> ErrPRNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, err := svc.GetPR(ctx, "missing")
		if err == nil || !errors.Is(err, utils.ErrPRNotFound) {
			t.Fatalf("want ErrPRNotFound got %v", err)
		}
	})

	// ListPRsByAssignee OPEN filter
	t.Run("ListPRsByAssignee OPEN filter", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		repo := prrepo.NewPRRepository(pgC.Pool, logger.New("test"))
		if err := InsertUser(ctx, pgC.Pool, "u1", "author", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "rev", true); err != nil {
			t.Fatalf("insert u2: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("team: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("member: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u2"); err != nil {
			t.Fatalf("member: %v", err)
		}
		prA, _ := svc.CreatePR(ctx, "pr-A", "u1", "A")
		prB, _ := svc.CreatePR(ctx, "pr-B", "u1", "B")
		// u2 may already be assigned by CreatePR, check and add only if not present
		for _, pr := range []string{prA.ID, prB.ID} {
			reviewers, err := GetPRReviewers(ctx, pgC.Pool, pr)
			if err != nil {
				t.Fatalf("get reviewers: %v", err)
			}
			hasU2 := false
			for _, r := range reviewers {
				if r == "u2" {
					hasU2 = true
					break
				}
			}
			if !hasU2 {
				if err := AddPRReviewer(ctx, pgC.Pool, pr, "u2"); err != nil {
					t.Fatalf("add reviewer: %v", err)
				}
			}
		}
		now := time.Now()
		_ = repo.UpdateStatus(ctx, prA.ID, models.PRStatusMERGED, &now)
		st := models.PRStatusOPEN
		list, err := svc.ListPRsByAssignee(ctx, "u2", &st)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 1 || list[0].ID != prB.ID {
			t.Fatalf("filter OPEN mismatch: %+v", list)
		}
	})

	// ListPRsByAssignee invalid arg
	t.Run("ListPRsByAssignee invalid arg -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newPRService()
		_, err := svc.ListPRsByAssignee(ctx, "", nil)
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})
}
