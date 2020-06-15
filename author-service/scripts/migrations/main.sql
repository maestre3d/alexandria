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

CREATE TYPE alexa1.state_enum AS ENUM(
    'STATUS_PENDING',
    'STATUS_DONE'
);

CREATE TABLE IF NOT EXISTS alexa1.author(
	id 				bigserial NOT NULL UNIQUE,
	external_id 	varchar(128) NOT NULL UNIQUE,
	first_name 		varchar(255) NOT NULL,
	last_name 		varchar(255) NOT NULL,
	display_name 	varchar(255) NOT NULL UNIQUE,
	owner_id        varchar(128) NOT NULL,
	ownership_type  alexa1.ownership_enum NOT NULL DEFAULT 'private',
	create_time 	timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	update_time 	timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	delete_time 	timestamp DEFAULT NULL,
	active   		bool DEFAULT TRUE,
	verified        bool DEFAULT FALSE,
	picture         text DEFAULT NULL,
	total_views     bigint DEFAULT 0,
	status          alexa1.state_enum NOT NULL DEFAULT 'STATUS_PENDING',
	PRIMARY KEY(id, external_id)
);

CREATE PROCEDURE alexa1.create_author(_external_id varchar(128), _first_name varchar(255), _last_name varchar(255), _display_name varchar(255),
	_ownership_type alexa1.ownership_enum, _owner varchar(128))
LANGUAGE SQL
AS $$
	INSERT INTO alexa1.author(external_id, first_name, last_name, display_name, ownership_type, owner_id) VALUES (
		_external_id, _first_name, _last_name, _display_name, _ownership_type, _owner);
$$;

CALL alexa1.create_author('WS34YXqskjGZIdyq', 'Elon', 'Musk', 'Elon Musk', 'public', 'd1d4469b-8502-4792-a1e7-13391aa67f2c');
CALL alexa1.create_author('DmdienU5Ll-uYj4O', 'Moris', 'Dieck', 'Moris Dieck', 'private', 'd1d4469b-8502-4792-a1e7-13391aa67f2c');

UPDATE alexa1.author SET total_views = 1998540, status = 'STATUS_DONE' WHERE external_id = 'WS34YXqskjGZIdyq';
UPDATE alexa1.author SET total_views = 1545, status = 'STATUS_DONE' WHERE external_id = 'DmdienU5Ll-uYj4O';

-- Get all authors from the most popular to the worst, using token_id (external_id)
SELECT * FROM alexa1.author WHERE 
total_views <= (SELECT total_views FROM alexa1.author WHERE external_id = 'WS34YXqskjGZIdyq' AND active = TRUE) AND 
active = TRUE ORDER BY total_views DESC FETCH FIRST 10 ROWS ONLY;


