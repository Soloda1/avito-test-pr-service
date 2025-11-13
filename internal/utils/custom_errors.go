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
)
