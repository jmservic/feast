-- +goose Up
CREATE TABLE households (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL, 
	name TEXT NOT NULL
);

-- +goose Down
DROP TABLE households;
