package pr_repository

import (
	"avito-test-pr-service/internal/domain/models"
	ports "avito-test-pr-service/internal/domain/ports/output"
	pr_port "avito-test-pr-service/internal/domain/ports/output/pr"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres"
	"avito-test-pr-service/internal/utils"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PRRepository struct {
	querier postgres.Querier
	log     ports.Logger
}

func NewPRRepository(querier postgres.Querier, log ports.Logger) pr_port.PRRepository {
	return &PRRepository{querier: querier, log: log}
}

func (r *PRRepository) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	if pr.Title == "" || pr.AuthorID == uuid.Nil {
		return utils.ErrInvalidArgument
	}
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	const insertPR = `
		INSERT INTO prs (id, title, author_id, status, created_at, updated_at)
		VALUES (@id, @title, @author_id, 'OPEN', now(), now())
		ON CONFLICT (id) DO NOTHING
		RETURNING id;
	`
	row := r.querier.QueryRow(ctx, insertPR, pgx.NamedArgs{"id": pr.ID, "title": pr.Title, "author_id": pr.AuthorID})
	var insertedID uuid.UUID
	if err := row.Scan(&insertedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn("CreatePR conflict", "pr_id", pr.ID)
			return utils.ErrPRExists
		}
		r.log.Error("CreatePR failed", "pr_id", pr.ID, "err", err)
		return err
	}
	pr.ID = insertedID
	for _, reviewerID := range pr.ReviewerIDs {
		if reviewerID == uuid.Nil {
			continue
		}
		if err := r.AddReviewer(ctx, pr.ID, reviewerID); err != nil {
			return err
		}
	}
	return nil
}

func (r *PRRepository) loadReviewers(ctx context.Context, prID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pr_id = @pr_id
		ORDER BY assigned_at;
	`
	rows, err := r.querier.Query(ctx, q, pgx.NamedArgs{"pr_id": prID})
	if err != nil {
		r.log.Error("loadReviewers query failed", "pr_id", prID, "err", err)
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			r.log.Error("loadReviewers scan failed", "pr_id", prID, "err", err)
			return nil, err
		}
		ids = append(ids, id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ids, nil
}

func (r *PRRepository) GetPRByID(ctx context.Context, id uuid.UUID) (*models.PullRequest, error) {
	const q = `
		SELECT id, title, author_id, status, created_at, merged_at, updated_at
		FROM prs
		WHERE id = @id;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"id": id})
	var pr models.PullRequest
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrPRNotFound
		}
		r.log.Error("GetPRByID failed", "pr_id", id, "err", err)
		return nil, err
	}
	reviewers, err := r.loadReviewers(ctx, pr.ID)
	if err != nil {
		return nil, err
	}
	pr.ReviewerIDs = reviewers
	return &pr, nil
}

func (r *PRRepository) LockPRByID(ctx context.Context, id uuid.UUID) (*models.PullRequest, error) {
	const q = `
		SELECT id, title, author_id, status, created_at, merged_at, updated_at
		FROM prs
		WHERE id = @id
		FOR UPDATE;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"id": id})
	var pr models.PullRequest
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrPRNotFound
		}
		r.log.Error("LockPRByID failed", "pr_id", id, "err", err)
		return nil, err
	}
	reviewers, err := r.loadReviewers(ctx, pr.ID)
	if err != nil {
		return nil, err
	}
	pr.ReviewerIDs = reviewers
	return &pr, nil
}

func (r *PRRepository) CountReviewersByPRID(ctx context.Context, prID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM pr_reviewers WHERE pr_id = @pr_id;`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"pr_id": prID})
	var c int
	if err := row.Scan(&c); err != nil {
		r.log.Error("CountReviewersByPRID failed", "pr_id", prID, "err", err)
		return 0, err
	}
	return c, nil
}

func (r *PRRepository) AddReviewer(ctx context.Context, prID uuid.UUID, reviewerID uuid.UUID) error {
	count, err := r.CountReviewersByPRID(ctx, prID)
	if err != nil {
		return err
	}
	if count >= 2 {
		return utils.ErrTooManyReviewers
	}
	const q = `
		INSERT INTO pr_reviewers (pr_id, reviewer_id, assigned_at)
		VALUES (@pr_id, @reviewer_id, now());
	`
	_, err = r.querier.Exec(ctx, q, pgx.NamedArgs{"pr_id": prID, "reviewer_id": reviewerID})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return utils.ErrReviewerAlreadyAssigned
			}
		}
		r.log.Error("AddReviewer failed", "pr_id", prID, "reviewer_id", reviewerID, "err", err)
		return err
	}
	return nil
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID uuid.UUID, reviewerID uuid.UUID) error {
	const q = `
		DELETE FROM pr_reviewers
		WHERE pr_id = @pr_id AND reviewer_id = @reviewer_id;
	`
	tag, err := r.querier.Exec(ctx, q, pgx.NamedArgs{"pr_id": prID, "reviewer_id": reviewerID})
	if err != nil {
		r.log.Error("RemoveReviewer failed", "pr_id", prID, "reviewer_id", reviewerID, "err", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrReviewerNotAssigned
	}
	return nil
}

func (r *PRRepository) UpdateStatus(ctx context.Context, prID uuid.UUID, status models.PRStatus, mergedAt *time.Time) error {
	if status != models.PRStatusOPEN && status != models.PRStatusMERGED {
		return utils.ErrInvalidStatus
	}
	const q = `
		UPDATE prs
		SET status = @status,
			merged_at = COALESCE(@merged_at, merged_at),
			updated_at = now()
		WHERE id = @id AND status != 'MERGED'
		RETURNING id;
	`
	row := r.querier.QueryRow(ctx, q, pgx.NamedArgs{"status": status, "merged_at": mergedAt, "id": prID})
	var returnedID uuid.UUID
	if err := row.Scan(&returnedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			const existsQ = `SELECT 1 FROM prs WHERE id = @id;`
			row2 := r.querier.QueryRow(ctx, existsQ, pgx.NamedArgs{"id": prID})
			var one int
			if err2 := row2.Scan(&one); err2 != nil {
				if errors.Is(err2, pgx.ErrNoRows) {
					return utils.ErrPRNotFound
				}
				r.log.Error("UpdateStatus exists check failed", "pr_id", prID, "err", err2)
				return err2
			}
			// Exists but no rows affected -> idempotent no-op
			return nil
		}
		r.log.Error("UpdateStatus failed", "pr_id", prID, "err", err)
		return err
	}
	return nil
}

func (r *PRRepository) ListPRsByReviewer(ctx context.Context, reviewerID uuid.UUID, status *models.PRStatus) ([]*models.PullRequest, error) {
	const q = `
		SELECT p.id, p.title, p.author_id, p.status, p.created_at, p.merged_at, p.updated_at
		FROM prs p
		JOIN pr_reviewers r ON p.id = r.pr_id
		WHERE r.reviewer_id = @reviewer_id AND (@status::text IS NULL OR p.status = @status)
		ORDER BY p.created_at DESC;
	`
	rows, err := r.querier.Query(ctx, q, pgx.NamedArgs{"reviewer_id": reviewerID, "status": status})
	if err != nil {
		r.log.Error("ListPRsByReviewer query failed", "reviewer_id", reviewerID, "err", err)
		return nil, err
	}
	defer rows.Close()
	var res []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
			r.log.Error("ListPRsByReviewer scan failed", "reviewer_id", reviewerID, "err", err)
			return nil, err
		}
		res = append(res, &pr)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}
