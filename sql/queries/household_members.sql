-- name: CreateHouseholdMember :exec
CALL user_create_household_member($1, $2, $3, $4);

-- name: DeleteHouseholdMember :exec
CALL user_delete_household_member($1, $2);

-- name: InviteUserToHousehold :exec
CALL invite_user_to_household($1, $2, $3, $4);

-- name: AcceptHouseholdInvite :exec
CALL accept_household_invite($1, $2);

-- name: GetHouseholdMember :one
SELECT * FROM household_members
WHERE id = $1;

-- name: UpdateHouseholdMember :exec
CALL user_update_household_member(@updater_id, @new_name, @new_role, @new_user_id, @household_member_id);

-- name: GetHouseholdMembers :many
SELECT * FROM household_members
WHERE household_id = $1;

