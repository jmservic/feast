-- +goose Up

CREATE TABLE food_categories (
	id serial PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE ingredient_log (
	id bigserial PRIMARY KEY,
	date date,
	ingredients_id bigint REFERENCES ingredients (id) ON DELETE CASCADE,
	quantity decimal, --total amount
	servings_id smallint,
	user_id uuid REFERENCES users (id) ON DELETE CASCADE
);


-- +goose Down
