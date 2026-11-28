-- name: CreateMealPlan :one
INSERT INTO meal_plans (name, description, user_id)
VALUES (
	$1,
	$2,
	$3
) RETURNING *;

-- name: CreateUpdateMealPlanItems :exec
CALL create_update_meal_plan_items(@meal_plan_id, @meal_plan_items::meal_plan_item[]);

-- name: GetMealPlanItems :many
SELECT
	id,
	name,
	day,
	category,
	fats,
	carbohydrates,
	protein,
	calories,
	unit_type,
	brand,
	ingredient_id,
	amount,
	servings_name,
	servings_amount,
	meal_plan_id
FROM meal_plan_items
WHERE meal_plan_id = $1;


