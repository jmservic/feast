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
	ingredient_id uuid REFERENCES ingredients (id) ON DELETE SET NULL,
	amount decimal NOT NULL, --total amount
	servings_name text NOT NULL,
	servings_amount decimal NOT NULL,
	meal_plan_id uuid NOT NULL REFERENCES meal_plans (id) ON DELETE CASCADE
);

CREATE TYPE meal_plan_item AS (
	id uuid,
	name text,
	day int,
	category int,
	fats numeric(26, 16),
	carbohydrates numeric(26, 16),
	protein numeric(26, 16),
	calories numeric(26, 16),
	unit_type boolean,
	brand text,
	ingredient_id uuid,
	servings_name text,
	servings_amount decimal
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
	ingredient_id uuid REFERENCES ingredients (id) ON DELETE SET NULL,
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

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE create_update_meal_plan_items ( v_meal_plan_id uuid, v_items meal_plan_item[] ) AS $$
BEGIN
	MERGE INTO meal_plan_items mpi
		USING unnest(v_items) vi ON ( vi.id AND v_meal_plan_id = mpi.meal_plan_id )
		WHEN MATCHED THEN
			UPDATE 
				SET name = vi.name,
					day = vi.day,
					category = vi.category,
					fats = vi.fats,
					carbohydrates = vi.carbohydrates,
					protein = vi.protein,
					calories = vi.calories,
					unit_type = vi.unit_type,
					brand = vi.brand,
					ingredient_id = vi.ingredient_id,
					servings_name = vi.servings_name,
					servings_amount = vi.servings_amount
		WHEN NOT MATCHED THEN
			INSERT ( name, day, category, fats, carbohydrates, protein, calories, unit_type, brand, ingredient_id, servings_name, servings_amount, meal_plan_id ) VALUES ( 
				vi.name,
				vi.day,
				vi.category,
				vi.fats,
				vi.carbohydrates,
				vi.protein,
				vi.calories,
				vi.unit_type,
				vi.brand,
				vi.ingredient_id,
				vi.servings_name,
				vi.servings_amount,
				v_meal_plan_id
			);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS meal_plan_items;
DROP TABLE IF EXISTS meal_plan;
DROP TYPE IF EXISTS meal_plan_item;
DROP PROCEDURE IF EXISTS create_update_meal_plan_items;
