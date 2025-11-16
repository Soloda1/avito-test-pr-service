package integration

import (
	teamapp "avito-test-pr-service/internal/application/team"
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	pguow "avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/utils"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func newTeamService() *teamapp.Service {
	log := logger.New("test")
	u := pguow.NewPostgresUOW(pgC.Pool, log)
	svc := teamapp.NewService(u, log)
	return svc.(*teamapp.Service)
}

func TestTeamService_Integration(t *testing.T) {
	ctx := testCtx

	insertUser := func(t *testing.T, id, name string, active bool) {
		t.Helper()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}

	t.Run("CreateTeam happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("CreateTeam: %v", err)
		}
		if team.ID == uuid.Nil || team.Name != "core" {
			t.Fatalf("unexpected team: %+v", team)
		}
	})

	t.Run("CreateTeam invalid -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := svc.CreateTeam(ctx, "")
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("AddMember happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		uid := uuid.New()
		insertUser(t, uid.String(), "alice", true)
		if err := svc.AddMember(ctx, team.ID, uid); err != nil {
			t.Fatalf("AddMember: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM team_members WHERE team_id=$1 AND user_id=$2`, team.ID, uid.String())
		var cnt int
		if err := row.Scan(&cnt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if cnt != 1 {
			t.Fatalf("expected 1 got %d", cnt)
		}
	})

	t.Run("AddMember duplicate -> ErrAlreadyExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		uid := uuid.New()
		insertUser(t, uid.String(), "alice", true)
		if err := svc.AddMember(ctx, team.ID, uid); err != nil {
			t.Fatalf("first add: %v", err)
		}
		err = svc.AddMember(ctx, team.ID, uid)
		if err == nil || !errors.Is(err, utils.ErrAlreadyExists) {
			t.Fatalf("want ErrAlreadyExists got %v", err)
		}
	})

	t.Run("AddMember user not found -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		err = svc.AddMember(ctx, team.ID, uuid.New())
		if err == nil || !errors.Is(err, utils.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound got %v", err)
		}
	})

	t.Run("AddMember team not found -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		uid := uuid.New()
		insertUser(t, uid.String(), "alice", true)
		err := svc.AddMember(ctx, uuid.New(), uid)
		if err == nil || !errors.Is(err, utils.ErrTeamNotFound) {
			t.Fatalf("want ErrTeamNotFound got %v", err)
		}
	})

	t.Run("RemoveMember happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		uid := uuid.New()
		insertUser(t, uid.String(), "alice", true)
		if err := svc.AddMember(ctx, team.ID, uid); err != nil {
			t.Fatalf("add member: %v", err)
		}
		if err := svc.RemoveMember(ctx, team.ID, uid); err != nil {
			t.Fatalf("remove: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM team_members WHERE team_id=$1 AND user_id=$2`, team.ID, uid.String())
		var cnt int
		if err := row.Scan(&cnt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if cnt != 0 {
			t.Fatalf("expected 0 got %d", cnt)
		}
	})

	t.Run("RemoveMember not found -> ErrNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		uid := uuid.New()
		insertUser(t, uid.String(), "alice", true)
		err = svc.RemoveMember(ctx, team.ID, uid)
		if err == nil || !errors.Is(err, utils.ErrNotFound) {
			t.Fatalf("want ErrNotFound got %v", err)
		}
	})

	t.Run("GetTeam happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := svc.GetTeam(ctx, team.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != team.ID || got.Name != team.Name {
			t.Fatalf("mismatch: %+v vs %+v", got, team)
		}
	})

	t.Run("GetTeam not found -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := svc.GetTeam(ctx, uuid.New())
		if err == nil || !errors.Is(err, utils.ErrTeamNotFound) {
			t.Fatalf("want ErrTeamNotFound got %v", err)
		}
	})

	t.Run("ListTeams returns all", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		for _, n := range []string{"a", "b", "c"} {
			if _, err := svc.CreateTeam(ctx, n); err != nil {
				t.Fatalf("create %s: %v", n, err)
			}
		}
		list, err := svc.ListTeams(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 3 {
			t.Fatalf("want 3 got %d", len(list))
		}
	})

	t.Run("CreateTeamWithMembers mixed (new + existing) and idempotent add", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		insertUser(t, "u-bob", "Bob", false)
		members := []*models.User{
			{ID: "u-alice", Name: "Alice", IsActive: true},
			{ID: "u-bob", Name: "BobNew", IsActive: true},  // обновим существующего
			{ID: "u-alice", Name: "Alice", IsActive: true}, // повторная попытка — должна быть идемпотентной на связке
		}
		team, users, err := svc.CreateTeamWithMembers(ctx, "backend", members)
		if err != nil {
			t.Fatalf("CreateTeamWithMembers: %v", err)
		}
		if team.Name != "backend" {
			t.Fatalf("unexpected team name: %s", team.Name)
		}
		if len(users) != len(members) {
			t.Fatalf("want %d users in result got %d", len(members), len(users))
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM team_members WHERE team_id=$1`, team.ID)
		var cnt int
		if err := row.Scan(&cnt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if cnt != 2 {
			t.Fatalf("want 2 unique members got %d", cnt)
		}
		row2 := pgC.Pool.QueryRow(ctx, `SELECT is_active, name FROM users WHERE id=$1`, "u-bob")
		var active bool
		var name string
		if err := row2.Scan(&active, &name); err != nil {
			t.Fatalf("scan2: %v", err)
		}
		if !active || name != "BobNew" {
			t.Fatalf("bob not updated: active=%v name=%s", active, name)
		}
	})

	t.Run("CreateTeamWithMembers invalid name -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, _, err := svc.CreateTeamWithMembers(ctx, "", []*models.User{{ID: "u1", Name: "x"}})
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("CreateTeamWithMembers invalid member (nil)", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, _, err := svc.CreateTeamWithMembers(ctx, "core", []*models.User{nil})
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("CreateTeamWithMembers invalid member (empty id)", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, _, err := svc.CreateTeamWithMembers(ctx, "core", []*models.User{{ID: ""}})
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("GetTeamByName happy", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := svc.GetTeamByName(ctx, "core")
		if err != nil {
			t.Fatalf("get by name: %v", err)
		}
		if got.ID != team.ID || got.Name != team.Name {
			t.Fatalf("mismatch: %+v vs %+v", got, team)
		}
	})

	t.Run("GetTeamByName not found -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := svc.GetTeamByName(ctx, "absent")
		if err == nil || !errors.Is(err, utils.ErrTeamNotFound) {
			t.Fatalf("want ErrTeamNotFound got %v", err)
		}
	})

	t.Run("CreateTeam duplicate name -> ErrTeamExists", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("first create: %v", err)
		}
		_, err = svc.CreateTeam(ctx, "core")
		if err == nil || !errors.Is(err, utils.ErrTeamExists) {
			if err == nil || (!errors.Is(err, utils.ErrAlreadyExists) && !errors.Is(err, utils.ErrTeamExists)) {
				t.Fatalf("want ErrTeamExists/ErrAlreadyExists got %v", err)
			}
		}
	})

	t.Run("AddMember invalid args -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		err := svc.AddMember(ctx, uuid.Nil, uuid.Nil)
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("RemoveMember invalid args -> ErrInvalidArgument", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		err := svc.RemoveMember(ctx, uuid.Nil, uuid.Nil)
		if err == nil || !errors.Is(err, utils.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument got %v", err)
		}
	})

	t.Run("RemoveMember team not found -> ErrTeamNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		uid := uuid.New()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, uid.String(), "alice", true)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
		err = svc.RemoveMember(ctx, uuid.New(), uid)
		if err == nil || !errors.Is(err, utils.ErrTeamNotFound) {
			if err == nil || (!errors.Is(err, utils.ErrNotFound) && !errors.Is(err, utils.ErrTeamNotFound)) {
				t.Fatalf("want ErrTeamNotFound/ErrNotFound got %v", err)
			}
		}
	})

	t.Run("RemoveMember user not found -> ErrUserNotFound", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		team, err := svc.CreateTeam(ctx, "core")
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		err = svc.RemoveMember(ctx, team.ID, uuid.New())
		if err == nil || !errors.Is(err, utils.ErrUserNotFound) {
			if err == nil || (!errors.Is(err, utils.ErrNotFound) && !errors.Is(err, utils.ErrUserNotFound)) {
				t.Fatalf("want ErrUserNotFound/ErrNotFound got %v", err)
			}
		}
	})

	t.Run("CreateTeamWithMembers update existing user active only", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u1", "Alice", false)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
		members := []*models.User{{ID: "u1", Name: "Alice", IsActive: true}}
		_, users, err := svc.CreateTeamWithMembers(ctx, "backend", members)
		if err != nil {
			t.Fatalf("CreateTeamWithMembers: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT is_active FROM users WHERE id=$1`, "u1")
		var active bool
		if err := row.Scan(&active); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if !active {
			t.Fatalf("expected active updated to true")
		}
		if users[0].IsActive != true {
			t.Fatalf("returned user not updated")
		}
	})

	t.Run("CreateTeamWithMembers update existing user name only", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u1", "Old", true)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
		members := []*models.User{{ID: "u1", Name: "New", IsActive: true}}
		_, users, err := svc.CreateTeamWithMembers(ctx, "backend", members)
		if err != nil {
			t.Fatalf("CreateTeamWithMembers: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT name FROM users WHERE id=$1`, "u1")
		var name string
		if err := row.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if name != "New" {
			t.Fatalf("expected name updated to New got %s", name)
		}
		if users[0].Name != "New" {
			t.Fatalf("returned user name not updated")
		}
	})

	t.Run("CreateTeamWithMembers update existing user name and active", func(t *testing.T) {
		if err := TruncateAll(ctx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		svc := newTeamService()
		_, err := pgC.Pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, "u1", "Old", false)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
		members := []*models.User{{ID: "u1", Name: "New", IsActive: true}}
		_, users, err := svc.CreateTeamWithMembers(ctx, "backend", members)
		if err != nil {
			t.Fatalf("CreateTeamWithMembers: %v", err)
		}
		row := pgC.Pool.QueryRow(ctx, `SELECT name, is_active FROM users WHERE id=$1`, "u1")
		var name string
		var active bool
		if err := row.Scan(&name, &active); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if name != "New" || !active {
			t.Fatalf("expected updated both fields got name=%s active=%v", name, active)
		}
		if users[0].Name != "New" || users[0].IsActive != true {
			t.Fatalf("returned user not updated fully")
		}
	})

}
