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
) RETURNING *;

-- name: CreateIngredientServings :exec
CALL create_ingredient_servings(@ingredient_id, @servings::serving_info[]);

-- name: GetIngredientsServingsById :many
SELECT 
	name,
	quantity,
	id
FROM ingredient_servings
WHERE ingredient_id = $1;

-- name: GetIngredientsServingsByIds :many
SELECT 
	name,
	quantity,
	id,
	ingredient_id
FROM ingredient_servings
WHERE ingredient_id = ANY(@ingredient_ids::int[]);

-- name: DeleteIngredientServings :exec
DELETE FROM ingredient_servings
WHERE ingredient_id = @ingredient_id AND id = ANY(@ids::smallint[]);

-- name: GetIngredientById :one
SELECT 
	id,
	name,
	fats,
	carbohydrates,
	protein,
	calories,
	unit_type,
	brand,
	user_id
FROM ingredients
WHERE id = $1;

-- name: GetIngredientsByQuery :many
SELECT
	id,
	name,
	fats,
	carbohydrates,
	protein,
	calories,
	unit_type,
	brand,
	user_id,
	ts_rank(search_vector, query) AS rank
FROM ingredients, to_tsquery(@query) query
WHERE search_vector @@ query AND user_id = @user_id
ORDER BY rank DESC
LIMIT 50;

-- name: UpdateIngredient :one
UPDATE ingredients
SET
	name = $2,
	fats = $3,
	carbohydrates = $4,
	protein = $5,
	calories = $6,
	unit_type = $7,
	brand = $8
WHERE id = $1
RETURNING *;
