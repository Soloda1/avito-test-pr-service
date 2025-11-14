package team

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

func NewService(uow uow.UnitOfWork, log ports.Logger) input.TeamInputPort {
	return &Service{uow: uow, log: log}
}

func (s *Service) CreateTeam(ctx context.Context, name string) (*models.Team, error) {
	if name == "" {
		return nil, utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("CreateTeam begin tx failed", "err", err, "name", name)
		return nil, err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	repo := tx.TeamRepository()
	team := &models.Team{ID: uuid.New(), Name: name}
	if err := repo.CreateTeam(ctx, team); err != nil {
		s.log.Error("CreateTeam repo failed", "err", err, "name", name)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("CreateTeam commit failed", "err", err, "team_id", team.ID)
		return nil, err
	}
	commit = true
	s.log.Info("CreateTeam success", "team_id", team.ID, "name", team.Name)
	return team, nil
}

func (s *Service) AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	if teamID == uuid.Nil || userID == uuid.Nil {
		return utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("AddMember begin tx failed", "err", err, "team_id", teamID, "user_id", userID)
		return err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	teamrepo := tx.TeamRepository()
	if _, err := teamrepo.GetTeamByID(ctx, teamID); err != nil {
		s.log.Error("AddMember team fetch failed", "err", err, "team_id", teamID)
		return err
	}

	userrepo := tx.UserRepository()
	if _, err := userrepo.GetUserByID(ctx, userID); err != nil {
		s.log.Error("AddMember user fetch failed", "err", err, "user_id", userID, "team_id", teamID)
		return err
	}

	if err := teamrepo.AddMember(ctx, teamID, userID); err != nil {
		s.log.Error("AddMember repo failed", "err", err, "team_id", teamID, "user_id", userID)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("AddMember commit failed", "err", err, "team_id", teamID, "user_id", userID)
		return err
	}
	commit = true
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	if teamID == uuid.Nil || userID == uuid.Nil {
		return utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("RemoveMember begin tx failed", "err", err, "team_id", teamID, "user_id", userID)
		return err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	teamrepo := tx.TeamRepository()

	if _, err := teamrepo.GetTeamByID(ctx, teamID); err != nil {
		s.log.Error("RemoveMember team fetch failed", "err", err, "team_id", teamID)
		return err
	}

	userrepo := tx.UserRepository()
	if _, err := userrepo.GetUserByID(ctx, userID); err != nil {
		s.log.Error("RemoveMember user fetch failed", "err", err, "user_id", userID, "team_id", teamID)
		return err
	}

	if err := teamrepo.RemoveMember(ctx, teamID, userID); err != nil {
		s.log.Error("RemoveMember repo failed", "err", err, "team_id", teamID, "user_id", userID)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("RemoveMember commit failed", "err", err, "team_id", teamID, "user_id", userID)
		return err
	}
	commit = true
	return nil
}

func (s *Service) GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	if id == uuid.Nil {
		return nil, utils.ErrInvalidArgument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("GetTeam begin tx failed", "err", err, "team_id", id)
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := tx.TeamRepository()
	team, err := repo.GetTeamByID(ctx, id)
	if err != nil {
		s.log.Error("GetTeam repo failed", "err", err, "team_id", id)
		return nil, err
	}

	return team, nil
}

func (s *Service) ListTeams(ctx context.Context) ([]*models.Team, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("ListTeams begin tx failed", "err", err)
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := tx.TeamRepository()
	res, err := repo.ListTeams(ctx)
	if err != nil {
		s.log.Error("ListTeams repo failed", "err", err)
		return nil, err
	}
	return res, nil
}
