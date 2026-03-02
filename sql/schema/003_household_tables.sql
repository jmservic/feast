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
	household_id UUID NOT NULL REFERENCES households (id) ON DELETE CASCADE,
	user_id UUID UNIQUE REFERENCES users (id)
);

CREATE TABLE household_invites (
	inviter_id UUID REFERENCES users (id),
	invitee_id UUID REFERENCES users (id),
	household_member_id UUID REFERENCES household_members (id),
	created_at TIMESTAMP NOT NULL,
);


-- procedures
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE create_household (name TEXT, userId UUID, out household_id UUID) AS $$
DECLARE
	member_name TEXT := (SELECT name FROM users WHERE id = userId); 
BEGIN 
	-- Is the user already in a household?
	SELECT household_id INTO household_id
	FROM household_members 
	WHERE user_id = userId;

	IF FOUND THEN 
		RAISE unique_violation USING DETAIL = 'User is already a part of a household.';
	END IF;
	
	household_id := gen_random_uuid();
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
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE create_household_member (member_name TEXT, userId UUID, household_id UUID, out household_member_id UUID) AS $$
BEGIN
	IF userId IS NOT NULL THEN
		-- Is the user already in a household?
		SELECT household_id 
		FROM household_members 
		WHERE user_id = userId;

		IF FOUND THEN 
			RAISE unique_violation USING DETAIL = 'User is already a part of a household.';
		END IF;
	END IF;

	household_member_id := gen_random_uuid();

	INSERT INTO household_members ( id, name, created_at, updated_at, role, household_id, user_id )
	VALUES
	(
		household_member_id,
		member_name,
		NOW(),
		NOW(),
		4,
		household_id,
		userId
	);

END;
$$ LANGUAGE plpgsql
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE delete_household_member ( household_member_id uuid ) AS $$
DECLARE
	member_info record;
	next_highest_member uuid;
	household_member_count integer;
BEGIN
	-- get the household_member info
	SELECT role, household_id 
	INTO member_info 
	FROM household_members
	WHERE id = household_member_id;

	IF NOT FOUND THEN
		RAISE EXCEPTION 'household member % not found', household_member_id;
	END IF;

	-- if there is no other members, delete the household -- Done
	SELECT COUNT(id) INTO household_member_count FROM household_members
	WHERE household_id = member_info.household_id AND user_id IS NOT NULL;

	IF household_member_count = 1 THEN -- This will need to change as household deletion gets more complicated
		DELETE FROM households WHERE id = member_info.household_id;
		RETURN;
	END IF;

	-- Check if the user is the head
	IF member_info.role = 0 THEN
		-- if the user is the head, find the next highest role
		SELECT id INTO STRICT next_highest_member FROM household_members
		WHERE id <> household_member_id
		ORDER BY role ASC;

		--  Promote the next highest role member to head
		UPDATE household_members
		SET role = 0
		WHERE id = next_highest_member;
	END IF;
	
	-- remove the household_member_id
	DELETE FROM household_members WHERE id = household_member_id;
END;
$$ LANGUAGE plpgsql 
-- +goose StatementEnd

-- Need Insert, invite and delete procedures

-- initial data
INSERT INTO household_roles (name) 
VALUES
	('head'),
	('administrator'),
	('manager'),
	('member');

-- +goose Down
DROP TABLE household_members;
DROP TABLE household_roles;
DROP TABLE households;
DROP PROCEDURE create_household;
DROP PROCEDURE create_household_member;
