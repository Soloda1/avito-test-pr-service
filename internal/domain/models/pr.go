package models

import (
	"time"

	"github.com/google/uuid"
)

type PullRequest struct {
	ID          uuid.UUID
	Title       string
	AuthorID    uuid.UUID
	Status      PRStatus
	ReviewerIDs []uuid.UUID
	CreatedAt   time.Time
	MergedAt    *time.Time
	UpdatedAt   time.Time
}
