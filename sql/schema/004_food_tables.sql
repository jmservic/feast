-- +goose Up
CREATE TABLE ingredients (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
	fats numeric(26, 16) NOT NULL,
	carbohydrates numeric(26, 16) NOT NULL,
	protein numeric(26, 16) NOT NULL,
	calories numeric(26, 16) NOT NULL,
	unit_type boolean NOT NULL, -- weight vs volume
	brand text,
	user_id uuid REFERENCES users (id) ON DELETE CASCADE,
	search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || coalesce(brand, ''))) STORED
);

CREATE TABLE ingredient_servings (
	name text NOT NULL,
	quantity decimal NOT NULL,
	ingredient_id bigint NOT NULL REFERENCES ingredients (id) ON DELETE CASCADE,
	id smallint NOT NULL,
	UNIQUE(ingredient_id, id) 
);

CREATE VIEW v_foods AS 
SELECT
	i.id,
	i.name,
	fats,
	carbohydrates,
	protein,
	calories,
	unit_type,
	brand,
	user_id,
	search_vector
FROM ingredients i;

CREATE VIEW v_servings AS 
SELECT
	name,
	quantity,
	ingredient_id,
	id
FROM ingredient_servings;

-- types
CREATE TYPE serving_info AS (
	id smallint,
	name text,
	quantity decimal
);

-- procedures
-- +goose StatementBegin
-- Update this to create and update. We need an integer id for the servings, so they can be targeted. 
-- The ingredient_servings Unique will then be the ingredient_id and id
CREATE OR REPLACE PROCEDURE create_ingredient_servings ( v_ingredient_id bigint, v_servings serving_info[] ) AS $$
DECLARE
	v_serving_info serving_info;
	serving_id_start smallint;
	servings_len integer;
BEGIN
	SELECT array_length(v_servings, 1) INTO servings_len;

	IF servings_len < 1 THEN
		RAISE EXCEPTION 'There must be atleast one servings passed in';
	END IF;

	--Get the next id for the ingredient
	SELECT MAX(id) + 1 INTO serving_id_start FROM ingredient_servings 
	WHERE ingredient_id = v_ingredient_id;

	IF serving_id_start IS NULL THEN
		serving_id_start = 1;
	END IF;

	--Create the sequence and merge the results.
	EXECUTE format(
	'CREATE TEMP SEQUENCE v_serving_id_sequence AS smallint START WITH %s; '
	' MERGE INTO ingredient_servings iser '
		'USING unnest($1) vs ON (vs.id = iser.id AND $2 = iser.ingredient_id) '
		'WHEN MATCHED THEN '
			'UPDATE ' 
			'	SET name = vs.name, ' 
					' quantity = vs.quantity '
		'WHEN NOT MATCHED THEN '
			'INSERT (name, quantity, ingredient_id, id) VALUES ( '
				'vs.name, '
				'vs.quantity, '
				'$2, '
				'nextval(''v_serving_id_sequence'') '
			'); '
	'DROP SEQUENCE v_serving_id_sequence;', serving_id_start) USING v_servings, v_ingredient_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE INDEX ON ingredients (id);
CREATE INDEX ON ingredient_servings (ingredient_id);
CREATE INDEX ON ingredients USING GIN (search_vector);
CREATE INDEX ON ingredients (user_id);

-- +goose Down
DROP VIEW IF EXISTS v_foods;
DROP VIEW IF EXISTS v_servings;
DROP TABLE IF EXISTS ingredient_servings;
DROP TABLE IF EXISTS ingredients;
DROP PROCEDURE IF EXISTS create_ingredient_servings;
DROP TYPE IF EXISTS serving_info;
