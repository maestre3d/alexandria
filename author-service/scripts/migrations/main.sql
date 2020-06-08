/******************************
**	File:   main.sql
**	Name:	Database migrations scripts
**	Desc:	Main database migrations scripts for Author microservice
**	Auth:	Alonso R
**	Lic:	MIT	
**	Date:	2020-06-1
*******************************/

CREATE DATABASE 'alexandria/author';
SET search_path TO 'alexandria/author';
CREATE SCHEMA IF NOT EXISTS alexa1;

CREATE TYPE alexa1.ownership_enum AS ENUM(
    'public',
    'private'
);

CREATE TABLE IF NOT EXISTS alexa1.author(
	id 				bigserial NOT NULL UNIQUE,
	external_id 	varchar(128) NOT NULL UNIQUE,
	first_name 		varchar(255) NOT NULL,
	last_name 		varchar(255) NOT NULL,
	display_name 	varchar(255) NOT NULL UNIQUE,
	ownership_type  alexa1.ownership_enum NOT NULL DEFAULT 'private',
	create_time 	timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	update_time 	timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	delete_time 	timestamp DEFAULT NULL,
	active   		bool DEFAULT FALSE,
	verified        bool DEFAULT FALSE,
	picture         text DEFAULT NULL,
	total_views     bigint DEFAULT 0,
	PRIMARY KEY(id, external_id)
);

CREATE TYPE alexa1.role_enum AS ENUM (
	'root',
	'admin',
	'contrib'
);

CREATE TABLE IF NOT EXISTS alexa1.author_user(
    fk_author   varchar(128) NOT NULL REFERENCES alexa1.author(external_id) ON DELETE CASCADE,
    "user"      varchar(128) NOT NULL,
	role_type	alexa1.role_enum NOT NULL DEFAULT 'contrib',
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE PROCEDURE alexa1.create_author(_external_id varchar(128), _first_name varchar(255), _last_name varchar(255), _display_name varchar(255),
	_ownership_type alexa1.ownership_enum, _user varchar(128), _role alexa1.role_enum)
LANGUAGE SQL
AS $$
	INSERT INTO alexa1.author(external_id, first_name, last_name, display_name, ownership_type) VALUES (
		_external_id, _first_name, _last_name, _display_name, _ownership_type);
	INSERT INTO alexa1.author_user(fk_author, "user", role_type) VALUES(_external_id, _user, _role);
$$;

CREATE PROCEDURE alexa1.update_author(_external_id varchar(128), _first_name varchar(255), _last_name varchar(255), _display_name varchar(255),
	_ownership_type alexa1.ownership_enum, _total_views bigint)
LANGUAGE SQL
AS $$
    UPDATE alexa1.author SET first_name = _first_name, last_name = _last_name, display_name = _display_name, ownership_type = _ownership_type,
    update_time = CURRENT_TIMESTAMP, total_views = _total_views WHERE external_id = _external_id AND active = true;
$$;

CREATE PROCEDURE alexa1.add_user_author(_external_id varchar(128), _user varchar(128), _role alexa1.role_enum)
LANGUAGE SQL
AS $$
    INSERT INTO alexa1.author_user(fk_author, "user", role_type) VALUES (_external_id, _user, _role)
$$;

CALL alexa1.create_author('WS34YXqskjGZIdyq', 'Elon', 'Musk', 'Elon Musk', 'public', 'd1d4469b-8502-4792-a1e7-13391aa67f2c', 'root');
CALL alexa1.create_author('DmdienU5Ll-uYj4O', 'Moris', 'Dieck', 'Moris Dieck', 'private', 'd1d4469b-8502-4792-a1e7-13391aa67f2c', 'root');

-- Get users from author user pool
SELECT "user", role_type FROM alexa1.author_user WHERE fk_author = 'WS34YXqskjGZIdyq';

-- Get all authors from an owner, not necessarily root
SELECT * FROM alexa1.author WHERE external_id IN (SELECT fk_author FROM alexa1.author_user WHERE "user" = 'd1d4469b-8502-4792-a1e7-13391aa67f2c') AND 
active = FALSE ORDER BY update_time DESC FETCH FIRST 15 ROWS ONLY;

UPDATE alexa1.author SET total_views = 1998540 WHERE external_id = 'WS34YXqskjGZIdyq';
UPDATE alexa1.author SET total_views = 1545 WHERE external_id = 'DmdienU5Ll-uYj4O';

-- Get all authors from the most popular to the worst, using token_id (external_id)
SELECT * FROM alexa1.author WHERE 
total_views <= (SELECT total_views FROM alexa1.author WHERE external_id = 'WS34YXqskjGZIdyq' AND active = FALSE) AND 
active = FALSE ORDER BY total_views DESC FETCH FIRST 10 ROWS ONLY;


