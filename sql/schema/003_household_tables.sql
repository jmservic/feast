-- +goose Up
CREATE TABLE households (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL, 
	name TEXT NOT NULL
);
CREATE TABLE household_members (
	id UUID PRIMARY KEY,
	name TEXT,
	role INT, -- Foreign key
	household_id UUID, --Foreign key
	user_id UUID --Foreign key
);

-- +goose Down
DROP TABLE households;
