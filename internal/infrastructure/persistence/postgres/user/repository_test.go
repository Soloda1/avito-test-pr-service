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
	"avito-test-pr-service/internal/infrastructure/logger"
	user_repository "avito-test-pr-service/internal/infrastructure/persistence/postgres/user"
	"avito-test-pr-service/internal/utils"
	"avito-test-pr-service/mocks"
)

func newRepo(t *testing.T) (user_port.UserRepository, *mocks.Querier) {
	q := mocks.NewQuerier(t)
	log := logger.New("dev")
	return user_repository.NewUserRepository(q, log), q
}

func TestUserRepository_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		user      *models.User
		mockSetup func(*mocks.Querier, *models.User)
		wantErr   bool
		wantIsErr error
	}{
		{
			name: "success",
			user: &models.User{ID: uuid.New(), Name: "alice", IsActive: true},
			mockSetup: func(q *mocks.Querier, u *models.User) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
			},
		},
		{
			name: "unique violation",
			user: &models.User{ID: uuid.New(), Name: "alice", IsActive: true},
			mockSetup: func(q *mocks.Querier, u *models.User) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), &pgconn.PgError{Code: "23505"})
			},
			wantErr:   true,
			wantIsErr: utils.ErrUserExists,
		},
		{
			name: "db error",
			user: &models.User{ID: uuid.New(), Name: "alice", IsActive: true},
			mockSetup: func(q *mocks.Querier, u *models.User) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.user)
			}
			err := repo.CreateUser(context.Background(), tt.user)
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

func TestUserRepository_GetUserByID(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	id := uuid.New()
	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*mocks.Querier, uuid.UUID)
		wantErr   bool
		wantIsErr error
		wantUser  *models.User
	}{
		{
			name: "success",
			id:   id,
			mockSetup: func(q *mocks.Querier, id uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args ...interface{}) {
						*(args[0].(*uuid.UUID)) = id
						*(args[1].(*string)) = "bob"
						*(args[2].(*bool)) = true
						*(args[3].(*time.Time)) = now
						*(args[4].(*time.Time)) = now
					}).Return(nil)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantUser: &models.User{ID: id, Name: "bob", IsActive: true, CreatedAt: now, UpdatedAt: now},
		},
		{
			name: "not found",
			id:   uuid.New(),
			mockSetup: func(q *mocks.Querier, id uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr:   true,
			wantIsErr: utils.ErrUserNotFound,
		},
		{
			name: "scan error",
			id:   uuid.New(),
			mockSetup: func(q *mocks.Querier, id uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error"))
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.id)
			}
			u, err := repo.GetUserByID(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(t, err, tt.wantIsErr)
				}
				assert.Nil(t, u)
			} else {
				require.NoError(t, err)
				require.NotNil(t, u)
				assert.Equal(t, tt.wantUser.ID, u.ID)
				assert.Equal(t, tt.wantUser.Name, u.Name)
				assert.Equal(t, tt.wantUser.IsActive, u.IsActive)
			}
		})
	}
}

func TestUserRepository_UpdateUserActive(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name      string
		id        uuid.UUID
		active    bool
		mockSetup func(*mocks.Querier, uuid.UUID, bool)
		wantErr   bool
		wantIsErr error
	}{
		{
			name:   "success",
			id:     id,
			active: true,
			mockSetup: func(q *mocks.Querier, id uuid.UUID, active bool) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
		},
		{
			name:   "not found",
			id:     uuid.New(),
			active: true,
			mockSetup: func(q *mocks.Querier, id uuid.UUID, active bool) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("UPDATE 0"), nil)
			},
			wantErr:   true,
			wantIsErr: utils.ErrUserNotFound,
		},
		{
			name:   "db error",
			id:     uuid.New(),
			active: true,
			mockSetup: func(q *mocks.Querier, id uuid.UUID, active bool) {
				q.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag(""), errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.id, tt.active)
			}
			err := repo.UpdateUserActive(context.Background(), tt.id, tt.active)
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

func TestUserRepository_GetTeamIDByUserID(t *testing.T) {
	tests := []struct {
		name      string
		uid       uuid.UUID
		mockSetup func(*mocks.Querier, uuid.UUID)
		wantErr   bool
		wantIsErr error
		wantID    uuid.UUID
	}{
		{
			name: "success",
			uid:  uuid.New(),
			mockSetup: func(q *mocks.Querier, uid uuid.UUID) {
				row := mocks.NewRow(t)
				teamID := uuid.New()
				row.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) { *(args[0].(*uuid.UUID)) = teamID }).Return(nil)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
		},
		{
			name: "no team (no rows)",
			uid:  uuid.New(),
			mockSetup: func(q *mocks.Querier, uid uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything).Return(pgx.ErrNoRows)
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr:   true,
			wantIsErr: utils.ErrUserNoTeam,
		},
		{
			name: "scan error",
			uid:  uuid.New(),
			mockSetup: func(q *mocks.Querier, uid uuid.UUID) {
				row := mocks.NewRow(t)
				row.EXPECT().Scan(mock.Anything).Return(errors.New("scan error"))
				q.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(row)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.uid)
			}
			id, err := repo.GetTeamIDByUserID(context.Background(), tt.uid)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(t, err, tt.wantIsErr)
				}
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, id)
			}
		})
	}
}

func TestUserRepository_ListActiveMembersByTeamID(t *testing.T) {
	tests := []struct {
		name      string
		teamID    uuid.UUID
		mockSetup func(*mocks.Querier, uuid.UUID)
		wantErr   bool
		wantIsErr error
		wantIDs   []uuid.UUID
	}{
		{
			name:   "success",
			teamID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID uuid.UUID) {
				rows := mocks.NewRows(t)
				id1 := uuid.New()
				id2 := uuid.New()
				rows.EXPECT().Next().Return(true).Once()
				rows.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) { *(args[0].(*uuid.UUID)) = id1 }).Return(nil).Once()
				rows.EXPECT().Next().Return(true).Once()
				rows.EXPECT().Scan(mock.Anything).Run(func(args ...interface{}) { *(args[0].(*uuid.UUID)) = id2 }).Return(nil).Once()
				rows.EXPECT().Next().Return(false).Once()
				rows.EXPECT().Err().Return(nil)
				rows.EXPECT().Close()
				q.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(rows, nil)
			},
		},
		{
			name:   "query error",
			teamID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID uuid.UUID) {
				q.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "scan error",
			teamID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID uuid.UUID) {
				rows := mocks.NewRows(t)
				rows.EXPECT().Next().Return(true).Once()
				rows.EXPECT().Scan(mock.Anything).Return(errors.New("scan error")).Once()
				rows.EXPECT().Close()
				q.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(rows, nil)
			},
			wantErr: true,
		},
		{
			name:   "rows error",
			teamID: uuid.New(),
			mockSetup: func(q *mocks.Querier, teamID uuid.UUID) {
				rows := mocks.NewRows(t)
				rows.EXPECT().Next().Return(false).Once()
				rows.EXPECT().Err().Return(errors.New("rows err"))
				rows.EXPECT().Close()
				q.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(rows, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, q := newRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q, tt.teamID)
			}
			ids, err := repo.ListActiveMembersByTeamID(context.Background(), tt.teamID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, ids)
			}
		})
	}
}

func TestUserRepository_ListUsers(t *testing.T) {
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
				rows.EXPECT().Scan(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("scan error")).Once()
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
			repo, q := newRepo(t)
			if tt.mockSetup != nil {
				tt.mockSetup(q)
			}
			users, err := repo.ListUsers(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, users)
			}
		})
	}
}
