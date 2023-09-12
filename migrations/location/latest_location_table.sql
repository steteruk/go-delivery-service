CREATE DATABASE IF NOT EXISTS courier_location;
GRANT ALL PRIVILEGES ON DATABASE courier_location TO citizix_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS courier_latest_cord (
                                                   courier_id UUID NOT NULL,
                                                   latitude double precision NOT NULL,
                                                   longitude double precision NOT NULL ,
                                                   created_at TIMESTAMPTZ NOT NULL,
                                                   PRIMARY KEY (courier_id, created_at)
    );