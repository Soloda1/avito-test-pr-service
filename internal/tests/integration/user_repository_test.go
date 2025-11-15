package integration

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	userrepo "avito-test-pr-service/internal/infrastructure/persistence/postgres/user"
	"avito-test-pr-service/internal/utils"
	"testing"
)

func TestUserRepository_Integration(t *testing.T) {
	ctx := testCtx
	logger := logger.New("test")
	repo := userrepo.NewUserRepository(pgC.Pool, logger)

	t.Run("Create and GetUserByID", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		u := &models.User{ID: "u1", Name: "alice", IsActive: true}
		if err := repo.CreateUser(ctx, u); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.GetUserByID(ctx, "u1")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != "u1" || got.Name != "alice" || !got.IsActive {
			t.Fatalf("unexpected user: %+v", got)
		}
	})

	t.Run("CreateUser duplicate id -> ErrUserExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		u := &models.User{ID: "u1", Name: "alice", IsActive: true}
		if err := repo.CreateUser(ctx, u); err != nil {
			t.Fatalf("create: %v", err)
		}
		u2 := &models.User{ID: "u1", Name: "bob", IsActive: false}
		err := repo.CreateUser(ctx, u2)
		if err == nil || err != utils.ErrUserExists {
			t.Fatalf("expected ErrUserExists, got %v", err)
		}
	})

	t.Run("UpdateUserActive and UpdateUserName", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		u := &models.User{ID: "u1", Name: "alice", IsActive: true}
		if err := repo.CreateUser(ctx, u); err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := repo.UpdateUserActive(ctx, "u1", false); err != nil {
			t.Fatalf("update active: %v", err)
		}
		if err := repo.UpdateUserName(ctx, "u1", "Alice"); err != nil {
			t.Fatalf("update name: %v", err)
		}
		got, err := repo.GetUserByID(ctx, "u1")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.IsActive != false || got.Name != "Alice" {
			t.Fatalf("unexpected user: %+v", got)
		}
	})

	t.Run("ListUsers returns all", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		if err := repo.CreateUser(ctx, &models.User{ID: "u1", Name: "a", IsActive: true}); err != nil {
			t.Fatalf("create user u1: %v", err)
		}
		if err := repo.CreateUser(ctx, &models.User{ID: "u2", Name: "b", IsActive: false}); err != nil {
			t.Fatalf("create user u2: %v", err)
		}
		list, err := repo.ListUsers(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("want 2, got %d", len(list))
		}
	})

	t.Run("Team relations helpers", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(), 'core', now(), now())`)
		if err != nil {
			t.Fatalf("seed team: %v", err)
		}
		var teamID string
		row := pgC.Pool.QueryRow(ctx, `SELECT id::text FROM teams WHERE name='core' LIMIT 1`)
		if err := row.Scan(&teamID); err != nil {
			t.Fatalf("select team: %v", err)
		}

		if err := repo.CreateUser(ctx, &models.User{ID: "u1", Name: "a", IsActive: true}); err != nil {
			t.Fatalf("create user u1: %v", err)
		}
		if err := repo.CreateUser(ctx, &models.User{ID: "u2", Name: "b", IsActive: false}); err != nil {
			t.Fatalf("create user u2: %v", err)
		}
		_, err = pgC.Pool.Exec(ctx, `INSERT INTO team_members(team_id, user_id) VALUES ($1, $2), ($1, $3)`, teamID, "u1", "u2")
		if err != nil {
			t.Fatalf("seed members: %v", err)
		}

		gid, err := repo.GetTeamIDByUserID(ctx, "u1")
		if err != nil {
			t.Fatalf("GetTeamIDByUserID: %v", err)
		}
		if gid.String() == "" {
			t.Fatalf("empty team id")
		}

		active, err := repo.ListActiveMembersByTeamID(ctx, gid)
		if err != nil {
			t.Fatalf("ListActiveMembersByTeamID: %v", err)
		}
		if len(active) != 1 || active[0] != "u1" {
			t.Fatalf("unexpected active: %+v", active)
		}

		members, err := repo.ListMembersByTeamID(ctx, gid)
		if err != nil {
			t.Fatalf("ListMembersByTeamID: %v", err)
		}
		if len(members) != 2 {
			t.Fatalf("want 2, got %d", len(members))
		}
	})

}
