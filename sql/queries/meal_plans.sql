-- name: CreateMealPlan :one
INSERT INTO meal_plans (name, description, user_id)
VALUES (
	$1,
	$2,
	$3
) RETURNING *;
