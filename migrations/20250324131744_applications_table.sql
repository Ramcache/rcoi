-- +goose Up
CREATE TABLE IF NOT EXISTS applications (
                                            id SERIAL PRIMARY KEY,
                                            title VARCHAR(255) NOT NULL,
                                            description TEXT,
                                            filename VARCHAR(500),
                                            url VARCHAR(500),
                                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS applications;
