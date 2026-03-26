-- name: CreateHouseholdMember :exec
CALL user_create_household_member($1, $2, $3, $4);

-- name: DeleteHouseholdMember :exec
CALL user_delete_household_member($1, $2);

-- name: InviteUserToHousehold :exec
CALL invite_user_to_household($1, $2, $3, $4);

-- name: AcceptHouseholdInvite :exec
CALL accept_household_invite($1, $2);

