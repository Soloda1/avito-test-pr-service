package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
