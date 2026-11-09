-- +goose Up
CREATE TABLE food_categories (
	id serial PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE food_log(
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	name text NOT NULL,
	date date NOT NULL,
	category int REFERENCES food_categories (id) ON DELETE SET NULL,
	fats numeric(26, 16) NOT NULL,
	carbohydrates numeric(26, 16) NOT NULL,
	protein numeric(26, 16) NOT NULL,
	calories numeric(26, 16) NOT NULL,
	unit_type boolean NOT NULL, -- weight vs volume
	brand text,
	ingredients_id uuid REFERENCES ingredients (id) ON DELETE SET NULL,
	quantity decimal NOT NULL, --total amount
	servings_name text NOT NULL,
	servings_quantity decimal NOT NULL,
	user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	created_at timestamp WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ON food_log( user_id );
CREATE INDEX ON food_log ( date );

INSERT INTO food_categories (name)
VALUES ('Breakfast'), ('Lunch'), ('Dinner'), ('Snacks');

-- +goose Down
DROP TABLE IF EXISTS food_log;
DROP TABLE IF EXISTS food_categories;

