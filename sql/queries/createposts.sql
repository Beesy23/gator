-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPosts :many
SELECT posts.*, feeds.name AS feed_name 
FROM posts
JOIN feed_follows ON feed_follows.feed.id = posts.feed_id
JOIN feeds ON posts.feeds_id=feed.id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;