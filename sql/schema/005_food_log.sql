-- +goose Up

CREATE TABLE food_categories (
	id serial PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE ingredient_log (
	id bigserial PRIMARY KEY,
	category int REFERENCES food_categories (id) ON DELETE SET NULL,
	date date,
	ingredients_id bigint REFERENCES ingredients (id) ON DELETE CASCADE,
	quantity decimal, --total amount
	servings_id smallint,
	user_id uuid REFERENCES users (id) ON DELETE CASCADE
);

INSERT INTO food_categories (name)
VALUES ('Breakfast'), ('Lunch'), ('Dinner'), ('Snacks');

-- +goose Down
DROP TABLE IF EXISTS food_categories;
DROP TABLE IF EXISTS ingredient_log;

