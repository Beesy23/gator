// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: createposts.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createPost = `-- name: CreatePost :one
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
RETURNING id, created_at, updated_at, title, url, description, published_at, feed_id
`

type CreatePostParams struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Url         string
	Description sql.NullString
	PublishedAt sql.NullTime
	FeedID      uuid.UUID
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, createPost,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Title,
		arg.Url,
		arg.Description,
		arg.PublishedAt,
		arg.FeedID,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Url,
		&i.Description,
		&i.PublishedAt,
		&i.FeedID,
	)
	return i, err
}

const getPosts = `-- name: GetPosts :many
SELECT posts.id, posts.created_at, posts.updated_at, posts.title, posts.url, posts.description, posts.published_at, posts.feed_id, feeds.name AS feed_name 
FROM posts
JOIN feed_follows ON feed_follows.feed.id = posts.feed_id
JOIN feeds ON posts.feeds_id=feed.id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2
`

type GetPostsParams struct {
	UserID uuid.UUID
	Limit  int32
}

type GetPostsRow struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Url         string
	Description sql.NullString
	PublishedAt sql.NullTime
	FeedID      uuid.UUID
	FeedName    string
}

func (q *Queries) GetPosts(ctx context.Context, arg GetPostsParams) ([]GetPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPosts, arg.UserID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsRow
	for rows.Next() {
		var i GetPostsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Url,
			&i.Description,
			&i.PublishedAt,
			&i.FeedID,
			&i.FeedName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
