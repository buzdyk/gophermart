CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    number VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    accrual NUMERIC(10, 2) DEFAULT 0,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT unique_number UNIQUE (number)
);