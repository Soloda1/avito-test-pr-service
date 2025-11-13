package user_repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"avito-test-pr-service/internal/domain/models"
	user_port "avito-test-pr-service/internal/domain/ports/output/user"
	user_repository "avito-test-pr-service/internal/infrastructure/persistence/postgres/user"
	"avito-test-pr-service/internal/utils"
	"avito-test-pr-service/mocks"
)

func newRepo(t *testing.T) (user_port.UserRepository, *mocks.Querier, *mocks.Logger) {
	q := mocks.NewQuerier(t)
	log := mocks.NewLogger(t)

	log.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	log.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	repo := user_repository.NewUserRepository(q, log)
	return repo, q, log
}

func TestUserRepository_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		u := &models.User{ID: uuid.New(), Name: "alice", IsActive: true}

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

		err := repo.CreateUser(ctx, u)
		require.NoError(t, err)
	})

	t.Run("unique violation", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		u := &models.User{ID: uuid.New(), Name: "alice", IsActive: true}

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), &pgconn.PgError{Code: "23505"})

		err := repo.CreateUser(ctx, u)
		assert.ErrorIs(t, err, utils.ErrUserExists)
	})

	t.Run("db error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		u := &models.User{ID: uuid.New(), Name: "alice", IsActive: true}

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))

		err := repo.CreateUser(ctx, u)
		require.Error(t, err)
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		id := uuid.New()
		now := time.Now().Truncate(time.Microsecond)

		row := mocks.NewRow(t)
		row.EXPECT().Scan(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Run(func(args ...interface{}) {
			*(args[0].(*uuid.UUID)) = id
			*(args[1].(*string)) = "bob"
			*(args[2].(*bool)) = true
			*(args[3].(*time.Time)) = now
			*(args[4].(*time.Time)) = now
		}).Return(nil)

		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		u, err := repo.GetUserByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, u.ID)
		assert.Equal(t, "bob", u.Name)
		assert.True(t, u.IsActive)
	})

	t.Run("not found", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		id := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		_, err := repo.GetUserByID(ctx, id)
		assert.ErrorIs(t, err, utils.ErrUserNotFound)
	})

	t.Run("scan error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		id := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error"))
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		_, err := repo.GetUserByID(ctx, id)
		require.Error(t, err)
	})
}

func TestUserRepository_UpdateUserActive(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		id := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("UPDATE 1"), nil)

		err := repo.UpdateUserActive(ctx, id, true)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		id := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("UPDATE 0"), nil)

		err := repo.UpdateUserActive(ctx, id, true)
		assert.ErrorIs(t, err, utils.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		id := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))

		err := repo.UpdateUserActive(ctx, id, true)
		require.Error(t, err)
	})
}

func TestUserRepository_GetTeamIDByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		uid := uuid.New()
		tid := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) {
			*(args[0].(*uuid.UUID)) = tid
		}).Return(nil)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		got, err := repo.GetTeamIDByUserID(ctx, uid)
		require.NoError(t, err)
		assert.Equal(t, tid, got)
	})

	t.Run("no team (no rows)", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		uid := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything).Return(pgx.ErrNoRows)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		_, err := repo.GetTeamIDByUserID(ctx, uid)
		assert.ErrorIs(t, err, utils.ErrUserNoTeam)
	})

	// generic scan error
	t.Run("scan error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		uid := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything).Return(errors.New("scan error"))
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		_, err := repo.GetTeamIDByUserID(ctx, uid)
		require.Error(t, err)
	})
}

