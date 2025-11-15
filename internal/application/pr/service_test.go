package pr_test

import (
	"context"
	"testing"

	app "avito-test-pr-service/internal/application/pr"
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/infrastructure/logger"
	"avito-test-pr-service/internal/utils"
	"avito-test-pr-service/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPRService_CreatePR(t *testing.T) {
	ctx := context.Background()
	prID := "pr-1"
	authorID := "user-author"
	teamID := uuid.New()
	c1, c2, c3 := "user-r1", "user-r2", "user-r3"

	tests := []struct {
		name    string
		title   string
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector)
		wantErr error
	}{
		{
			name:  "happy two reviewers",
			title: "feat",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetUserByID(ctx, authorID).Return(&models.User{ID: authorID, Name: "a"}, nil)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, authorID).Return(teamID, nil)
				userRepo.EXPECT().ListActiveMembersByTeamID(ctx, teamID).Return([]string{authorID, c1, c2, c3}, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				sel.EXPECT().Select(mock.MatchedBy(func(ids []string) bool { return len(ids) == 3 }), 2).Return([]string{c1, c2})
				prRepo.EXPECT().CreatePR(ctx, mock.MatchedBy(func(pr *models.PullRequest) bool {
					return pr.ID == prID && pr.AuthorID == authorID && pr.Title == "feat" && len(pr.ReviewerIDs) == 2 && pr.ReviewerIDs[0] == c1 && pr.ReviewerIDs[1] == c2
				})).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:  "one candidate -> one reviewer",
			title: "fix",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetUserByID(ctx, authorID).Return(&models.User{ID: authorID}, nil)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, authorID).Return(teamID, nil)
				userRepo.EXPECT().ListActiveMembersByTeamID(ctx, teamID).Return([]string{authorID, c1}, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				sel.EXPECT().Select(mock.Anything, 2).Return([]string{c1})
				prRepo.EXPECT().CreatePR(ctx, mock.MatchedBy(func(pr *models.PullRequest) bool {
					return pr.ID == prID && len(pr.ReviewerIDs) == 1 && pr.ReviewerIDs[0] == c1
				})).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:  "no candidates",
			title: "chore",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetUserByID(ctx, authorID).Return(&models.User{ID: authorID}, nil)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, authorID).Return(teamID, nil)
				userRepo.EXPECT().ListActiveMembersByTeamID(ctx, teamID).Return([]string{authorID}, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				sel.EXPECT().Select([]string{}, 2).Return(nil)
				prRepo.EXPECT().CreatePR(ctx, mock.MatchedBy(func(pr *models.PullRequest) bool { return pr.ID == prID && len(pr.ReviewerIDs) == 0 })).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name:  "author not found",
			title: "feat",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetUserByID(ctx, authorID).Return(nil, utils.ErrUserNotFound)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserNotFound,
		},
		{
			name:  "author no team",
			title: "feat",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetUserByID(ctx, authorID).Return(&models.User{ID: authorID}, nil)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, authorID).Return(uuid.Nil, utils.ErrUserNoTeam)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrUserNoTeam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockUserRepo := mocks.NewUserRepository(t)
			mockPRRepo := mocks.NewPRRepository(t)
			mockSel := mocks.NewReviewerSelector(t)
			log := logger.New("dev")
			if tt.setup != nil {
				tt.setup(mockUOW, mockTx, mockUserRepo, mockPRRepo, mockSel)
			}
			mockTx.EXPECT().PRRepository().Maybe().Return(mockPRRepo)
			mockTx.EXPECT().UserRepository().Maybe().Return(mockUserRepo)
			svc := app.NewService(mockUOW, mockSel, log)
			pr, err := svc.CreatePR(ctx, prID, authorID, tt.title)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, pr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pr)
			}
		})
	}
}

