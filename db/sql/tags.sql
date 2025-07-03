-- name: CreateTag :one
INSERT INTO tags (
    id, title, searched_count, created
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: GetTag :one
SELECT * FROM tags WHERE id = ?;

-- name: GetTagByTitle :one
SELECT * FROM tags WHERE title = ?;

-- name: UpdateTag :one
UPDATE tags 
SET 
    title = COALESCE(?, title),
    searched_count = COALESCE(?, searched_count)
WHERE id = ?
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- name: ListTags :many
SELECT * FROM tags 
ORDER BY searched_count DESC, title ASC 
LIMIT ? OFFSET ?;

-- name: SearchTags :many
SELECT * FROM tags 
WHERE title LIKE ?
ORDER BY searched_count DESC, title ASC 
LIMIT ? OFFSET ?;

-- name: GetPopularTags :many
SELECT * FROM tags 
WHERE searched_count > 0
ORDER BY searched_count DESC 
LIMIT ?;

-- name: IncrementTagSearchCount :exec
UPDATE tags 
SET searched_count = searched_count + 1 
WHERE id = ?;

-- name: AddTagToPost :one
INSERT INTO post_tags (
    id, post_id, tag_id, created
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: RemoveTagFromPost :exec
DELETE FROM post_tags 
WHERE post_id = ? AND tag_id = ?;

-- name: GetPostTags :many
SELECT t.* FROM tags t
JOIN post_tags pt ON t.id = pt.tag_id
WHERE pt.post_id = ?
ORDER BY t.title;

-- name: GetTagPosts :many
SELECT p.* FROM posts p
JOIN post_tags pt ON p.id = pt.post_id
WHERE pt.tag_id = ? AND p.visible = true
ORDER BY p.created DESC;
