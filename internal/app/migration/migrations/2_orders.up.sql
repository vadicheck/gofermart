CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders
(
    id         SERIAL PRIMARY KEY,
    user_id    INT                      NOT NULL,
    order_id   VARCHAR(255)             NOT NULL UNIQUE,
    accrual    NUMERIC(10, 2)           NOT NULL DEFAULT 0,
    status     order_status             NOT NULL DEFAULT 'NEW'::order_status,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
);

CREATE UNIQUE INDEX idx_user_order ON orders (user_id, order_id);
