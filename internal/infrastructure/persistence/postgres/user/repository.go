package user_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	user_port "avito-test-pr-service/internal/domain/ports/output/user"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	querier postgres.Querier
	log     ports.Logger
}

func NewUserRepository(querier postgres.Querier, log ports.Logger) user_port.UserRepository {
	return &UserRepository{querier: querier, log: log}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := pgx.NamedArgs{
		"id":        user.ID,
		"name":      user.Name,
		"is_active": user.IsActive,
	}

	const q = `
		INSERT INTO users (id, name, is_active, created_at, updated_at)
		VALUES (@id, @name, @is_active, now(), now())
	`
	_, err := r.querier.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return utils.ErrUserExists
			}
		}
		r.log.Error("CreateUser failed", "user_id", user.ID, "err", err)
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const q = `
		SELECT id, name, is_active, created_at, updated_at
		FROM users
		WHERE id = @id;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"id": id})
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrUserNotFound
		}
		r.log.Error("GetUserByID failed", "user_id", id, "err", err)
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdateUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	const q = `
		UPDATE users
		SET is_active = @is_active,
			updated_at = now()
		WHERE id = @id;
	`
	args := pgx.NamedArgs{"is_active": isActive, "id": id}
	tag, err := r.querier.Exec(ctx, q, args)
	if err != nil {
		r.log.Error("UpdateUserActive failed", "user_id", id, "err", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	const q = `
		SELECT id, name, is_active, created_at, updated_at
		FROM users;
	`
	rows, err := r.querier.Query(ctx, q)
	if err != nil {
		r.log.Error("ListUsers query failed", "err", err)
		return nil, err
	}
	defer rows.Close()

	var res []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			r.log.Error("ListUsers scan failed", "err", err)
			return nil, err
		}
		res = append(res, &u)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

func (r *UserRepository) GetTeamIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	const q = `
		SELECT team_id
		FROM team_members
		WHERE user_id = @user_id
		LIMIT 1;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"user_id": userID})
	var teamID uuid.UUID
	if err := row.Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, utils.ErrUserNoTeam
		}
		r.log.Error("GetTeamIDByUserID failed", "user_id", userID, "err", err)
		return uuid.Nil, err
	}
	return teamID, nil
}

func (r *UserRepository) ListActiveMembersByTeamID(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT u.id
		FROM users u
		JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_id = @team_id AND u.is_active = true;
	`
	rows, err := r.querier.Query(ctx, q, pgx.NamedArgs{"team_id": teamID})
	if err != nil {
		r.log.Error("ListActiveMembersByTeamID query failed", "team_id", teamID, "err", err)
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			r.log.Error("ListActiveMembersByTeamID scan failed", "err", err)
			return nil, err
		}
		ids = append(ids, id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ids, nil
}
