-- name: CreateFoodItem :one
INSERT INTO food_log (name, date, category, fats, carbohydrates, protein, calories, unit_type, brand, ingredients_id, amount, servings_name, servings_amount, user_id)
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
WHERE date BETWEEN @start_date and @end_date AND user_id = @user_id
ORDER BY date ASC;

-- name: UpdateFoodItem :one
UPDATE food_log
SET
	date = $2,
	category = $3,
	fats = $4,
	carbohydrates = $5,
	protein = $6,
	calories = $7,
	amount = $8,
	servings_name = $9,
	servings_amount = $10
WHERE id = $1 
RETURNING *;

-- name: DeleteFoodItem :exec
DELETE FROM food_log
WHERE id = $1;
