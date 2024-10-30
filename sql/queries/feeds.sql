-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name, url, users.name AS user_name 
FROM feeds
INNER JOIN users
ON feeds.user_id = users.id;

-- name: GetFeedFromURL :one
SELECT id AS feed_id, name as feed_name, url, user_id FROM feeds WHERE url = $1;