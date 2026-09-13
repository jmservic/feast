CREATE TABLE foods (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
	fats decimal NOT NULL,
	carbohydrates decimal NOT NULL,
	protein decimal NOT NULL,
	brand text
);

CREATE TABLE food_servings (
	food_id bigint NOT NULL REFERENCES foods (id) ON DELETE CASCASDE,
	quantity decimal NOT NULL,
	unit_id int NOT NULL REFERENCES units (id) ON DELETE CASCADE,
	UNIQUE(food_id, unit_id)
);

CREATE TABLE units (
	id serial PRIMARY KEY,
	name text NOT NULL,
	type bit NOT NULL, -- weight vs volume
);
