-- +goose Up

CREATE TABLE meal_plans (
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	name text NOT NULL,
	description text NOT NULL,
	user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE meal_plan_items (
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	name text NOT NULL,
	day int NOT NULL,
	category int REFERENCES food_categories (id) ON DELETE SET NULL,
	fats numeric(26, 16) NOT NULL,
	carbohydrates numeric(26, 16) NOT NULL,
	protein numeric(26, 16) NOT NULL,
	calories numeric(26, 16) NOT NULL,
	unit_type boolean NOT NULL, -- weight vs volume
	brand text,
	ingredients_id uuid REFERENCES ingredients (id) ON DELETE SET NULL,
	amount decimal NOT NULL, --total amount
	servings_name text NOT NULL,
	servings_amount decimal NOT NULL,
	meal_plan_id uuid NOT NULL REFERENCES meal_plans (id) ON DELETE CASCADE
);

CREATE TABLE meal_planner (
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
	amount decimal NOT NULL, --total amount
	servings_name text NOT NULL,
	servings_amount decimal NOT NULL,
	user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	meal_plan_instance uuid,
	food_log_id uuid REFERENCES food_log (id) ON DELETE SET NULL
);

CREATE INDEX ON meal_plans ( user_id );
CREATE INDEX ON meal_plan_items ( meal_plan_id);
CREATE INDEX ON meal_planner ( user_id);
CREATE INDEX ON meal_planner ( date );

-- +goose Down
DROP TABLE IF EXISTS meal_plan_items;
DROP TABLE IF EXISTS meal_plan;
