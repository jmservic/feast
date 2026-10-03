-- name: CreateIngredient :one
INSERT INTO ingredients (name, fats, carbohydrates, protein, calories, unit_type, brand, user_id)
VALUES(
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8
) returning *;

-- name: CreateIngredientServings :exec
CALL create_ingredient_servings(@ingredient_id, @servings::serving_info[]);

-- name: GetIngredientsServingsById :many
SELECT 
	name,
	quantity,
	id
FROM ingredient_servings
WHERE ingredient_id = $1;

