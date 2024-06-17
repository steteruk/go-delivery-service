CREATE DATABASE "order";
GRANT ALL PRIVILEGES ON DATABASE "order" TO citizix_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
create type order_status as enum ('pending', 'accepted', 'in_progress', 'delivered', 'canceled');
