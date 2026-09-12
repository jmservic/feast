package dto

import (
	"github.com/google/uuid"
	"time"
)

type HouseholdResources struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated-at"`
	Name      string    `json:"name"`
}
