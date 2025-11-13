package team_repository_test

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
	team_port "avito-test-pr-service/internal/domain/ports/output/team"
	"avito-test-pr-service/internal/infrastructure/logger"
	repository "avito-test-pr-service/internal/infrastructure/persistence/postgres/team"
	"avito-test-pr-service/internal/utils"
	"avito-test-pr-service/mocks"
)

func newTeamRepo(t *testing.T) (team_port.TeamRepository, *mocks.Querier) {
	q := mocks.NewQuerier(t)
	log := logger.New("dev")
	repo := repository.NewTeamRepository(q, log)
	return repo, q
}

func TestTeamRepository_CreateTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		team := &models.Team{Name: "devs"}

		row := mocks.NewRow(t)
		generatedID := uuid.New()
		row.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) {
			*(args[0].(*uuid.UUID)) = generatedID
		}).Return(nil)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		require.NoError(t, repo.CreateTeam(ctx, team))
		assert.Equal(t, generatedID, team.ID)
	})

	t.Run("already exists", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		team := &models.Team{Name: "devs"}

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything).Return(pgx.ErrNoRows)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		err := repo.CreateTeam(ctx, team)
		assert.ErrorIs(t, err, utils.ErrAlreadyExists)
	})

	t.Run("scan error", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		team := &models.Team{Name: "devs"}

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything).Return(errors.New("scan error"))
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		require.Error(t, repo.CreateTeam(ctx, team))
	})

	t.Run("invalid arg", func(t *testing.T) {
		repo, _ := newTeamRepo(t)
		ctx := context.Background()
		err := repo.CreateTeam(ctx, &models.Team{Name: ""})
		assert.ErrorIs(t, err, utils.ErrInvalidArgument)
	})
}

func TestTeamRepository_GetTeamByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		id := uuid.New()
		now := time.Now().Truncate(time.Microsecond)

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args ...interface{}) {
				*(args[0].(*uuid.UUID)) = id
				*(args[1].(*string)) = "devs"
				*(args[2].(*time.Time)) = now
				*(args[3].(*time.Time)) = now
			}).Return(nil)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		team, err := repo.GetTeamByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, team.ID)
		assert.Equal(t, "devs", team.Name)
	})

	t.Run("not found", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		id := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		_, err := repo.GetTeamByID(ctx, id)
		assert.ErrorIs(t, err, utils.ErrTeamNotFound)
	})

	t.Run("scan error", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		id := uuid.New()

		row := mocks.NewRow(t)
		row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error"))
		q.EXPECT().QueryRow(ctx, mock.Anything, mock.Anything).Return(row)

		_, err := repo.GetTeamByID(ctx, id)
		require.Error(t, err)
	})
}

func TestTeamRepository_ListTeams(t *testing.T) {
	repo, q := newTeamRepo(t)
	ctx := context.Background()
	now := time.Now().Truncate(time.Microsecond)
	id1 := uuid.New()
	id2 := uuid.New()

	rows := mocks.NewRows(t)
	rows.EXPECT().Next().Return(true).Once()
	rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args ...interface{}) {
			*(args[0].(*uuid.UUID)) = id1
			*(args[1].(*string)) = "t1"
			*(args[2].(*time.Time)) = now
			*(args[3].(*time.Time)) = now
		}).Return(nil).Once()
	rows.EXPECT().Next().Return(true).Once()
	rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args ...interface{}) {
			*(args[0].(*uuid.UUID)) = id2
			*(args[1].(*string)) = "t2"
			*(args[2].(*time.Time)) = now
			*(args[3].(*time.Time)) = now
		}).Return(nil).Once()
	rows.EXPECT().Next().Return(false).Once()
	rows.EXPECT().Err().Return(nil)
	rows.EXPECT().Close()

	q.EXPECT().Query(ctx, mock.Anything).Return(rows, nil)

	teams, err := repo.ListTeams(ctx)
	require.NoError(t, err)
	require.Len(t, teams, 2)
	assert.Equal(t, "t1", teams[0].Name)
	assert.Equal(t, "t2", teams[1].Name)
}

func TestTeamRepository_AddMember(t *testing.T) {
	t.Run("success or already exists", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		userID := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
		require.NoError(t, repo.AddMember(ctx, teamID, userID))
	})

	t.Run("fk violation -> user not found", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		userID := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), &pgconn.PgError{Code: "23503"})
		err := repo.AddMember(ctx, teamID, userID)
		assert.ErrorIs(t, err, utils.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		userID := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))
		require.Error(t, repo.AddMember(ctx, teamID, userID))
	})
}

func TestTeamRepository_RemoveMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		userID := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("DELETE 1"), nil)
		require.NoError(t, repo.RemoveMember(ctx, teamID, userID))
	})

	t.Run("not found", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		userID := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("DELETE 0"), nil)
		err := repo.RemoveMember(ctx, teamID, userID)
		assert.ErrorIs(t, err, utils.ErrNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, q := newTeamRepo(t)
		ctx := context.Background()
		teamID := uuid.New()
		userID := uuid.New()

		q.EXPECT().Exec(ctx, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))
		require.Error(t, repo.RemoveMember(ctx, teamID, userID))
	})
}
