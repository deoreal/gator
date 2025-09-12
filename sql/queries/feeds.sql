-- name: CreateFeed :one
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
    feeds.name,
    feeds.url,
    feeds.user_id,
    users.name AS user_name
    FROM feeds
JOIN users  ON feeds.user_id = users.id;
--
-- name: MarkFeedFetched :exec
UPDATE feeds SET last_fetched_at = NOW() WHERE id = $1;
UPDATE feeds SET updated_at = NOW() WHERE id = $1;
--

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at NULLS FIRST
LIMIT 1;
--

-- name: GetFeed :one
SELECT
    feeds.id,
    feeds.created_at,
    feeds.updated_at,
    feeds.name,
    users.name AS user_name
    FROM feeds
JOIN users  ON feeds.user_id = users.id
WHERE feeds.url = $1;
--

-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, $2)
    RETURNING *
)
SELECT
    i.id,
    i.created_at,
    i.updated_at,
    i.user_id,
    i.feed_id,
    u.name AS user_name,
    f.name AS feed_name
FROM inserted i
JOIN users u ON i.user_id = u.id
JOIN feeds f ON i.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
SELECT
    f.name
FROM feed_follows ff
JOIN feeds f ON ff.feed_id = f.id
JOIN users u ON ff.user_id = u.id
WHERE u.name = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;
