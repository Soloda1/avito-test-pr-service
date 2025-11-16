package integration

import (
	"avito-test-pr-service/internal/application/user"
	"avito-test-pr-service/internal/infrastructure/logger"
	pguow "avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func newUserService() *user.Service {
	log := logger.New("test")
	u := pguow.NewPostgresUOW(pgC.Pool, log)
	svc := user.NewService(u, log)
	return svc.(*user.Service)
}

func insertTeam(t *testing.T, ctx context.Context, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pgC.Pool.Exec(ctx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES ($1,$2,now(),now())`, id, name)
	if err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func addMemberRel(t *testing.T, ctx context.Context, teamID uuid.UUID, userID string) {
	t.Helper()
	_, err := pgC.Pool.Exec(ctx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2)`, teamID, userID)
	if err != nil {
		t.Fatalf("insert team_member: %v", err)
	}
}

func TestUserService_Integration(t *testing.T) {
	ctx := testCtx
	insertUser := func(t *testing.T, id, name string, active bool) {
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}

	t.Run("CreateUser happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		u, err := svc.CreateUser(ctx, "u1", "alice", true)
		if err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		if u.ID != "u1" || u.Name != "alice" || !u.IsActive {
			t.Fatalf("mismatch: %+v", u)
		}
	})

	t.Run("CreateUser duplicate -> ErrUserExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		_, err := svc.CreateUser(ctx, "u1", "alice", true)
		if err != nil {
			t.Fatalf("first create: %v", err)
		}
		_, err = svc.CreateUser(ctx, "u1", "alice", true)
		if err == nil || !errors.Is(err, utils.ErrUserExists) {
			t.Fatalf("expected ErrUserExists got %v", err)
		}
	})

	t.Run("CreateUser invalid -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		_, err := svc.CreateUser(ctx, "", "alice", true)
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("UpdateUserActive happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "alice", false)
		svc := newUserService()
		if err := svc.UpdateUserActive(ctx, "u1", true); err != nil {
			t.Fatalf("UpdateUserActive: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT is_active FROM users WHERE id=$1`, "u1")
		var active bool
		if err := row.Scan(&active); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if !active {
			t.Fatalf("expected active=true")
		}
	})

	t.Run("UpdateUserActive not found -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		err := svc.UpdateUserActive(ctx, "missing", true)
		if err == nil || !errors.Is(err, utils.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound got %v", err)
		}
	})

	t.Run("GetUser happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "alice", true)
		svc := newUserService()
		u, err := svc.GetUser(ctx, "u1")
		if err != nil {
			t.Fatalf("GetUser: %v", err)
		}
		if u.Name != "alice" {
			t.Fatalf("unexpected name: %s", u.Name)
		}
	})

	t.Run("GetUser not found -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		_, err := svc.GetUser(ctx, "missing")
		if err == nil || !errors.Is(err, utils.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound got %v", err)
		}
	})

	t.Run("ListUsers returns all", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "a", true)
		insertUser(t, "u2", "b", false)
		svc := newUserService()
		list, err := svc.ListUsers(ctx)
		if err != nil {
			t.Fatalf("ListUsers: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("want 2 got %d", len(list))
		}
	})

	t.Run("GetUserTeamName no team -> empty name, nil err", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "alice", true)
		svc := newUserService()
		name, err := svc.GetUserTeamName(ctx, "u1")
		if err != nil {
			t.Fatalf("GetUserTeamName: %v", err)
		}
		if name != "" {
			t.Fatalf("expected empty team name, got %q", name)
		}
	})

	t.Run("GetUserTeamName with team -> team name", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "alice", true)
		teamID := insertTeam(t, ctx, "core")
		addMemberRel(t, ctx, teamID, "u1")
		svc := newUserService()
		name, err := svc.GetUserTeamName(ctx, "u1")
		if err != nil {
			t.Fatalf("GetUserTeamName: %v", err)
		}
		if name != "core" {
			t.Fatalf("want core got %s", name)
		}
	})

	t.Run("UpdateUserName happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "alice", true)
		svc := newUserService()
		if err := svc.UpdateUserName(ctx, "u1", "newname"); err != nil {
			t.Fatalf("UpdateUserName: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT name FROM users WHERE id=$1`, "u1")
		var name string
		if err := row.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if name != "newname" {
			t.Fatalf("want newname got %s", name)
		}
	})

	t.Run("UpdateUserName not found -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		err := svc.UpdateUserName(ctx, "missing", "name")
		if err == nil || !errors.Is(err, utils.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound got %v", err)
		}
	})

	t.Run("ListMembersByTeamID happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUser(t, "u1", "alice", true)
		insertUser(t, "u2", "bob", false)
		teamID := insertTeam(t, ctx, "core")
		addMemberRel(t, ctx, teamID, "u1")
		addMemberRel(t, ctx, teamID, "u2")
		svc := newUserService()
		list, err := svc.ListMembersByTeamID(ctx, teamID.String())
		if err != nil {
			t.Fatalf("ListMembersByTeamID: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("want 2 got %d", len(list))
		}
	})

	t.Run("ListMembersByTeamID invalid id -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newUserService()
		_, err := svc.ListMembersByTeamID(ctx, "not-a-uuid")
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})
}
