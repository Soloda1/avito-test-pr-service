package services

import "github.com/google/uuid"

//go:generate mockery --name ReviewerSelector --dir . --output ../../../mocks --outpkg mocks --with-expecter --filename ReviewerSelector.go

type ReviewerSelector interface {
	Select(candidates []uuid.UUID, count int) []uuid.UUID
}
