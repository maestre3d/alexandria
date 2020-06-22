/******************************
**	File:   mainqueries.sql
**	Name:	Database migrations scripts
**	Desc:	Main database migrations scripts for media microservice
**	Auth:	Alonso R
**	Lic:	MIT	
**	Date:	2020-04-06
*******************************/

CREATE DATABASE 'alexandria/media';
SET search_path TO 'alexandria/media';

CREATE SCHEMA IF NOT EXISTS alexa1;

-- Media Entity
CREATE TYPE alexa1.media_enum AS ENUM(
	'MEDIA_BOOK',
	'MEDIA_DOC',
	'MEDIA_PODCAST',
	'MEDIA_VIDEO'
);

CREATE TYPE alexa1.state_enum AS ENUM(
    'STATUS_PENDING',
    'STATUS_DONE'
);

-- public=available for everyone, private=only available for owners, unlisted=available for people with url
CREATE TYPE alexa1.visibility_enum AS ENUM(
    'public',
    'private',
    'unlisted'
);

-- Using ISO 639-1 Language code
CREATE TABLE IF NOT EXISTS alexa1.media(
	id	 	        bigserial NOT NULL,
    external_id 	varchar(128) NOT NULL UNIQUE,
    title 			varchar(255) NOT NULL UNIQUE,
    display_name 	varchar(128) NOT NULL,
    description 	text DEFAULT NULL,
    language_code   varchar(5) NOT NULL DEFAULT 'en',
    publisher_id 	varchar(128) NOT NULL,
    author_id 		varchar(128) NOT NULL,
    publish_date 	date NOT NULL DEFAULT CURRENT_DATE,
	media_type		alexa1.media_enum NOT NULL DEFAULT 'MEDIA_BOOK',
	create_time 	timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	update_time 	timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	delete_time 	timestamp DEFAULT NULL,
	active   		bool DEFAULT TRUE,
	content_url     text DEFAULT NULL,
	total_views     bigint DEFAULT 0,
	status          alexa1.state_enum NOT NULL DEFAULT 'STATUS_PENDING',
	PRIMARY KEY(id, external_id)
);

-- Insert Media entity mock persistence
INSERT INTO alexa1.media(external_id, title, display_name, description, publisher_id, author_id)
VALUES (
	'Bg7-4rPtC-Kl2fGh',
	'Building Microservices: Designing Fine-Grained Systems',
	'Building Microservices',
	'In this book, Sam Newman explains the whole process of microservices developing. Microservices are widely used and should be the first choice for API developing.',
	'af06cc79-1d66-4fb8-820a-07d7d75a9ada',
	'WS34YXqskjGZIdyq'
), 
(
	'UPj6SSMvaKIuXwnY',
	'Design, Build, Ship: Faster, Safer Software Delivery',
	'Design, Build and Ship',
	'What''s the best way to get code from your laptop into a production environment? With this highly actionable guide, architects, developers, engineers, and others in the IT space will learn everything.',
	'af06cc79-1d66-4fb8-820a-07d7d75a9ada',
	'WS34YXqskjGZIdyq'
), (
	'jZeM577fgcSrulKc',
	'I Heard You Paint Houses: Frank Sheeran "The Irishman" and Closing the Case on Jimmy Hoffa',
	'I Heard You Paint Houses',
	'I Heard You Paint Houses: Frank "The Irishman" Sheeran and Closing the Case on Jimmy Hoffa is a 2004 work of narrative nonfiction written by former homicide prosecutor, investigator and defense attorney Charles Brandt that chronicles the life of Frank Sheeran, an alleged mafia hitman who confesses the crimes he committed working for the Bufalino crime family.',
	'2d42f63c-d76f-4dd4-8d89-ad95b0706b08',
	'DmdienU5Ll-uYj4O'
), (
	'WfSPP636sByUECgl',
	'Monolith to Microservices: Evolutionary Patterns to Transform Your Monolith',
	'Monolith to Microservices',
	'How do you detangle a monolithic system and migrate it to a microservice architecture? How do you do it while maintaining business-as-usual?',
	'af06cc79-1d66-4fb8-820a-07d7d75a9ada',
	'DmdienU5Ll-uYj4O'
);

-- Data querying
SELECT * FROM alexa1.media;

