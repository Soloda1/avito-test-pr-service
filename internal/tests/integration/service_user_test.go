package integration

import (
	"avito-test-pr-service/internal/application/user"
	"avito-test-pr-service/internal/infrastructure/logger"
	pguow "avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/utils"
	"errors"
	"testing"
)

func newUserService() *user.Service {
	log := logger.New("test")
	u := pguow.NewPostgresUOW(pgC.Pool, log)
	svc := user.NewService(u, log)
	return svc.(*user.Service)
}

func TestUserService_Integration(t *testing.T) {
	ctx := testCtx

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
		if err := InsertUser(ctx, pgC.Pool, "u1", "alice", false); err != nil {
			t.Fatalf("insert user: %v", err)
		}
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
		if err := InsertUser(ctx, pgC.Pool, "u1", "alice", true); err != nil {
			t.Fatalf("insert user: %v", err)
		}
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
		if err := InsertUser(ctx, pgC.Pool, "u1", "a", true); err != nil {
			t.Fatalf("insert u1: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "b", false); err != nil {
			t.Fatalf("insert u2: %v", err)
		}
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
		if err := InsertUser(ctx, pgC.Pool, "u1", "alice", true); err != nil {
			t.Fatalf("insert user: %v", err)
		}
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
		if err := InsertUser(ctx, pgC.Pool, "u1", "alice", true); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		teamID, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("InsertTeam: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("AddTeamMember: %v", err)
		}
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
		if err := InsertUser(ctx, pgC.Pool, "u1", "alice", true); err != nil {
			t.Fatalf("insert: %v", err)
		}
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
		if err := InsertUser(ctx, pgC.Pool, "u1", "alice", true); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		if err := InsertUser(ctx, pgC.Pool, "u2", "bob", false); err != nil {
			t.Fatalf("insert user2: %v", err)
		}
		teamID2, err := InsertTeam(ctx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("InsertTeam: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID2, "u1"); err != nil {
			t.Fatalf("AddTeamMember u1: %v", err)
		}
		if err := AddTeamMember(ctx, pgC.Pool, teamID2, "u2"); err != nil {
			t.Fatalf("AddTeamMember u2: %v", err)
		}
		// В тесте ListMembersByTeamID объявляем svc локально
		svc := newUserService()
		list, err := svc.ListMembersByTeamID(ctx, teamID2.String())
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
