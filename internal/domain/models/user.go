package models

import "time"

type User struct {
	ID        string
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
