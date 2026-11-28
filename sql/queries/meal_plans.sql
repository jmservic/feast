-- name: CreateMealPlan :one
INSERT INTO meal_plans (name, description, user_id)
VALUES (
	$1,
	$2,
	$3
) RETURNING *;

-- name: CreateUpdateMealPlanItems :exec
CALL create_update_meal_plan_items(@meal_plan_id, @meal_plan_items::meal_plan_item[]);
