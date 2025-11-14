package team_test

import (
	"context"
	"testing"

	app "avito-test-pr-service/internal/application/team"
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	"avito-test-pr-service/internal/utils"
	"avito-test-pr-service/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTeamService_CreateTeam(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		nameArg string
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository)
		wantErr error
	}{
		{
			name:    "happy",
			nameArg: "core",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.MatchedBy(func(tm *models.Team) bool { return tm != nil && tm.Name == "core" && tm.ID != uuid.Nil })).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:    "invalid name",
			nameArg: "",
			setup:   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {},
			wantErr: utils.ErrInvalidArgument,
		},
		{
			name:    "repo fail",
			nameArg: "core",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.MatchedBy(func(tm *models.Team) bool { return tm != nil && tm.Name == "core" })).Return(utils.ErrAlreadyExists)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrAlreadyExists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockTeamRepo := mocks.NewTeamRepository(t)
			log := logger.New("dev")
			if tt.setup != nil {
				tt.setup(mockUOW, mockTx, mockTeamRepo)
			}
			svc := app.NewService(mockUOW, log)
			team, err := svc.CreateTeam(ctx, tt.nameArg)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, team)
			} else {
				require.NoError(t, err)
				require.NotNil(t, team)
				require.Equal(t, tt.nameArg, team.Name)
			}
		})
	}
}

func TestTeamService_AddMember(t *testing.T) {
	ctx := context.Background()
	tid := uuid.New()
	uid := uuid.New()
	tests := []struct {
		name    string
		teamID  uuid.UUID
		userID  uuid.UUID
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository)
		wantErr error
	}{
		{
			name:   "happy",
			teamID: tid,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(&models.Team{ID: tid, Name: "core"}, nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, uid).Return(&models.User{ID: uid, Name: "u"}, nil)
				trepo.EXPECT().AddMember(ctx, tid, uid).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:   "invalid args",
			teamID: uuid.Nil,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
			},
			wantErr: utils.ErrInvalidArgument,
		},
		{
			name:   "team not found",
			teamID: tid,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(nil, utils.ErrTeamNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrTeamNotFound,
		},
		{
			name:   "user not found",
			teamID: tid,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(&models.Team{ID: tid, Name: "core"}, nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, uid).Return(nil, utils.ErrUserNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockTeamRepo := mocks.NewTeamRepository(t)
			mockUserRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")
			if tt.setup != nil {
				tt.setup(mockUOW, mockTx, mockTeamRepo, mockUserRepo)
			}
			svc := app.NewService(mockUOW, log)
			err := svc.AddMember(ctx, tt.teamID, tt.userID)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTeamService_RemoveMember(t *testing.T) {
	ctx := context.Background()
	tid := uuid.New()
	uid := uuid.New()
	tests := []struct {
		name    string
		teamID  uuid.UUID
		userID  uuid.UUID
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository)
		wantErr error
	}{
		{
			name:   "happy",
			teamID: tid,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(&models.Team{ID: tid, Name: "core"}, nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, uid).Return(&models.User{ID: uid, Name: "u"}, nil)
				trepo.EXPECT().RemoveMember(ctx, tid, uid).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:   "invalid args",
			teamID: uuid.Nil,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
			},
			wantErr: utils.ErrInvalidArgument,
		},
		{
			name:   "team not found",
			teamID: tid,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(nil, utils.ErrTeamNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrTeamNotFound,
		},
		{
			name:   "user not found",
			teamID: tid,
			userID: uid,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(&models.Team{ID: tid, Name: "core"}, nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, uid).Return(nil, utils.ErrUserNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockTeamRepo := mocks.NewTeamRepository(t)
			mockUserRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")
			if tt.setup != nil {
				tt.setup(mockUOW, mockTx, mockTeamRepo, mockUserRepo)
			}
			svc := app.NewService(mockUOW, log)
			err := svc.RemoveMember(ctx, tt.teamID, tt.userID)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTeamService_GetTeam_ListTeams(t *testing.T) {
	ctx := context.Background()
	tid := uuid.New()
	tests := []struct {
		name      string
		setupGet  func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository)
		setupList func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository)
	}{
		{
			name: "get ok & list ok",
			setupGet: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByID(ctx, tid).Return(&models.Team{ID: tid, Name: "core"}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			setupList: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().ListTeams(ctx).Return([]*models.Team{{ID: tid, Name: "core"}}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockTeamRepo := mocks.NewTeamRepository(t)
			log := logger.New("dev")
			svc := app.NewService(mockUOW, log)

			if tt.setupGet != nil {
				tt.setupGet(mockUOW, mockTx, mockTeamRepo)
			}
			team, err := svc.GetTeam(ctx, tid)
			require.NoError(t, err)
			require.NotNil(t, team)

			if tt.setupList != nil {
				tt.setupList(mockUOW, mockTx, mockTeamRepo)
			}
			lst, err := svc.ListTeams(ctx)
			require.NoError(t, err)
			require.Len(t, lst, 1)
		})
	}
}
