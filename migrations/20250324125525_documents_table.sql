-- +goose Up
CREATE TABLE IF NOT EXISTS documents (
                                         id SERIAL PRIMARY KEY,
                                         title VARCHAR(255) NOT NULL,
                                         filename VARCHAR(500) NOT NULL,
                                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS documents;
