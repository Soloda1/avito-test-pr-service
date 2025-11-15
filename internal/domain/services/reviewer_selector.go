package services

//go:generate mockery --name ReviewerSelector --dir . --output ../../../mocks --outpkg mocks --with-expecter --filename ReviewerSelector.go

type ReviewerSelector interface {
	Select(candidates []string, count int) []string
}
