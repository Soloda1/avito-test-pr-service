package services

import "github.com/google/uuid"

type ReviewerSelector interface {
	Select(candidates []uuid.UUID, count int) []uuid.UUID
}
