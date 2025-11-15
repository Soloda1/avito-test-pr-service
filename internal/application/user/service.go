package user

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
	uow "avito-test-pr-service/internal/domain/ports/output/uow"
	"avito-test-pr-service/internal/utils"
	"context"
	"github.com/google/uuid"
)

type Service struct {
	uow uow.UnitOfWork
	log ports.Logger
}

func NewService(uow uow.UnitOfWork, log ports.Logger) input.UserInputPort {
	return &Service{uow: uow, log: log}
}

func (s *Service) CreateUser(ctx context.Context, id string, name string, isActive bool) (*models.User, error) {
	if id == "" || name == "" {
		s.log.Error("CreateUser invalid argument", "id", id, "name", name)
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("CreateUser begin tx failed", "err", err, "id", id)
		return nil, err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()
	repo := tx.UserRepository()
	u := &models.User{ID: id, Name: name, IsActive: isActive}
	if err := repo.CreateUser(ctx, u); err != nil {
		s.log.Error("CreateUser repo failed", "err", err, "id", id)
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("CreateUser commit failed", "err", err, "id", id)
		return nil, err
	}
	commit = true
	return u, nil
}

func (s *Service) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	if id == "" {
		return utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()
	repo := tx.UserRepository()
	if _, err := repo.GetUserByID(ctx, id); err != nil {
		return err
	}
	if err := repo.UpdateUserActive(ctx, id, isActive); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	commit = true
	return nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	repo := tx.UserRepository()
	u, err := repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]*models.User, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	repo := tx.UserRepository()
	users, err := repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) GetUserTeamName(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	repo := tx.UserRepository()
	teamID, err := repo.GetTeamIDByUserID(ctx, id)
	if err != nil {
		if err == utils.ErrUserNoTeam {
			return "", nil
		}
		return "", err
	}
	teamRepo := tx.TeamRepository()
	team, err := teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return "", err
	}
	return team.Name, nil
}

func (s *Service) UpdateUserName(ctx context.Context, id string, name string) error {
	if id == "" || name == "" {
		return utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()
	repo := tx.UserRepository()
	if err := repo.UpdateUserName(ctx, id, name); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	commit = true
	return nil
}

func (s *Service) ListMembersByTeamID(ctx context.Context, teamID string) ([]*models.User, error) {
	if teamID == "" {
		return nil, utils.ErrInvalidArgument
	}
	parsed, err := uuid.Parse(teamID)
	if err != nil {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	repo := tx.UserRepository()
	res, err := repo.ListMembersByTeamID(ctx, parsed)
	if err != nil {
		return nil, err
	}
	return res, nil
}
