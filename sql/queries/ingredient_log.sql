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
