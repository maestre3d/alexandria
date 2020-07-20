-- Change replication strategy for NetworkTopology when deploying to production
CREATE KEYSPACE IF NOT EXISTS alexa1
    WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
    AND durable_writes = false;

USE alexa1;

DROP TABLE IF EXISTS alexa1.category;
DROP TABLE IF EXISTS alexa1.category_by_entity;

CREATE TABLE alexa1.category (
    id timeuuid,
    external_id text,
    category_name text,
    create_time timestamp,
    update_time timestamp,
    active boolean,
    PRIMARY KEY(external_id, id)
) WITH CLUSTERING ORDER BY (id DESC);

CREATE TABLE alexa1.category_by_entity (
    category_id text,
    entity_id text,
    create_time timeuuid,
    service text,
    category_name text,
    PRIMARY KEY(category_id, create_time)
) WITH CLUSTERING ORDER BY (create_time DESC);