func TestUserRepository_ListActiveMembersByTeamID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		id1 := uuid.New()
		id2 := uuid.New()

		rows := mocks.NewRows(t)
		rows.EXPECT().Next().Return(true).Once()
		rows.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) { *(args[0].(*uuid.UUID)) = id1 }).Return(nil).Once()
		rows.EXPECT().Next().Return(true).Once()
		rows.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) { *(args[0].(*uuid.UUID)) = id2 }).Return(nil).Once()
		rows.EXPECT().Next().Return(false).Once()
		rows.EXPECT().Err().Return(nil)
		rows.EXPECT().Close()

		q.EXPECT().Query(ctx, mock.Anything, mock.Anything).Return(rows, nil)

		ids, err := repo.ListActiveMembersByTeamID(ctx, teamID)
		require.NoError(t, err)
		require.Len(t, ids, 2)
		assert.Equal(t, id1, ids[0])
		assert.Equal(t, id2, ids[1])
	})

	t.Run("query error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		teamID := uuid.New()

		q.EXPECT().Query(ctx, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

		_, err := repo.ListActiveMembersByTeamID(ctx, teamID)
		require.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		teamID := uuid.New()

		rows := mocks.NewRows(t)
		rows.EXPECT().Next().Return(true).Once()
		rows.EXPECT().Scan(mock.Anything).Return(errors.New("scan error")).Once()
		rows.EXPECT().Close()

		q.EXPECT().Query(ctx, mock.Anything, mock.Anything).Return(rows, nil)

		_, err := repo.ListActiveMembersByTeamID(ctx, teamID)
		require.Error(t, err)
	})

	t.Run("rows error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		teamID := uuid.New()

		rows := mocks.NewRows(t)
		rows.EXPECT().Next().Return(false).Once()
		rows.EXPECT().Err().Return(errors.New("rows err"))
		rows.EXPECT().Close()

		q.EXPECT().Query(ctx, mock.Anything, mock.Anything).Return(rows, nil)

		_, err := repo.ListActiveMembersByTeamID(ctx, teamID)
		require.Error(t, err)
	})
}

func TestUserRepository_ListUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()
		now := time.Now().Truncate(time.Microsecond)
		id1 := uuid.New()
		id2 := uuid.New()

		rows := mocks.NewRows(t)
		rows.EXPECT().Next().Return(true).Once()
		rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args ...interface{}) {
				*(args[0].(*uuid.UUID)) = id1
				*(args[1].(*string)) = "u1"
				*(args[2].(*bool)) = true
				*(args[3].(*time.Time)) = now
				*(args[4].(*time.Time)) = now
			}).Return(nil).Once()
		rows.EXPECT().Next().Return(true).Once()
		rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args ...interface{}) {
				*(args[0].(*uuid.UUID)) = id2
				*(args[1].(*string)) = "u2"
				*(args[2].(*bool)) = false
				*(args[3].(*time.Time)) = now
				*(args[4].(*time.Time)) = now
			}).Return(nil).Once()
		rows.EXPECT().Next().Return(false).Once()
		rows.EXPECT().Err().Return(nil)
		rows.EXPECT().Close()

		q.EXPECT().Query(ctx, mock.Anything).Return(rows, nil)

		users, err := repo.ListUsers(ctx)
		require.NoError(t, err)
		require.Len(t, users, 2)
		assert.Equal(t, "u1", users[0].Name)
		assert.Equal(t, "u2", users[1].Name)
	})

	t.Run("query error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()

		q.EXPECT().Query(ctx, mock.Anything).Return(nil, errors.New("db error"))

		_, err := repo.ListUsers(ctx)
		require.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()

		rows := mocks.NewRows(t)
		rows.EXPECT().Next().Return(true).Once()
		rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error")).Once()
		rows.EXPECT().Close()

		q.EXPECT().Query(ctx, mock.Anything).Return(rows, nil)

		_, err := repo.ListUsers(ctx)
		require.Error(t, err)
	})

	t.Run("rows error", func(t *testing.T) {
		repo, q, _ := newRepo(t)
		ctx := context.Background()

		rows := mocks.NewRows(t)
		rows.EXPECT().Next().Return(false).Once()
		rows.EXPECT().Err().Return(errors.New("rows err"))
		rows.EXPECT().Close()

		q.EXPECT().Query(ctx, mock.Anything).Return(rows, nil)

		_, err := repo.ListUsers(ctx)
		require.Error(t, err)
	})
}
