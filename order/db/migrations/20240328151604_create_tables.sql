-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    id UUID DEFAULT gen_random_uuid(),
    courier_id UUID NULL,
    customer_phone_number char(15) NOT NULL,
    status order_status DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS order_validations (
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    courier_validated_at TIMESTAMPTZ,
    courier_error VARCHAR(256),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (order_id)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE order_validations;
DROP TABLE orders;
-- +goose StatementEnd
