-- name: CreateFoodItem :one
INSERT INTO food_log (name, date, category, fats, carbohydrates, protein, calories, unit_type, brand, ingredients_id, quantity, servings_name, servings_quantity, user_id)
VALUES (
$1, 
$2,
$3,
$4,
$5,
$6,
$7,
$8,
$9,
$10,
$11,
$12,
$13,
$14
)
RETURNING *;

-- name: GetFoodItemById :one
SELECT * FROM food_log
WHERE id = $1;

-- name: GetUserFoodItemsByRange :many
SELECT *
FROM food_log
WHERE date BETWEEN @start_date and @end_date AND user_id = @user_id;

-- name: UpdateFoodItem :one
UPDATE food_log
SET
	date = $2,
	category = $3,
	fats = $4,
	carbohydrates = $5,
	protein = $6,
	calories = $7,
	quantity = $8,
	servings_name = $9,
	servings_quantity = $10
WHERE id = $1 
RETURNING *;

-- name: DeleteFoodItem :exec
DELETE FROM food_log
WHERE id = $1;
