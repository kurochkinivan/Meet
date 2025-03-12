package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID      uuid.UUID
	Name      string
	Email     string
	Password  string
	Location  Coordiantes
	CreatedAt time.Time
	Photos    []*Photo
}

type Coordiantes struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
