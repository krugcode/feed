-- name: CreatePost :one
INSERT INTO posts (
    id, type, visible, title, subtitle, content, slug, permalink, created
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetPost :one
SELECT * FROM posts WHERE id = ?;

-- name: GetPostBySlug :one
SELECT * FROM posts WHERE slug = ?;

-- name: GetPostByPermalink :one
SELECT * FROM posts WHERE permalink = ?;

-- name: UpdatePost :one
UPDATE posts 
SET 
    type = COALESCE(?, type),
    visible = COALESCE(?, visible),
    title = COALESCE(?, title),
    subtitle = COALESCE(?, subtitle),
    content = COALESCE(?, content),
    slug = COALESCE(?, slug),
    permalink = COALESCE(?, permalink)
WHERE id = ?
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = ?;

-- name: ListPosts :many
SELECT * FROM posts 
WHERE visible = ? 
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: ListPostsByType :many
SELECT * FROM posts 
WHERE type = ? AND visible = ? 
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: SearchPosts :many
SELECT * FROM posts 
WHERE visible = ? 
AND (title LIKE ? OR subtitle LIKE ? OR content LIKE ?)
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: CountPosts :one
SELECT COUNT(*) FROM posts WHERE visible = ?;

-- name: CountPostsByType :one
SELECT COUNT(*) FROM posts WHERE type = ? AND visible = ?;

-- name: GetRecentPosts :many
SELECT * FROM posts 
WHERE visible = true 
ORDER BY created DESC 
LIMIT ?;
