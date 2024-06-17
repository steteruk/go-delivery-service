-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS couriers (
                                       id UUID DEFAULT gen_random_uuid(),
    first_name char(30) NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (id)
    );
CREATE TABLE IF NOT EXISTS order_assignments (
                                                 courier_id UUID NOT NULL,
                                                 order_id UUID NOT NULL,
                                                 created_at TIMESTAMPTZ NOT NULL,
                                                 PRIMARY KEY (order_id)
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE couriers;
DROP TABLE order_assignments;
-- +goose StatementEnd
