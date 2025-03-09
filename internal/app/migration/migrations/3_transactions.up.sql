CREATE TABLE IF NOT EXISTS transactions
(
    id         SERIAL PRIMARY KEY,
    user_id    INT                      NOT NULL,
    order_id   VARCHAR(255)             NOT NULL UNIQUE,
    sum        NUMERIC(10, 2)                    DEFAULT 0 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
);

CREATE UNIQUE INDEX idx_tr_user_order ON orders (user_id, order_id);
