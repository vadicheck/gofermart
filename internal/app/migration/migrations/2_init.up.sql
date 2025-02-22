CREATE TYPE order_status AS ENUM ('new', 'completed');

CREATE TABLE IF NOT EXISTS orders
(
    id         SERIAL PRIMARY KEY,
    user_id    INT                      NOT NULL,
    order_id   BIGINT                   NOT NULL UNIQUE,
    status     order_status             NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_order ON orders (user_id, order_id);
