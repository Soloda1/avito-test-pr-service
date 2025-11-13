package team_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	team_port "avito-test-pr-service/internal/domain/ports/output/team"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamRepository struct {
	querier postgres.Querier
	log     ports.Logger
}

func NewTeamRepository(querier postgres.Querier, log ports.Logger) team_port.TeamRepository {
	return &TeamRepository{querier: querier, log: log}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	if team.Name == "" {
		return utils.ErrInvalidArgument
	}
	if team.ID == uuid.Nil {
		team.ID = uuid.New()
	}
	const q = `
		INSERT INTO teams (id, name, created_at, updated_at)
		VALUES (@id, @name, now(), now())
		ON CONFLICT (name) DO NOTHING
		RETURNING id;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"id": team.ID, "name": team.Name})
	var insertedID uuid.UUID
	if err := row.Scan(&insertedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn("CreateTeam conflict", "team_name", team.Name)
			return utils.ErrAlreadyExists
		}
		r.log.Error("CreateTeam failed", "team_name", team.Name, "err", err)
		return err
	}
	team.ID = insertedID
	return nil
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	const q = `
		SELECT id, name, created_at, updated_at
		FROM teams
		WHERE id = @id;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"id": id})
	var t models.Team
	if err := row.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrTeamNotFound
		}
		r.log.Error("GetTeamByID failed", "team_id", id, "err", err)
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepository) ListTeams(ctx context.Context) ([]*models.Team, error) {
	const q = `
		SELECT id, name, created_at, updated_at
		FROM teams;
	`
	rows, err := r.querier.Query(ctx, q)
	if err != nil {
		r.log.Error("ListTeams query failed", "err", err)
		return nil, err
	}
	defer rows.Close()

	var res []*models.Team
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			r.log.Error("ListTeams scan failed", "err", err)
			return nil, err
		}
		res = append(res, &t)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	const q = `
		INSERT INTO team_members (team_id, user_id)
		VALUES (@team_id, @user_id)
		ON CONFLICT DO NOTHING;
	`
	tag, err := r.querier.Exec(ctx, q, pgx.NamedArgs{"team_id": teamID, "user_id": userID})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				cn := strings.ToLower(pgErr.ConstraintName)
				if strings.Contains(cn, "user") {
					r.log.Error("AddMember FK violation (user)", "team_id", teamID, "user_id", userID, "err", err)
					return utils.ErrUserNotFound
				}
				if strings.Contains(cn, "team") {
					r.log.Error("AddMember FK violation (team)", "team_id", teamID, "user_id", userID, "err", err)
					return utils.ErrTeamNotFound
				}
				r.log.Error("AddMember FK violation (unknown constraint)", "constraint", pgErr.ConstraintName, "team_id", teamID, "user_id", userID, "err", err)
				return utils.ErrNotFound
			}
		}
		r.log.Error("AddMember failed", "team_id", teamID, "user_id", userID, "err", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		r.log.Warn("AddMember already exists", "team_id", teamID, "user_id", userID)
		return utils.ErrAlreadyExists
	}
	return nil
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	const q = `
		DELETE FROM team_members
		WHERE team_id = @team_id AND user_id = @user_id;
	`
	tag, err := r.querier.Exec(ctx, q, pgx.NamedArgs{"team_id": teamID, "user_id": userID})
	if err != nil {
		r.log.Error("RemoveMember failed", "team_id", teamID, "user_id", userID, "err", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}
