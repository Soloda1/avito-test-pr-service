package team_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	team_port "avito-test-pr-service/internal/domain/ports/output/team"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"
	"time"

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
		RETURNING id, name, created_at, updated_at;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"id": team.ID, "name": team.Name})
	var createdID uuid.UUID
	var createdName string
	var createdAt, updatedAt time.Time
	if err := row.Scan(&createdID, &createdName, &createdAt, &updatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation on name
			r.log.Error("CreateTeam unique violation", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_name", team.Name, "err", pgErr)
			return utils.ErrAlreadyExists
		}
		if errors.As(err, &pgErr) {
			r.log.Error("CreateTeam pg error", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_name", team.Name, "err", pgErr)
		}
		r.log.Error("CreateTeam failed", "team_name", team.Name, "err", err)
		return err
	}
	team.ID = createdID
	team.Name = createdName
	team.CreatedAt = createdAt
	team.UpdatedAt = updatedAt
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

func (r *TeamRepository) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	const q = `
		SELECT id, name, created_at, updated_at
		FROM teams
		WHERE name = @name;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"name": name})
	var t models.Team
	if err := row.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrTeamNotFound
		}
		r.log.Error("GetTeamByName failed", "team_name", name, "err", err)
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

func (r *TeamRepository) AddMember(ctx context.Context, teamID uuid.UUID, userID string) error {
	const q = `
		INSERT INTO team_members (team_id, user_id)
		VALUES (@team_id, @user_id)
		ON CONFLICT DO NOTHING;
	`
	tag, err := r.querier.Exec(ctx, q, pgx.NamedArgs{"team_id": teamID, "user_id": userID})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "team_members_user_id_fkey":
					r.log.Error("AddMember FK violation (user)", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_id", teamID, "user_id", userID, "err", pgErr)
					return utils.ErrUserNotFound
				case "team_members_team_id_fkey":
					r.log.Error("AddMember FK violation (team)", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_id", teamID, "user_id", userID, "err", pgErr)
					return utils.ErrTeamNotFound
				default:
					r.log.Error("AddMember FK violation (unknown)", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_id", teamID, "user_id", userID, "err", pgErr)
					return utils.ErrNotFound
				}
			default:
				r.log.Error("AddMember pg error", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_id", teamID, "user_id", userID, "err", pgErr)
			}
		}
		r.log.Error("AddMember failed", "team_id", teamID, "user_id", userID, "err", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrAlreadyExists
	}
	return nil
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID uuid.UUID, userID string) error {
	const q = `
		DELETE FROM team_members
		WHERE team_id = @team_id AND user_id = @user_id
		RETURNING team_id;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"team_id": teamID, "user_id": userID})
	var returnedTeamID uuid.UUID
	if err := row.Scan(&returnedTeamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("RemoveMember pg error", "code", pgErr.Code, "constraint", pgErr.ConstraintName, "team_id", teamID, "user_id", userID, "err", pgErr)
		}
		r.log.Error("RemoveMember failed", "team_id", teamID, "user_id", userID, "err", err)
		return err
	}
	return nil
}
