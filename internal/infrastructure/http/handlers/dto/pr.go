package dto

import (
	"avito-test-pr-service/internal/domain/models"
	"time"
)

type PRDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

func ToPRDTO(pr *models.PullRequest) PRDTO {
	reviewers := make([]string, 0, len(pr.ReviewerIDs))
	for _, id := range pr.ReviewerIDs {
		reviewers = append(reviewers, id.String())
	}
	return PRDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Title,
		AuthorID:          pr.AuthorID.String(),
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
