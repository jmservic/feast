-- name: CreateHousehold :exec
CALL create_household($1, $2);

-- name: DeleteHousehold :exec
CALL user_delete_household($1, $2);

-- name: GetHouseholdById :one
SELECT * FROM households
WHERE id = $1;

-- name: GetHouseholdByUserId :one
SELECT h.* FROM households h
INNER JOIN household_members m ON ( h.id = m.household_id )
INNER JOIN users u ON ( m.user_id = u.id )
WHERE u.id = $1;

-- name: CreateHouseholdMember :exec
CALL user_create_household_member($1, $2, $3, $4);

-- name: DeleteHouseholdMember :exec
CALL user_delete_household_member($1, $2);

-- name: InviteUserToHousehold :exec
CALL invite_user_to_household($1, $2, $3, $4);

-- name: AcceptHouseholdInvite :exec
CALL accept_household_invite($1, $2);

