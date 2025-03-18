package entity

import (
	"time"

	"github.com/google/uuid"
)

type Photo struct {
	ID        int64
	UserID    uuid.UUID
	URL       string
	ObjectKey string
	CreatedAt time.Time
}
