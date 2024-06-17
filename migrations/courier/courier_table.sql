CREATE DATABASE IF NOT EXISTS courier;
GRANT ALL PRIVILEGES ON DATABASE courier TO citizix_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS courier (
        courier_id UUID NOT NULL DEFAULT gen_random_uuid(),
        firstname varchar(40) NOT NULL,
        is_available boolean NOT NULL default true,
        PRIMARY KEY (courier_id)
    );

CREATE TABLE IF NOT EXISTS order_assignments (
        order_id UUID NOT NULL,
        courier_id UUID NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        PRIMARY KEY (order_id)
    );
