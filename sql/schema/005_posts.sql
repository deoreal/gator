-- +goose Up
CREATE TABLE posts(
id SERIAL PRIMARY KEY,
created_at timestamptz NOT NULL DEFAULT NOW(),
updated_at timestamptz NOT NULL DEFAULT NOW(),
title TEXT,
url TEXT UNIQUE NOT NULL,
description TEXT,
published_at timestamptz,
feed_id INTEGER NOT NULL,
FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
