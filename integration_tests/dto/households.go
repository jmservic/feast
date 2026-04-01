package dto

import (
	"github.com/google/uuid"
	"time"
)

type HouseholdCreatePayload struct {
	Name string `json:"name"`
}

type HouseholdCreateResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
}
