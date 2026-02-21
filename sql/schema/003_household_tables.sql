-- +goose Up
CREATE TABLE households (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL, 
	updated_at TIMESTAMP NOT NULL,
	name TEXT NOT NULL
);

CREATE TABLE household_roles (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE household_members (
	id UUID PRIMARY KEY,
	name TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	role INTEGER REFERENCES household_roles (id), 
	household_id UUID REFERENCES households (id),
	user_id UUID REFERENCES users (id)
);


-- procedures
-- +goose StatementBegin
CREATE PROCEDURE create_household (name TEXT, userId UUID) AS $$
DECLARE
	household_id UUID := gen_random_uuid();
	member_name TEXT := (SELECT name FROM users WHERE id = userId); 
BEGIN 

	INSERT INTO households ( id, created_at, updated_at, name )
    VALUES
	(
		household_id,
		NOW(), 
		NOW(),
		name
	);

	INSERT INTO household_members ( id, name, created_at, updated_at, role, household_id, user_id )
	VALUES
	(
		gen_random_uuid(),
		member_name,
		NOW(),
		NOW(),
		1,
		household_id,
		userId
	);
	
	--SELECT * FROM households WHERE id = household_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- initial data
INSERT INTO household_roles (name) 
VALUES
	('administrator'),
	('manager'),
	('member');

-- +goose Down
DROP TABLE household_members;
DROP TABLE household_roles;
DROP TABLE households;
DROP PROCEDURE create_household;
