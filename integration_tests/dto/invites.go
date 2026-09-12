package dto

import (
	"github.com/google/uuid"
	"time"
)

type InviteResponse struct {
	InviterId         uuid.UUID  `json:"inviter_id"`
	InviteeId         uuid.UUID  `json:"invitee_id"`
	HouseholdMemberId *uuid.UUID `json:"household_member_id"`
	HouseholdId       uuid.UUID  `json:"household_id"`
	CreatedAt         time.Time  `json:"created_at"`
}

type InvitePayload struct {
	UserId            uuid.UUID  `json:"user_id"`
	HouseholdId       uuid.UUID  `json:"household_id"`
	HouseholdMemberId *uuid.UUID `json:"household_member_id"`
}

type InviteHandlePayload struct {
	HouseholdId uuid.UUID `json:"household_id"`
	Action      string    `json:"action"`
}
