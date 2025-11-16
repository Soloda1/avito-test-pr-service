package integration

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	teamrepo "avito-test-pr-service/internal/infrastructure/persistence/postgres/team"
	"avito-test-pr-service/internal/utils"
	"testing"

	"github.com/google/uuid"
)

func TestTeamRepository_Integration(t *testing.T) {
	ctx := testCtx
	log := logger.New("test")
	repo := teamrepo.NewTeamRepository(pgC.Pool, log)

	t.Run("CreateTeam success", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("CreateTeam: %v", err)
		}
		if team.ID == uuid.Nil {
			t.Fatalf("expected non-nil team ID")
		}
		if team.Name != "core" {
			t.Fatalf("unexpected name: %s", team.Name)
		}
		if team.CreatedAt.IsZero() || team.UpdatedAt.IsZero() {
			t.Fatalf("timestamps not set")
		}
	})

	t.Run("CreateTeam duplicate name -> ErrAlreadyExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team1 := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team1); err != nil {
			t.Fatalf("create first: %v", err)
		}
		team2 := &models.Team{Name: "core"}
		err := repo.CreateTeam(ctx, team2)
		if err == nil || err != utils.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("CreateTeam invalid arg (empty name)", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: ""}
		err := repo.CreateTeam(ctx, team)
		if err == nil || err != utils.ErrInvalidArgument {
			t.Fatalf("expected ErrInvalidArgument, got %v", err)
		}
	})

	t.Run("GetTeamByID success", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "backend"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.GetTeamByID(ctx, team.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != team.ID || got.Name != team.Name {
			t.Fatalf("mismatch: %+v vs %+v", got, team)
		}
	})

	t.Run("GetTeamByID not found -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		_, err := repo.GetTeamByID(ctx, uuid.New())
		if err == nil || err != utils.ErrTeamNotFound {
			t.Fatalf("expected ErrTeamNotFound, got %v", err)
		}
	})

	t.Run("GetTeamByName success", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "frontend"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.GetTeamByName(ctx, "frontend")
		if err != nil {
			t.Fatalf("get by name: %v", err)
		}
		if got.Name != "frontend" {
			t.Fatalf("unexpected name: %s", got.Name)
		}
	})

	t.Run("GetTeamByName not found -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		_, err := repo.GetTeamByName(ctx, "absent")
		if err == nil || err != utils.ErrTeamNotFound {
			t.Fatalf("expected ErrTeamNotFound, got %v", err)
		}
	})

	t.Run("ListTeams returns all", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		for _, n := range []string{"a", "b", "c"} {
			team := &models.Team{Name: n}
			if err := repo.CreateTeam(ctx, team); err != nil {
				t.Fatalf("create %s: %v", n, err)
			}
		}
		teams, err := repo.ListTeams(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(teams) != 3 {
			t.Fatalf("want 3, got %d", len(teams))
		}
	})

	t.Run("AddMember success", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create team: %v", err)
		}
		insertUser := func(id, name string) error { return InsertUser(ctx, pgC.Pool, id, name, true) }
		if err := insertUser("u1", "alice"); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		if err := repo.AddMember(ctx, team.ID, "u1"); err != nil {
			t.Fatalf("AddMember: %v", err)
		}
		members, err := GetTeamMemberIDs(ctx, pgC.Pool, team.ID)
		if err != nil {
			t.Fatalf("GetTeamMemberIDs: %v", err)
		}
		if len(members) != 1 || members[0] != "u1" {
			t.Fatalf("expected single member u1 got %v", members)
		}
	})

	t.Run("AddMember duplicate -> ErrAlreadyExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create team: %v", err)
		}
		insertUser := func(id, name string) error { return InsertUser(ctx, pgC.Pool, id, name, true) }
		if err := insertUser("u1", "alice"); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		if err := repo.AddMember(ctx, team.ID, "u1"); err != nil {
			t.Fatalf("first add: %v", err)
		}
		err := repo.AddMember(ctx, team.ID, "u1")
		if err == nil || err != utils.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists got %v", err)
		}
	})

	t.Run("AddMember user FK violation -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create team: %v", err)
		}
		err := repo.AddMember(ctx, team.ID, "missing-user")
		if err == nil || err != utils.ErrUserNotFound {
			t.Fatalf("expected ErrUserNotFound got %v", err)
		}
	})

	t.Run("AddMember team FK violation -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		insertUser := func(id, name string) error { return InsertUser(ctx, pgC.Pool, id, name, true) }
		if err := insertUser("u1", "alice"); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		err := repo.AddMember(ctx, uuid.New(), "u1")
		if err == nil || err != utils.ErrTeamNotFound {
			t.Fatalf("expected ErrTeamNotFound got %v", err)
		}
	})

	t.Run("RemoveMember success", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create team: %v", err)
		}
		insertUser := func(id, name string) error { return InsertUser(ctx, pgC.Pool, id, name, true) }
		if err := insertUser("u1", "alice"); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		if err := repo.AddMember(ctx, team.ID, "u1"); err != nil {
			t.Fatalf("add member: %v", err)
		}
		if err := repo.RemoveMember(ctx, team.ID, "u1"); err != nil {
			t.Fatalf("remove: %v", err)
		}
		members, err := GetTeamMemberIDs(ctx, pgC.Pool, team.ID)
		if err != nil {
			t.Fatalf("GetTeamMemberIDs: %v", err)
		}
		if len(members) != 0 {
			t.Fatalf("expected 0 members got %v", members)
		}
	})

	t.Run("RemoveMember not found -> ErrNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate failed: %v", err)
		}
		team := &models.Team{Name: "core"}
		if err := repo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("create team: %v", err)
		}
		insertUser := func(id, name string) error { return InsertUser(ctx, pgC.Pool, id, name, true) }
		if err := insertUser("u1", "alice"); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		err := repo.RemoveMember(ctx, team.ID, "u1")
		if err == nil || err != utils.ErrNotFound {
			t.Fatalf("expected ErrNotFound got %v", err)
		}
	})
}
