-- +goose Up
CREATE TABLE households (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL, 
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
