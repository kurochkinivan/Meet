package entity

import (
	"time"

	"github.com/google/uuid"
)

type Photo struct {
	ID        int
	UserID    uuid.UUID
	Url       string
	CreatedAt time.Time
}
