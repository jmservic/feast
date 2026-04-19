package dto

import (
	"github.com/google/uuid"
	"time"
)

type HouseholdMemberResources struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Role        int       `json:"role"`
	HouseholdId uuid.UUID `json:"household_id"`
	UserId      uuid.UUID `json:"user_id"`
}
