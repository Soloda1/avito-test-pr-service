package team

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
	uow "avito-test-pr-service/internal/domain/ports/output/uow"
	user_port "avito-test-pr-service/internal/domain/ports/output/user"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"

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
	userIDStr := userID.String()

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
	if _, err := userrepo.GetUserByID(ctx, userIDStr); err != nil {
		s.log.Error("AddMember user fetch failed", "err", err, "user_id", userIDStr, "team_id", teamID)
		return err
	}

	if err := teamrepo.AddMember(ctx, teamID, userIDStr); err != nil {
		s.log.Error("AddMember repo failed", "err", err, "team_id", teamID, "user_id", userIDStr)
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
	userIDStr := userID.String()

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
	if _, err := userrepo.GetUserByID(ctx, userIDStr); err != nil {
		s.log.Error("RemoveMember user fetch failed", "err", err, "user_id", userIDStr, "team_id", teamID)
		return err
	}

	if err := teamrepo.RemoveMember(ctx, teamID, userIDStr); err != nil {
		s.log.Error("RemoveMember repo failed", "err", err, "team_id", teamID, "user_id", userIDStr)
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

func (s *Service) CreateTeamWithMembers(ctx context.Context, name string, members []*models.User) (*models.Team, []*models.User, error) {
	if name == "" {
		return nil, nil, utils.ErrInvalidArgument
	}
	for idx, m := range members {
		if m == nil {
			s.log.Error("CreateTeamWithMembers invalid member (nil)", "index", idx)
			return nil, nil, utils.ErrInvalidArgument
		}
		if m.ID == "" {
			s.log.Error("CreateTeamWithMembers invalid member id (empty)", "index", idx)
			return nil, nil, utils.ErrInvalidArgument
		}
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("CreateTeamWithMembers begin tx failed", "err", err, "name", name)
		return nil, nil, err
	}
	var commit bool
	defer func() {
		if !commit {
			_ = tx.Rollback(ctx)
		}
	}()

	teamRepo := tx.TeamRepository()
	team := &models.Team{ID: uuid.New(), Name: name}
	if err := teamRepo.CreateTeam(ctx, team); err != nil {
		s.log.Error("CreateTeamWithMembers create team failed", "err", err, "name", name)
		return nil, nil, err
	}

	userRepo := tx.UserRepository()
	var resultUsers []*models.User
	for _, memberSpec := range members {
		processedUser, err := s.processTeamMember(ctx, userRepo, memberSpec)
		if err != nil {
			return nil, nil, err
		}
		if err := teamRepo.AddMember(ctx, team.ID, processedUser.ID); err != nil {
			if !errors.Is(err, utils.ErrAlreadyExists) {
				s.log.Error("CreateTeamWithMembers add member failed", "err", err, "team_id", team.ID, "user_id", processedUser.ID)
				return nil, nil, err
			}
		}
		resultUsers = append(resultUsers, processedUser)
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("CreateTeamWithMembers commit failed", "err", err, "team_id", team.ID)
		return nil, nil, err
	}
	commit = true
	s.log.Info("CreateTeamWithMembers success", "team_id", team.ID, "name", team.Name, "members_count", len(resultUsers))
	return team, resultUsers, nil
}

func (s *Service) processTeamMember(ctx context.Context, userRepo user_port.UserRepository, member *models.User) (*models.User, error) {
	if member.ID == "" {
		return nil, utils.ErrInvalidArgument
	}

	existingUser, err := userRepo.GetUserByID(ctx, member.ID)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			if err := userRepo.CreateUser(ctx, member); err != nil {
				if errors.Is(err, utils.ErrUserExists) {
					if existing, gerr := userRepo.GetUserByID(ctx, member.ID); gerr == nil {
						return existing, nil
					}
					s.log.Error("processTeamMember fetch after conflict failed", "err", err, "user_id", member.ID)
					return nil, err
				}
				s.log.Error("processTeamMember create user failed", "err", err, "user_id", member.ID)
				return nil, err
			}
			return member, nil
		}
		s.log.Error("processTeamMember get user failed", "err", err, "user_id", member.ID)
		return nil, err
	}

	return s.updateExistingUser(ctx, userRepo, existingUser, member)
}

func (s *Service) updateExistingUser(ctx context.Context, userRepo user_port.UserRepository, existing *models.User, spec *models.User) (*models.User, error) {
	updatedUser := &models.User{ID: existing.ID, Name: existing.Name, IsActive: existing.IsActive}
	if spec.IsActive != existing.IsActive {
		if err := userRepo.UpdateUserActive(ctx, existing.ID, spec.IsActive); err != nil {
			s.log.Error("updateExistingUser update active failed", "err", err, "user_id", existing.ID)
			return nil, err
		}
		updatedUser.IsActive = spec.IsActive
	}
	if spec.Name != "" && spec.Name != existing.Name {
		if err := userRepo.UpdateUserName(ctx, existing.ID, spec.Name); err != nil {
			s.log.Error("updateExistingUser update name failed", "err", err, "user_id", existing.ID)
			return nil, err
		}
		updatedUser.Name = spec.Name
	}
	return updatedUser, nil
}

func (s *Service) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	if name == "" {
		return nil, utils.ErrInvalidArgument
	}
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("GetTeamByName begin tx failed", "err", err, "team_name", name)
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := tx.TeamRepository()
	team, err := repo.GetTeamByName(ctx, name)
	if err != nil {
		s.log.Error("GetTeamByName repo failed", "err", err, "team_name", name)
		return nil, err
	}
	return team, nil
}
