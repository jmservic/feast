package dto

import (
	"github.com/google/uuid"
	"time"
)

type UserResources struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
