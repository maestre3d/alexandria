/******************************
**	File:   main.sql
**	Name:	Database migration script
**	Desc:	Main database migration script for Resource microservice
**	Auth:	Alonso R
**	Lic:	MIT	
**	Date:	2020-04-06
*******************************/

-- Media Entity
CREATE TYPE MEDIA_ENUM AS ENUM(
	'MEDIA_BOOK',
	'MEDIA_DOC',
	'MEDIA_PODCAST',
	'MEDIA_VIDEO'
);
CREATE TABLE IF NOT EXISTS MEDIA(
	MEDIA_ID	 	BIGSERIAL NOT NULL PRIMARY KEY,
    EXTERNAL_ID 	UUID NOT NULL UNIQUE,
    TITLE 			VARCHAR(255) NOT NULL UNIQUE,
    DISPLAY_NAME 	VARCHAR(100) NOT NULL,
    DESCRIPTION 	TEXT DEFAULT NULL,
    USER_ID 		UUID NOT NULL,
    AUTHOR_ID 		UUID NOT NULL,
    PUBLISH_DATE 	DATE NOT NULL DEFAULT CURRENT_DATE,
	MEDIA_TYPE		MEDIA_ENUM NOT NULL DEFAULT 'MEDIA_BOOK',
    CREATE_TIME 	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UPDATE_TIME 	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    DELETE_TIME 	TIMESTAMP DEFAULT NULL,
	METADATA 		TEXT DEFAULT NULL,
    DELETED 		BOOLEAN DEFAULT FALSE
);

-- Insert Media entity mock data
INSERT INTO MEDIA(EXTERNAL_ID, TITLE, DISPLAY_NAME, DESCRIPTION, USER_ID, AUTHOR_ID)
VALUES (
	'fb5c903e-22bc-4799-9a48-a74188b114fa',
	'Building Microservices: Designing Fine-Grained Systems',
	'Building Microservices',
	'In this media, Sam Newman explains the whole process of microservices developing. Microservices are widely used and should be the first choice for API developing.',
	'af06cc79-1d66-4fb8-820a-07d7d75a9ada',
	'933f76f5-f17c-46d5-9924-f6b9445be87c'
), 
(
	'0120ebf7-91fd-4b16-9ed1-d4c58c4b8795',
	'Design, Build, Ship: Faster, Safer Software Delivery',
	'Design, Build and Ship',
	'What''s the best way to get code from your laptop into a production environment? With this highly actionable guide, architects, developers, engineers, and others in the IT space will learn everything.',
	'af06cc79-1d66-4fb8-820a-07d7d75a9ada',
	'933f76f5-f17c-46d5-9924-f6b9445be87c'
), (
	'e15162e7-d32f-49a9-8297-f8757764c80a',
	'I Heard You Paint Houses: Frank Sheeran "The Irishman" and Closing the Case on Jimmy Hoffa',
	'I Heard You Paint Houses',
	'I Heard You Paint Houses: Frank "The Irishman" Sheeran and Closing the Case on Jimmy Hoffa is a 2004 work of narrative nonfiction written by former homicide prosecutor, investigator and defense attorney Charles Brandt that chronicles the life of Frank Sheeran, an alleged mafia hitman who confesses the crimes he committed working for the Bufalino crime family.',
	'2d42f63c-d76f-4dd4-8d89-ad95b0706b08',
	'862d2c11-96c3-4893-9c24-48964e7487ce'
), (
	'6d3ac6c1-a503-4638-9b7d-503f36b01f34',
	'Monolith to Microservices: Evolutionary Patterns to Transform Your Monolith',
	'Monolith to Microservices',
	'How do you detangle a monolithic system and migrate it to a microservice architecture? How do you do it while maintaining business-as-usual?',
	'af06cc79-1d66-4fb8-820a-07d7d75a9ada',
	'933f76f5-f17c-46d5-9924-f6b9445be87c'
);

-- Data querying
SELECT * FROM MEDIA;
