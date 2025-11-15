package team_test

import (
	"context"
	"errors"
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
				urepo.EXPECT().GetUserByID(ctx, uid.String()).Return(&models.User{ID: uid.String(), Name: "u"}, nil)
				trepo.EXPECT().AddMember(ctx, tid, uid.String()).Return(nil)
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
				urepo.EXPECT().GetUserByID(ctx, uid.String()).Return(nil, utils.ErrUserNotFound)
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
				urepo.EXPECT().GetUserByID(ctx, uid.String()).Return(&models.User{ID: uid.String(), Name: "u"}, nil)
				trepo.EXPECT().RemoveMember(ctx, tid, uid.String()).Return(nil)
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
				urepo.EXPECT().GetUserByID(ctx, uid.String()).Return(nil, utils.ErrUserNotFound)
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

func TestTeamService_CreateTeamWithMembers(t *testing.T) {
	ctx := context.Background()
	teamID := uuid.New()
	existingUserID := "user-existing"

	tests := []struct {
		name      string
		members   []*models.User
		mockSetup func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository)
		wantErr   error
		check     func(t *testing.T, team *models.Team, users []*models.User, err error)
	}{
		{
			name: "happy_mixed_new_and_existing_and_idempotent_add",
			members: []*models.User{
				{ID: "user-alice", Name: "alice", IsActive: true},
				{ID: existingUserID, Name: "bob-new", IsActive: true},
			},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.MatchedBy(func(tm *models.Team) bool { return tm != nil && tm.Name == "backend" })).Run(func(_ context.Context, tm *models.Team) { tm.ID = teamID }).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, "user-alice").Return(nil, utils.ErrUserNotFound)
				urepo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *models.User) bool { return u.ID == "user-alice" && u.Name == "alice" })).Return(nil)
				trepo.EXPECT().AddMember(ctx, teamID, "user-alice").Return(nil)
				urepo.EXPECT().GetUserByID(ctx, existingUserID).Return(&models.User{ID: existingUserID, Name: "bob", IsActive: false}, nil)
				urepo.EXPECT().UpdateUserActive(ctx, existingUserID, true).Return(nil)
				urepo.EXPECT().UpdateUserName(ctx, existingUserID, "bob-new").Return(nil)
				trepo.EXPECT().AddMember(ctx, teamID, existingUserID).Return(utils.ErrAlreadyExists)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
			check: func(t *testing.T, team *models.Team, users []*models.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, team)
				require.Equal(t, teamID, team.ID)
				require.Len(t, users, 2)
			},
		},
		{
			name:    "begin_fail",
			members: []*models.User{},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("begin fail"))
			},
			wantErr: errors.New("begin fail"),
		},
		{
			name:    "create_team_fail",
			members: []*models.User{},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.AnythingOfType("*models.Team")).Return(utils.ErrAlreadyExists)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrAlreadyExists,
		},
		{
			name:    "create_user_fail",
			members: []*models.User{{ID: "user-alice", Name: "alice", IsActive: true}},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.AnythingOfType("*models.Team")).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, "user-alice").Return(nil, utils.ErrUserNotFound) // added
				urepo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *models.User) bool { return u.ID == "user-alice" && u.Name == "alice" })).Return(errors.New("insert fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("insert fail"),
		},
		{
			name:    "get_user_fail_non_notfound",
			members: []*models.User{{ID: existingUserID, Name: "bob", IsActive: true}},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.AnythingOfType("*models.Team")).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, existingUserID).Return(nil, errors.New("db error"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("db error"),
		},
		{
			name:    "update_active_fail",
			members: []*models.User{{ID: existingUserID, Name: "bob", IsActive: true}},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.AnythingOfType("*models.Team")).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, existingUserID).Return(&models.User{ID: existingUserID, Name: "bob", IsActive: false}, nil)
				urepo.EXPECT().UpdateUserActive(ctx, existingUserID, true).Return(errors.New("upd active fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("upd active fail"),
		},
		{
			name:    "update_name_fail",
			members: []*models.User{{ID: existingUserID, Name: "bob-new", IsActive: false}},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.AnythingOfType("*models.Team")).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, existingUserID).Return(&models.User{ID: existingUserID, Name: "bob", IsActive: false}, nil)
				urepo.EXPECT().UpdateUserName(ctx, existingUserID, "bob-new").Return(errors.New("upd name fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("upd name fail"),
		},
		{
			name:    "add_member_fail_non_idempotent",
			members: []*models.User{{ID: "user-alice", Name: "alice", IsActive: true}},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.MatchedBy(func(tm *models.Team) bool { return tm != nil })).Run(func(_ context.Context, tm *models.Team) {
					tm.ID = teamID
				}).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, "user-alice").Return(nil, utils.ErrUserNotFound) // added
				urepo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *models.User) bool { return u.ID == "user-alice" && u.Name == "alice" })).Return(nil)
				trepo.EXPECT().AddMember(ctx, teamID, "user-alice").Return(errors.New("rel fail"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("rel fail"),
		},
		{
			name:    "commit_fail",
			members: []*models.User{{ID: "user-alice", Name: "alice", IsActive: true}},
			mockSetup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository, urepo *mocks.UserRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().CreateTeam(ctx, mock.AnythingOfType("*models.Team")).Return(nil)
				tx.EXPECT().UserRepository().Return(urepo)
				urepo.EXPECT().GetUserByID(ctx, "user-alice").Return(nil, utils.ErrUserNotFound) // added
				urepo.EXPECT().CreateUser(ctx, mock.AnythingOfType("*models.User")).Return(nil)
				trepo.EXPECT().AddMember(ctx, mock.Anything, mock.Anything).Return(nil)
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
			mockTeamRepo := mocks.NewTeamRepository(t)
			mockUserRepo := mocks.NewUserRepository(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockUOW, mockTx, mockTeamRepo, mockUserRepo)
			}

			svc := app.NewService(mockUOW, log)
			team, users, err := svc.CreateTeamWithMembers(ctx, "backend", tt.members)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if tt.check != nil {
				tt.check(t, team, users, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, team)
			}
		})
	}
}

func TestTeamService_GetTeamByName(t *testing.T) {
	ctx := context.Background()
	teamName := "core"
	team := &models.Team{ID: uuid.New(), Name: teamName}

	tests := []struct {
		name    string
		argName string
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository)
		wantErr error
	}{
		{
			name:    "happy",
			argName: teamName,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByName(ctx, teamName).Return(team, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
		},
		{
			name:    "invalid arg",
			argName: "",
			setup:   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {},
			wantErr: utils.ErrInvalidArgument,
		},
		{
			name:    "begin fail",
			argName: teamName,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(nil, errors.New("begin fail"))
			},
			wantErr: errors.New("begin fail"),
		},
		{
			name:    "not found",
			argName: teamName,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByName(ctx, teamName).Return(nil, utils.ErrTeamNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrTeamNotFound,
		},
		{
			name:    "db error",
			argName: teamName,
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, trepo *mocks.TeamRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().TeamRepository().Return(trepo)
				trepo.EXPECT().GetTeamByName(ctx, teamName).Return(nil, errors.New("db error"))
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: errors.New("db error"),
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
			res, err := svc.GetTeamByName(ctx, tt.argName)

			if tt.wantErr != nil {
				require.Error(t, err)
				switch {
				case errors.Is(tt.wantErr, utils.ErrInvalidArgument), errors.Is(tt.wantErr, utils.ErrTeamNotFound):
					require.ErrorIs(t, err, tt.wantErr)
				default:
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, teamName, res.Name)
		})
	}
}
