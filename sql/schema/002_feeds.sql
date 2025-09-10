-- +goose Up
CREATE TABLE feeds(
id SERIAL PRIMARY KEY,
created_at timestamptz NOT NULL DEFAULT NOW(),
updated_at timestamptz NOT NULL DEFAULT NOW(),
name TEXT,
url TEXT UNIQUE,
user_id UUID NOT NULL, 
FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
