-- name: CreateHouseholdMember :exec
CALL user_create_household_member(@creator_id, @member_name, sqlc.narg('user_id'), @household_id);

-- name: DeleteHouseholdMember :exec
CALL user_delete_household_member($1, $2);

-- name: InviteUserToHousehold :exec
CALL invite_user_to_household(@inviter, @invitee, @v_household_id, sqlc.narg('v_household_member_id'));

-- name: AcceptHouseholdInvite :exec
CALL accept_household_invite(@invitee_id, @household_id);

-- name: DeclineHouseholdInvite :one
DELETE FROM household_invites
WHERE invitee_id = $1 AND household_id = $2
returning *;

-- name: GetHouseholdMember :one
SELECT * FROM household_members
WHERE id = $1;

-- name: UpdateHouseholdMember :exec
CALL user_update_household_member(@updater_id, @new_name, @new_role, @new_user_id, @household_member_id);

-- name: GetHouseholdMembers :many
SELECT * FROM household_members
WHERE household_id = $1;

-- name: GetHouseholdInvites :many
SELECT * FROM household_invites
WHERE inviter_id = $1 OR invitee_id = $1; 


