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
	tests := []struct {
		name      string
		team      *models.Team
		mockSetup func(*mocks.Querier, *models.Team)
		wantErr   bool
		wantIsErr error
		checkID   bool
	}{
		{
			name: "success",
			team: &models.Team{Name: "devs"},
			mockSetup: func(q *mocks.Querier, team *models.Team) {
				row := mocks.NewRow(t)
				id := uuid.New()
				row.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) { *(args[0].(*uuid.UUID)) = id }).Return(nil)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			checkID: true,
		},
		{
			name: "already exists",
			team: &models.Team{Name: "devs"},
			mockSetup: func(q *mocks.Querier, team *models.Team) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything).Return(pgx.ErrNoRows)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr:   true,
			wantIsErr: utils.ErrAlreadyExists,
		},
		{
			name: "scan error",
			team: &models.Team{Name: "devs"},
			mockSetup: func(q *mocks.Querier, team *models.Team) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything).Return(errors.New("scan error"))
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr: true,
		},
		{
			name:      "invalid arg",
			team:      &models.Team{Name: ""},
			wantErr:   true,
			wantIsErr: utils.ErrInvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newTeamRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.team)
			}
			err := repo.CreateTeam(context.Background(), tt.team)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(t, err, tt.wantIsErr)
				}
			} else {
				require.NoError(t, err)
				if tt.checkID {
					assert.NotEqual(t, uuid.Nil, tt.team.ID)
				}
			}
		})
	}
}

func TestTeamRepository_GetTeamByID(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*mocks.Querier, uuid.UUID)
		wantErr   bool
		wantIsErr error
		wantTeam  *models.Team
	}{
		{
			name: "success",
			id:   uuid.New(),
			mockSetup: func(q *mocks.Querier, id uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args ...interface{}) {
						*(args[0].(*uuid.UUID)) = id
						*(args[1].(*string)) = "devs"
						*(args[2].(*time.Time)) = now
						*(args[3].(*time.Time)) = now
					}).Return(nil)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantTeam: &models.Team{Name: "devs"},
		},
		{
			name: "not found",
			id:   uuid.New(),
			mockSetup: func(q *mocks.Querier, id uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr:   true,
			wantIsErr: utils.ErrTeamNotFound,
		},
		{
			name: "scan error",
			id:   uuid.New(),
			mockSetup: func(q *mocks.Querier, id uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error"))
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newTeamRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.id)
			}
			team, err := repo.GetTeamByID(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(t, err, tt.wantIsErr)
				}
				assert.Nil(t, team)
			} else {
				require.NoError(t, err)
				require.NotNil(t, team)
				assert.Equal(t, tt.wantTeam.Name, team.Name)
				assert.Equal(t, tt.id, team.ID)
			}
		})
	}
}

func TestTeamRepository_ListTeams(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	tests := []struct {
		name      string
		mockSetup func(*mocks.Querier)
		wantErr   bool
	}{
		{
			name: "success",
			mockSetup: func(q *mocks.Querier) {
				rows := mocks.NewRows(t)
				id1 := uuid.New()
				id2 := uuid.New()
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
				q.EXPECT().Query(mock.Anything, mock.Anything).Return(rows, nil)
			},
		},
		{
			name: "query error",
			mockSetup: func(q *mocks.Querier) {
				q.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "scan error",
			mockSetup: func(q *mocks.Querier) {
				rows := mocks.NewRows(t)
				rows.EXPECT().Next().Return(true).Once()
				rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error")).Once()
				rows.EXPECT().Close()
				q.EXPECT().Query(mock.Anything, mock.Anything).Return(rows, nil)
			},
			wantErr: true,
		},
		{
			name: "rows error",
			mockSetup: func(q *mocks.Querier) {
				rows := mocks.NewRows(t)
				rows.EXPECT().Next().Return(false).Once()
				rows.EXPECT().Err().Return(errors.New("rows err"))
				rows.EXPECT().Close()
				q.EXPECT().Query(mock.Anything, mock.Anything).Return(rows, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newTeamRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q)
			}
			teams, err := repo.ListTeams(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, teams)
			}
		})
	}
}

func TestTeamRepository_AddMember(t *testing.T) {
	tests := []struct {
		name      string
		teamID    uuid.UUID
		userID    uuid.UUID
		mockSetup func(*mocks.Querier, uuid.UUID, uuid.UUID)
		wantErr   bool
		wantIsErr error
	}{
		{
			name:   "success",
			teamID: uuid.New(),
			userID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID, userID uuid.UUID) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
			},
		},
		{
			name:   "fk violation -> user not found",
			teamID: uuid.New(),
			userID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID, userID uuid.UUID) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), &pgconn.PgError{Code: "23503"})
			},
			wantErr:   true,
			wantIsErr: utils.ErrUserNotFound,
		},
		{
			name:   "db error",
			teamID: uuid.New(),
			userID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID, userID uuid.UUID) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newTeamRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.teamID, tt.userID)
			}
			err := repo.AddMember(context.Background(), tt.teamID, tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(t, err, tt.wantIsErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTeamRepository_RemoveMember(t *testing.T) {
	tests := []struct {
		name      string
		teamID    uuid.UUID
		userID    uuid.UUID
		mockSetup func(*mocks.Querier, uuid.UUID, uuid.UUID)
		wantErr   bool
		wantIsErr error
	}{
		{
			name:   "success",
			teamID: uuid.New(),
			userID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID, userID uuid.UUID) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("DELETE 1"), nil)
			},
		},
		{
			name:   "not found",
			teamID: uuid.New(),
			userID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID, userID uuid.UUID) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("DELETE 0"), nil)
			},
			wantErr:   true,
			wantIsErr: utils.ErrNotFound,
		},
		{
			name:   "db error",
			teamID: uuid.New(),
			userID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID, userID uuid.UUID) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newTeamRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.teamID, tt.userID)
			}
			err := repo.RemoveMember(context.Background(), tt.teamID, tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(t, err, tt.wantIsErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
