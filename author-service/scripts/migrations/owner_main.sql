/******************************
**	File:   main.sql
**	Name:	Database migrations scripts
**	Desc:	Main database migrations scripts for Owner microservice
**	Auth:	Alonso R
**	Lic:	MIT
**	Date:	2020-06-12
*******************************/

CREATE DATABASE 'alexandria/owner';
SET search_path TO 'alexandria/owner';
CREATE SCHEMA IF NOT EXISTS alexa1;

CREATE TYPE alexa1.role_enum AS ENUM (
	'admin',
	'contrib'
);

CREATE TABLE IF NOT EXISTS alexa1.owners(
    author_id   varchar(128) NOT NULL,
    owner_id    varchar(128) NOT NULL,
	role_type	alexa1.role_enum NOT NULL DEFAULT 'contrib',
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);