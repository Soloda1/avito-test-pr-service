package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
