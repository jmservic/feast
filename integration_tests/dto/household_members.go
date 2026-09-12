package dto

import (
	"github.com/google/uuid"
	"time"
)

type HouseholdMemberPayload struct {
	Name   string     `json:"name"`
	UserId *uuid.UUID `json:"user_id"`
}

type HouseholdMemberResponse struct {
	Id          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Role        int        `json:"role"`
	HouseholdId uuid.UUID  `json:"household_id"`
	UserId      *uuid.UUID `json:"user_id"`
}

type HouseholdMemberUpdatePayload struct {
	Name   string    `json:"name"`
	UserId uuid.UUID `json:"user_id"`
	Role   int       `json:"role"`
}
