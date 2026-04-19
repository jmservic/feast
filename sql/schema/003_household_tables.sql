-- +goose Up
CREATE TABLE households (
	id uuid PRIMARY KEY,
	created_at timestamp NOT NULL, 
	updated_at timestamp NOT NULL,
	name text NOT NULL
);

CREATE TABLE household_roles (
	id serial PRIMARY KEY,
	name text NOT NULL
);

CREATE TABLE household_members ( 
	id uuid PRIMARY KEY,
	name text NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	role integer NOT NULL REFERENCES household_roles (id), 
	household_id uuid NOT NULL REFERENCES households (id) ON DELETE CASCADE,
	user_id uuid UNIQUE REFERENCES users (id) ON DELETE SET NULL -- what happens if we delete the owner of the household...
);

CREATE INDEX ON household_members ( household_id );
CREATE UNIQUE INDEX ON household_members ( household_id, role) WHERE role = 0;

CREATE TABLE household_invites (
	inviter_id uuid NOT NULL REFERENCES users (id),
	invitee_id uuid NOT NULL REFERENCES users (id),
	household_member_id uuid REFERENCES household_members (id),
	household_id uuid NOT NULL REFERENCES households (id),
	created_at timestamp NOT NULL,
	UNIQUE(invitee_id, household_id)
);

-- functions
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION can_create_member ( user_id uuid ) RETURNS boolean AS $$
DECLARE 
	user_role_id integer;
	manager_row_id integer;
	manager_name text := 'manager';

BEGIN
	SELECT role INTO user_role_id FROM  household_members WHERE user_id = user_id;
	IF NOT FOUND THEN
		RETURN false;
	ELSE
		SELECT id FROM household_roles INTO manager_row_id WHERE name = manager_name;
		IF NOT FOUND THEN
			RAISE EXCEPTION 'household role % not found', manager_name;
		END IF;

		RETURN user_role <= manager_row_id;
	END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION can_delete_member ( user_id uuid ) RETURNS boolean AS $$
BEGIN
	RETURN can_create_member(user_id);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- procedures spent the day working on angular...
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE create_household ( household_name text, v_user_id uuid ) AS $$
DECLARE
	member_name text := (SELECT name FROM users WHERE id = v_user_id); 
	v_household_id uuid;
BEGIN 
	-- Is the user already in a household?
	PERFORM household_id 
	FROM household_members 
	WHERE user_id = v_user_id;

	IF FOUND THEN 
		RAISE unique_violation USING DETAIL = 'User is already a part of a household.';
	END IF;
	
	INSERT INTO households ( id, created_at, updated_at, name )
    VALUES
	(
		gen_random_uuid(),
		NOW(), 
		NOW(),
		household_name
	)
	RETURNING id INTO v_household_id;

	INSERT INTO household_members ( id, name, created_at, updated_at, role, household_id, user_id )
	VALUES
	(
		gen_random_uuid(),
		member_name,
		NOW(),
		NOW(),
		1,
		v_household_id,
		v_user_id
	);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE delete_household ( v_household_id uuid ) AS $$
BEGIN
	DELETE FROM households WHERE id = v_household_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE user_delete_household ( v_user_id uuid, v_household_id uuid ) AS $$
DECLARE
	user_household_member_info record;
BEGIN
	SELECT role, household_id INTO user_household_member_info FROM household_members WHERE user_id = v_user_id;
	IF NOT FOUND THEN
		RAISE EXCEPTION '% user is not a part of a household', v_user_id
			USING ERRCODE = '42501';
	END IF;

	IF user_household_member_info.role <> 1 THEN
		RAISE EXCEPTION 'You do not have the required permissions to delete the household'
			USING ERRCODE = '42501';
	END IF;

	IF user_household_member_info.household_id <> v_household_id THEN
		RAISE EXCEPTION 'you cannot delete a household you''re not apart of'
			USING ERRCODE = '42501';
	END IF;

	CALL delete_household(v_household_id);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE user_update_household ( v_user_id uuid, v_household_id uuid, new_name text ) AS $$
DECLARE
	user_household_member_info record;
BEGIN
	SELECT role, household_id INTO user_household_member_info FROM household_members WHERE user_id = v_user_id;
	IF NOT FOUND THEN
		RAISE EXCEPTION '% user is not a part of a household', v_user_id 
			USING ERRCODE = '42501';
	END IF;

	IF user_household_member_info.role <> 1 THEN
		RAISE EXCEPTION 'You do not have the required permissions to update the household: role value %', user_household_member_info.role
			USING ERRCODE = '42501';
	END IF;

	IF user_household_member_info.household_id <> v_household_id THEN
		RAISE EXCEPTION 'you cannot update a household you''re not apart of'
			USING ERRCODE = '42501';
	END IF;

	UPDATE households SET name = new_name WHERE id = v_household_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin 
-- Need to check whether this is rolled back if the user is in another household.
CREATE OR REPLACE PROCEDURE user_create_household_member ( creator_id uuid, member_name text,  user_id uuid, household_id uuid ) AS $$
DECLARE
	user_household_member_info record;
	household_member_id uuid;
BEGIN
	IF NOT can_create_member(creator_id) THEN
		RAISE EXCEPTION '% does not have sufficient permission to create a member in the household', user_id;
	END IF;

	SELECT role, household_id INTO user_household_member_info FROM household_members WHERE user_id = creator_id;
	IF NOT FOUND THEN
		RAISE EXCEPTION '% user is not a part of a household', creator_id;
	END IF;

	IF user_household_member_info.household_id <> household_id THEN
		RAISE EXCEPTION 'you cannot create a member in another household';
	END IF;

	CALL create_household_member(member_name, NULL, household_id, household_member_id);
	IF user_id IS NOT NULL THEN
		CALL invite_user_to_household(creator_id, user_id, household_id, household_member_id);
	END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE create_household_member ( member_name text, user_id uuid, household_id uuid, OUT household_member_id uuid ) AS $$
