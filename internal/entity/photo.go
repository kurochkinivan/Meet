package entity

import (
	"time"

	"github.com/google/uuid"
)

type Photo struct {
	ID        int
	UserID    uuid.UUID
	URL       string
	ObjectKey string
	CreatedAt time.Time
}
