package utils

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrTeamNotFound    = errors.New("team not found")
	ErrUserNoTeam      = errors.New("user has no team")
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidArgument = errors.New("invalid argument")
	// PR specific
	ErrPRExists                = errors.New("pr already exists")
	ErrPRNotFound              = errors.New("pr not found")
	ErrTooManyReviewers        = errors.New("too many reviewers")
	ErrReviewerAlreadyAssigned = errors.New("reviewer already assigned")
	ErrReviewerNotAssigned     = errors.New("reviewer not assigned")
	ErrInvalidStatus           = errors.New("invalid status")
)
