package integration

import (
	"context"
	"errors"
	"time"

	"avito-test-pr-service/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TruncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE pr_reviewers, team_members, prs, users, teams RESTART IDENTITY CASCADE;
	`)
	return err
}

func GetPR(ctx context.Context, pool *pgxpool.Pool, prID string) (*models.PullRequest, error) {
	row := pool.QueryRow(ctx, `SELECT id, title, author_id, status, created_at, merged_at, updated_at FROM prs WHERE id=$1`, prID)
	var pr models.PullRequest
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

func GetPRReviewers(ctx context.Context, pool *pgxpool.Pool, prID string) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1 ORDER BY assigned_at`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}
	return reviewers, rows.Err()
}

func GetPRReviewersCount(ctx context.Context, pool *pgxpool.Pool, prID string) (int, error) {
	row := pool.QueryRow(ctx, `SELECT COUNT(*) FROM pr_reviewers WHERE pr_id=$1`, prID)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}

func GetPRStatusMerged(ctx context.Context, pool *pgxpool.Pool, prID string) (string, *time.Time, error) {
	row := pool.QueryRow(ctx, `SELECT status, merged_at FROM prs WHERE id=$1`, prID)
	var status string
	var mergedAt *time.Time
	if err := row.Scan(&status, &mergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, nil
		}
		return "", nil, err
	}
	return status, mergedAt, nil
}

func GetTeamMemberIDs(ctx context.Context, pool *pgxpool.Pool, teamID uuid.UUID) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT user_id FROM team_members WHERE team_id=$1`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func GetUser(ctx context.Context, pool *pgxpool.Pool, userID string) (*models.User, error) {
	row := pool.QueryRow(ctx, `SELECT id, name, is_active, created_at, updated_at FROM users WHERE id=$1`, userID)
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func EqualStringSets(a, b []string) bool {
	if len(a) != len(b) { // quick length check (assumes no duplicates expected)
		return false
	}
	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		m[v]--
	}
	for _, c := range m {
		if c != 0 {
			return false
		}
	}
	return true
}

func HasDuplicates(vals []string) bool {
	seen := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		if _, ok := seen[v]; ok {
			return true
		}
		seen[v] = struct{}{}
	}
	return false
}

func InsertUser(ctx context.Context, pool *pgxpool.Pool, id, name string, active bool) error {
	_, err := pool.Exec(ctx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
	return err
}

func InsertTeam(ctx context.Context, pool *pgxpool.Pool, name string) (uuid.UUID, error) {
	row := pool.QueryRow(ctx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(),$1,now(),now()) RETURNING id`, name)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func AddTeamMember(ctx context.Context, pool *pgxpool.Pool, teamID uuid.UUID, userID string) error {
	_, err := pool.Exec(ctx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2)`, teamID, userID)
	return err
}

func InsertPR(ctx context.Context, pool *pgxpool.Pool, prID, title, authorID string) error {
	_, err := pool.Exec(ctx, `INSERT INTO prs(id, title, author_id, status, created_at, updated_at) VALUES ($1,$2,$3,'OPEN',now(),now())`, prID, title, authorID)
	return err
}

func AddPRReviewer(ctx context.Context, pool *pgxpool.Pool, prID, reviewerID string) error {
	_, err := pool.Exec(ctx, `INSERT INTO pr_reviewers(pr_id, reviewer_id, assigned_at) VALUES ($1,$2,now())`, prID, reviewerID)
	return err
}

func UpdateUserActive(ctx context.Context, pool *pgxpool.Pool, userID string, active bool) error {
	_, err := pool.Exec(ctx, `UPDATE users SET is_active=$2, updated_at=now() WHERE id=$1`, userID, active)
	return err
}

func UpdateUsersActive(ctx context.Context, pool *pgxpool.Pool, userIDs []string, active bool) error {
	if len(userIDs) == 0 {
		return nil
	}
	_, err := pool.Exec(ctx, `UPDATE users SET is_active=$2, updated_at=now() WHERE id = ANY($1)`, userIDs, active)
	return err
}
