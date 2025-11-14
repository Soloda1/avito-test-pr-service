package user_test

import (
	"context"
	"errors"
	"testing"

	app "avito-test-pr-service/internal/application/user"
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	"avito-test-pr-service/internal/utils"
	"avito-test-pr-service/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		nameArg   string
		active    bool
		mockSetup func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository)
		wantErr   error
		useIs     bool
	}{
		{
			name:    "success",
			nameArg: "alice",
			active:  true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *models.User) bool { return u.Name == "alice" && u.IsActive })).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:      "invalid name",
			nameArg:   "",
			active:    true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {},
			wantErr:   utils.ErrInvalidArgument,
			useIs:     true,
		},
		{
			name:    "begin tx fails",
			nameArg: "bob",
			active:  true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
		{
			name:    "repo create fails unique",
			nameArg: "carol",
			active:  true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *models.User) bool { return u.Name == "carol" && u.IsActive })).Return(utils.ErrUserExists)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserExists,
			useIs:   true,
		},
		{
			name:    "commit fails",
			nameArg: "dave",
			active:  true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *models.User) bool { return u.Name == "dave" && u.IsActive })).Return(nil)
				tx.EXPECT().Commit(ctx).Return(errors.New("commit fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("commit fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")
			if tt.mockSetup != nil {
				tt.mockSetup(mockUOW, mockTx, mockRepo)
			}
			svc := app.NewService(mockUOW, log)
			user, err := svc.CreateUser(ctx, tt.nameArg, tt.active)
			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.useIs {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tt.nameArg, user.Name)
			}
		})
	}
}

func TestUserService_UpdateUserActive(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	tests := []struct {
		name      string
		userID    uuid.UUID
		active    bool
		mockSetup func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository)
		wantErr   error
		useIs     bool
	}{
		{
			name:   "success",
			userID: uid,
			active: false,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().GetUserByID(ctx, uid).Return(&models.User{ID: uid, Name: "alice", IsActive: true}, nil)
				repo.EXPECT().UpdateUserActive(ctx, uid, false).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:      "invalid id",
			userID:    uuid.Nil,
			active:    true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {},
			wantErr:   utils.ErrInvalidArgument,
			useIs:     true,
		},
		{
			name:   "begin fails",
			userID: uid,
			active: true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("begin fail"))
			},
			wantErr: errors.New("begin fail"),
		},
		{
			name:   "get not found",
			userID: uid,
			active: true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().GetUserByID(ctx, uid).Return(nil, utils.ErrUserNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserNotFound,
			useIs:   true,
		},
		{
			name:   "update fails",
			userID: uid,
			active: true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().GetUserByID(ctx, uid).Return(&models.User{ID: uid, Name: "alice", IsActive: true}, nil)
				repo.EXPECT().UpdateUserActive(ctx, uid, true).Return(errors.New("update fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("update fail"),
		},
		{
			name:   "commit fails",
			userID: uid,
			active: true,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().GetUserByID(ctx, uid).Return(&models.User{ID: uid, Name: "alice", IsActive: true}, nil)
				repo.EXPECT().UpdateUserActive(ctx, uid, true).Return(nil)
				tx.EXPECT().Commit(ctx).Return(errors.New("commit fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("commit fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")
			if tt.mockSetup != nil {
				tt.mockSetup(mockUOW, mockTx, mockRepo)
			}
			svc := app.NewService(mockUOW, log)
			err := svc.UpdateUserActive(ctx, tt.userID, tt.active)
			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.useIs {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository)
		wantErr   error
		useIs     bool
	}{
		{
			name:   "success",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().GetUserByID(ctx, uid).Return(&models.User{ID: uid, Name: "alice", IsActive: true}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
		},
		{
			name:      "invalid id",
			userID:    uuid.Nil,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {},
			wantErr:   utils.ErrInvalidArgument,
			useIs:     true,
		},
		{
			name:   "begin fails",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("begin fail"))
			},
			wantErr: errors.New("begin fail"),
		},
		{
			name:   "repo not found",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().GetUserByID(ctx, uid).Return(nil, utils.ErrUserNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserNotFound,
			useIs:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")
			if tt.mockSetup != nil {
				tt.mockSetup(mockUOW, mockTx, mockRepo)
			}
			svc := app.NewService(mockUOW, log)
			u, err := svc.GetUser(ctx, tt.userID)
			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.useIs {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, u)
			} else {
				require.NoError(t, err)
				require.NotNil(t, u)
				require.Equal(t, tt.userID, u.ID)
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	tests := []struct {
		name      string
		mockSetup func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository)
		wantErr   error
		useIs     bool
	}{
		{
			name: "success",
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().ListUsers(ctx).Return([]*models.User{{ID: uid, Name: "alice", IsActive: true}}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
		},
		{
			name: "begin fails",
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("begin fail"))
			},
			wantErr: errors.New("begin fail"),
		},
		{
			name: "repo fails",
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, repo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(repo)
				repo.EXPECT().ListUsers(ctx).Return(nil, errors.New("query fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("query fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")
			if tt.mockSetup != nil {
				tt.mockSetup(mockUOW, mockTx, mockRepo)
			}
			svc := app.NewService(mockUOW, log)
			users, err := svc.ListUsers(ctx)
			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.useIs {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, users)
			} else {
				require.NoError(t, err)
				require.Len(t, users, 1)
				require.Equal(t, uid, users[0].ID)
			}
		})
	}
}

func TestUserService_GetUserTeamName(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	teamID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository)
		wantName  string
		wantErr   error
		useIs     bool
	}{
		{
			name:   "success returns team name",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, uid).Return(teamID, nil)
				tx.EXPECT().TeamRepository().Return(teamRepo)
				teamRepo.EXPECT().GetTeamByID(ctx, teamID).Return(&models.Team{ID: teamID, Name: "core"}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantName: "core",
		},
		{
			name:   "invalid id",
			userID: uuid.Nil,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository) {
			},
			wantErr: utils.ErrInvalidArgument,
			useIs:   true,
		},
		{
			name:   "begin fails",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("begin fail"))
			},
			wantErr: errors.New("begin fail"),
		},
		{
			name:   "no team returns empty name",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, uid).Return(uuid.Nil, utils.ErrUserNoTeam)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantName: "",
		},
		{
			name:   "get team id fails",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, uid).Return(uuid.Nil, errors.New("db fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("db fail"),
		},
		{
			name:   "get team by id fails",
			userID: uid,
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, teamRepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, uid).Return(teamID, nil)
				tx.EXPECT().TeamRepository().Return(teamRepo)
				teamRepo.EXPECT().GetTeamByID(ctx, teamID).Return(nil, errors.New("not found"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockUserRepo := mocks.NewUserRepository(t)
			mockTeamRepo := mocks.NewTeamRepository(t)
			log := logger.New("dev")
			if tt.mockSetup != nil {
				tt.mockSetup(mockUOW, mockTx, mockUserRepo, mockTeamRepo)
			}
			svc := app.NewService(mockUOW, log)
			name, err := svc.GetUserTeamName(ctx, tt.userID)
			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.useIs {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantName, name)
			}
		})
	}
}
