package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID        uuid.UUID
	Name      string
	Email     string
	Password  string
	Location  Coordiantes
	CreatedAt time.Time
}

type Coordiantes struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
