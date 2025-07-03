-- name: CreateContext :one
INSERT INTO contexts (
    id, title, description, context_description_post_id, created
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetContext :one
SELECT * FROM contexts WHERE id = ?;

-- name: GetContextByTitle :one
SELECT * FROM contexts WHERE title = ?;

-- name: UpdateContext :one
UPDATE contexts 
SET 
    title = COALESCE(?, title),
    description = COALESCE(?, description),
    context_description_post_id = COALESCE(?, context_description_post_id)
WHERE id = ?
RETURNING *;

-- name: DeleteContext :exec
DELETE FROM contexts WHERE id = ?;

-- name: ListContexts :many
SELECT * FROM contexts 
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: SearchContexts :many
SELECT * FROM contexts 
WHERE title LIKE ? OR description LIKE ?
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: AddPostToContext :one
INSERT INTO context_posts (
    id, post_id, context_id, created
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: RemovePostFromContext :exec
DELETE FROM context_posts 
WHERE context_id = ? AND post_id = ?;

-- name: GetContextPosts :many
SELECT p.* FROM posts p
JOIN context_posts cp ON p.id = cp.post_id
WHERE cp.context_id = ?
ORDER BY p.created DESC;
