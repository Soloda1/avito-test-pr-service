package models

import (
	"time"
)

type PullRequest struct {
	ID          string
	Title       string
	AuthorID    string
	Status      PRStatus
	ReviewerIDs []string
	CreatedAt   time.Time
	MergedAt    *time.Time
	UpdatedAt   time.Time
}
