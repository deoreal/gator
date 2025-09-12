-- name: CreateFeed :exec
INSERT INTO feeds (created_at, updated_at, name, url, user_id) 
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
--

-- name: GetFeeds :many
SELECT 
    feeds.id,
    feeds.created_at,
    feeds.updated_at,
    feeds.name,
    feeds.url,
    users.name AS user_name
    FROM feeds
JOIN users  ON feeds.user_id = users.id;
--