BEGIN
	IF user_id IS NOT NULL THEN
		-- Is the user already in a household?
		SELECT household_id 
		FROM household_members 
		WHERE user_id = user_id;

		IF FOUND THEN 
			RAISE unique_violation USING DETAIL = 'User is already a part of a household.';
		END IF;
	END IF;

	INSERT INTO household_members ( id, name, created_at, updated_at, role, household_id, user_id )
	VALUES
	(
		gen_random_uuid(),
		member_name,
		NOW(),
		NOW(),
		4,
		household_id,
		user_id
	) RETURNING id INTO household_member_id;

END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE user_delete_household_member ( user_id uuid, household_member_id uuid ) AS $$
DECLARE
	user_household_member_info record;
	household_member_info record;
BEGIN
	
	IF NOT can_delete_member(user_id) THEN
		RAISE EXCEPTION '% does not have sufficient permission to delete a member from the household', user_id;
	END IF;

	SELECT role, household_id INTO user_household_member_info FROM household_members WHERE user_id = user_id;
	IF NOT FOUND THEN
		RAISE EXCEPTION '% user is not a part of a household', user_id;
	END IF;

	SELECT role, household_id, user_id INTO household_member_info FROM household_members WHERE id = household_member_id;
	IF NOT FOUND THEN 
		RAISE EXCEPTION '% member does not exist', household_member_id;
	END IF;

	IF user_household_member_info.household_id <> household_member_info.household_id THEN
		RAISE EXCEPTION 'you cannot delete a member from another household';
	END IF;

	IF household_member_info.user_id IS NOT NULL AND user_household_member_info.role >= household_member_info.role THEN
		RAISE EXCEPTION 'you cannot delete a household member that has an equal or higher role than you and is a user';
	END IF;

	CALL delete_household_member(household_member_id);
END;
$$ LANGUAGE plpgsql;
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

	-- remove the household_member_id
	DELETE FROM household_members WHERE id = household_member_id;

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
	
END;
$$ LANGUAGE plpgsql; 
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE invite_user_to_household ( inviter uuid, invitee uuid, household_id uuid, household_member_id uuid ) AS $$
DECLARE 
	household_member_household_id uuid;
	inviter_household_id uuid;
BEGIN
	--check if user is already a part of another household.
	SELECT id FROM household_members WHERE user_id = invitee;
	IF FOUND THEN
		RAISE EXCEPTION 'Invitee is already a part of a household';
	END IF;

	SELECT household_id FROM household_members WHERE user_id = inviter;
	IF inviter_household_id <> household_id THEN
		RAISE EXCEPTION 'Inviter cannot invite to a household that isn''t their own';
	END IF;

	--Check if the invitee actually has permission to invite.
	IF NOT can_create_member(inviter) THEN
		RAISE EXCEPTION '% does not have sufficient permission to invite to the household', inviter;
	END IF;


	IF household_member_id IS NOT NULL THEN
		SELECT household_id INTO household_member_household_id 
		FROM household_members WHERE id = household_member_id;
		
		IF household_member_household_id <> household_id THEN
			RAISE EXCEPTION 'household member isn''t apart of the same household';
		END IF;
	END IF;

	INSERT INTO household_invites ( inviter_id, invitee_id, household_member_id, household_id, created_at )
	VALUES (
		inviter,
		invitee,
		household_member_id,
		household_id,
		NOW()
	);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE accept_household_invite ( invitee_id uuid, household_id uuid ) AS $$
DECLARE
	invite_info record;
	member_name text;
BEGIN
	SELECT * INTO invite_info FROM household_invites WHERE invitee_id = invitee_id AND household_id = household_id;
	IF NOT FOUND THEN
		RAISE EXCEPTION 'No invite found';
	END IF;

	IF invite_info.household_member_id IS NOT NULL
	THEN
		UPDATE household_members
			SET user_id = invitee_id
		WHERE id = invite_info.household_member_id;
	ELSE 
		SELECT name INTO member_name FROM users WHERE id = invitee_id;
		CALL create_household_member(member_name, invitee_id, household_id);
	END IF;

	DELETE FROM household_invites WHERE invitee_id = invitee_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- initial data
INSERT INTO household_roles (name) 
VALUES
	('head'),
	('administrator'),
	('manager'),
	('member');

-- +goose Down
DROP TABLE IF EXISTS household_invites;
DROP TABLE IF EXISTS household_members;
DROP TABLE IF EXISTS household_roles;
DROP TABLE IF EXISTS households;
DROP FUNCTION IF EXISTS can_create_member;
DROP FUNCTION IF EXISTS can_delete_member;
DROP PROCEDURE IF EXISTS create_household;
DROP PROCEDURE IF EXISTS delete_household;
DROP PROCEDURE IF EXISTS user_delete_household;
DROP PROCEDURE IF EXISTS user_update_household;
DROP PROCEDURE IF EXISTS create_household_member;
DROP PROCEDURE IF EXISTS user_create_household_member; 
DROP PROCEDURE IF EXISTS delete_household_member;
DROP PROCEDURE IF EXISTS user_delete_household_member;
DROP PROCEDURE IF EXISTS invite_user_to_household;
DROP PROCEDURE IF EXISTS accept_household_invite;
-- Need a leave household 
