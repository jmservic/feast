-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, email, hashed_password)
VALUES (
	gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2,
	$3
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET
	updated_at = NOW(),
	name = $1,
	email = $2,
	hashed_password = $3
WHERE id = $4
RETURNING *;