func TestPRService_ReassignReviewer(t *testing.T) {
	ctx := context.Background()
	prID := "pr-2"
	oldID := "user-old"
	newID := "user-new"
	authorID := "user-author"
	teamID := uuid.New()

	tests := []struct {
		name    string
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector)
		wantErr error
	}{
		{
			name: "happy replace",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().LockPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, AuthorID: authorID, Status: models.PRStatusOPEN, ReviewerIDs: []string{oldID}}, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, authorID).Return(teamID, nil)
				userRepo.EXPECT().ListActiveMembersByTeamID(ctx, teamID).Return([]string{authorID, oldID, newID}, nil)
				sel.EXPECT().Select(mock.MatchedBy(func(ids []string) bool { return len(ids) == 1 && ids[0] == newID }), 1).Return([]string{newID})
				prRepo.EXPECT().RemoveReviewer(ctx, prID, oldID).Return(nil)
				prRepo.EXPECT().AddReviewer(ctx, prID, newID).Return(nil)
				prRepo.EXPECT().GetPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, AuthorID: authorID, Status: models.PRStatusOPEN, ReviewerIDs: []string{newID}}, nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name: "merged -> error",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().LockPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, Status: models.PRStatusMERGED}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrAlreadyMerged,
		},
		{
			name: "old not assigned",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().LockPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, Status: models.PRStatusOPEN, ReviewerIDs: []string{}}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrReviewerNotAssigned,
		},
		{
			name: "no candidates",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, userRepo *mocks.UserRepository, prRepo *mocks.PRRepository, sel *mocks.ReviewerSelector) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().LockPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, AuthorID: authorID, Status: models.PRStatusOPEN, ReviewerIDs: []string{oldID}}, nil)
				tx.EXPECT().UserRepository().Return(userRepo)
				userRepo.EXPECT().GetTeamIDByUserID(ctx, authorID).Return(teamID, nil)
				userRepo.EXPECT().ListActiveMembersByTeamID(ctx, teamID).Return([]string{authorID, oldID}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			wantErr: utils.ErrNoReplacementCandidates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockUserRepo := mocks.NewUserRepository(t)
			mockPRRepo := mocks.NewPRRepository(t)
			mockSel := mocks.NewReviewerSelector(t)
			log := logger.New("dev")
			if tt.setup != nil {
				tt.setup(mockUOW, mockTx, mockUserRepo, mockPRRepo, mockSel)
			}
			svc := app.NewService(mockUOW, mockSel, log)
			pr, err := svc.ReassignReviewer(ctx, prID, oldID)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, pr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pr)
			}
		})
	}
}

func TestPRService_MergePR(t *testing.T) {
	ctx := context.Background()
	prID := "pr-merge"
	tests := []struct {
		name    string
		setup   func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository)
		wantErr error
	}{
		{
			name: "open->merged",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().LockPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, Status: models.PRStatusOPEN}, nil)
				prRepo.EXPECT().UpdateStatus(ctx, prID, models.PRStatusMERGED, mock.Anything).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
		{
			name: "already merged idempotent",
			setup: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().LockPRByID(ctx, prID).Return(&models.PullRequest{ID: prID, Status: models.PRStatusMERGED}, nil)
				tx.EXPECT().Commit(ctx).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockPRRepo := mocks.NewPRRepository(t)
			log := logger.New("dev")
			if tt.setup != nil {
				tt.setup(mockUOW, mockTx, mockPRRepo)
			}
			svc := app.NewService(mockUOW, mocks.NewReviewerSelector(t), log)
			pr, err := svc.MergePR(ctx, prID)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, pr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pr)
			}
		})
	}
}

func TestPRService_GetAndList(t *testing.T) {
	ctx := context.Background()
	prID := "pr-get"
	reviewer := "user-r1"
	stOpen := models.PRStatusOPEN

	tests := []struct {
		name      string
		setupGet  func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository)
		setupList func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository)
	}{
		{
			name: "get success and list with filter",
			setupGet: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().GetPRByID(ctx, prID).Return(&models.PullRequest{ID: prID}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			setupList: func(uow *mocks.UnitOfWork, tx *mocks.Transaction, prRepo *mocks.PRRepository) {
				uow.EXPECT().Begin(ctx).Return(tx, nil)
				tx.EXPECT().PRRepository().Return(prRepo)
				prRepo.EXPECT().ListPRsByReviewer(ctx, reviewer, &stOpen).Return([]*models.PullRequest{{ID: prID}}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUOW := mocks.NewUnitOfWork(t)
			mockTx := mocks.NewTransaction(t)
			mockPRRepo := mocks.NewPRRepository(t)
			log := logger.New("dev")
			svc := app.NewService(mockUOW, mocks.NewReviewerSelector(t), log)
			if tt.setupGet != nil {
				tt.setupGet(mockUOW, mockTx, mockPRRepo)
			}
			pr, err := svc.GetPR(ctx, prID)
			require.NoError(t, err)
			require.NotNil(t, pr)
			if tt.setupList != nil {
				tt.setupList(mockUOW, mockTx, mockPRRepo)
			}
			list, err := svc.ListPRsByAssignee(ctx, reviewer, &stOpen)
			require.NoError(t, err)
			require.Len(t, list, 1)
		})
	}
}
