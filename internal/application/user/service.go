package user

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
	uow "avito-test-pr-service/internal/domain/ports/output/uow"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	uow uow.UnitOfWork
	log ports.Logger
}

func NewService(uow uow.UnitOfWork, log ports.Logger) input.UserInputPort {
	return &Service{uow: uow, log: log}
}

func (s *Service) CreateUser(ctx context.Context, name string, isActive bool) (*models.User, error) {
	if name == "" {
		s.log.Error("CreateUser invalid argument", "err", utils.ErrInvalidArgument, "name", name)
		return nil, utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("CreateUser begin tx failed", "err", err, "name", name)
		return nil, err
	}

	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	userRepo := tx.UserRepository()
	u := &models.User{ID: uuid.New(), Name: name, IsActive: isActive}
	if err := userRepo.CreateUser(ctx, u); err != nil {
		s.log.Error("CreateUser repo failed", "err", err, "user_id", u.ID)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("CreateUser commit failed", "err", err, "user_id", u.ID)
		return nil, err
	}

	commit = true
	s.log.Info("CreateUser success", "user_id", u.ID, "name", u.Name)
	return u, nil
}

func (s *Service) UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	if id == uuid.Nil {
		s.log.Error("UpdateUserActive invalid argument", "err", utils.ErrInvalidArgument, "user_id", id)
		return utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("UpdateUserActive begin tx failed", "err", err, "user_id", id)
		return err
	}

	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	repo := tx.UserRepository()
	_, err = repo.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("UpdateUserActive get failed", "err", err, "user_id", id)
		return err
	}

	if err := repo.UpdateUserActive(ctx, id, isActive); err != nil {
		s.log.Error("UpdateUserActive update failed", "err", err, "user_id", id, "active", isActive)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("UpdateUserActive commit failed", "err", err, "user_id", id)
		return err
	}

	commit = true
	s.log.Info("UpdateUserActive success", "user_id", id, "active", isActive)
	return nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if id == uuid.Nil {
		s.log.Error("GetUser invalid argument", "err", utils.ErrInvalidArgument, "user_id", id)
		return nil, utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("GetUser begin tx failed", "err", err, "user_id", id)
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	userRepo := tx.UserRepository()
	u, err := userRepo.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("GetUser failed", "err", err, "user_id", id)
		return nil, err
	}

	s.log.Info("GetUser success", "user_id", id)
	return u, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]*models.User, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("ListUsers begin tx failed", "err", err)
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	userRepo := tx.UserRepository()
	users, err := userRepo.ListUsers(ctx)
	if err != nil {
		s.log.Error("ListUsers failed", "err", err)
		return nil, err
	}
	s.log.Info("ListUsers success", "count", len(users))
	return users, nil
}

func (s *Service) GetUserTeamName(ctx context.Context, id uuid.UUID) (string, error) {
	if id == uuid.Nil {
		s.log.Error("GetUserTeamName invalid argument", "err", utils.ErrInvalidArgument, "user_id", id)
		return "", utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("GetUserTeamName begin tx failed", "err", err, "user_id", id)
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userRepo := tx.UserRepository()
	teamID, err := userRepo.GetTeamIDByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrUserNoTeam) {
			return "", nil
		}
		s.log.Error("GetUserTeamName get team id failed", "err", err, "user_id", id)
		return "", err
	}

	teamRepo := tx.TeamRepository()
	team, err := teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		s.log.Error("GetUserTeamName get team failed", "err", err, "user_id", id, "team_id", teamID)
		return "", err
	}
	return team.Name, nil
}
