-- name: CreateCollection :one
INSERT INTO collections (
    id, title, slug, description, collection_description_post_id, clicked_count, created
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetCollection :one
SELECT * FROM collections WHERE id = ?;

-- name: GetCollectionBySlug :one
SELECT * FROM collections WHERE slug = ?;

-- name: UpdateCollection :one
UPDATE collections 
SET 
    title = COALESCE(?, title),
    slug = COALESCE(?, slug),
    description = COALESCE(?, description),
    collection_description_post_id = COALESCE(?, collection_description_post_id),
    clicked_count = COALESCE(?, clicked_count)
WHERE id = ?
RETURNING *;

-- name: DeleteCollection :exec
DELETE FROM collections WHERE id = ?;

-- name: ListCollections :many
SELECT * FROM collections 
ORDER BY clicked_count DESC, created DESC 
LIMIT ? OFFSET ?;

-- name: SearchCollections :many
SELECT * FROM collections 
WHERE title LIKE ? OR description LIKE ?
ORDER BY clicked_count DESC, created DESC 
LIMIT ? OFFSET ?;

-- name: IncrementCollectionClicks :exec
UPDATE collections 
SET clicked_count = clicked_count + 1 
WHERE id = ?;

-- name: AddPostToCollection :one
INSERT INTO collection_posts (
    id, collection_id, post_id, "order", created
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: RemovePostFromCollection :exec
DELETE FROM collection_posts 
WHERE collection_id = ? AND post_id = ?;

-- name: GetCollectionPosts :many
SELECT p.* FROM posts p
JOIN collection_posts cp ON p.id = cp.post_id
WHERE cp.collection_id = ?
ORDER BY cp."order", cp.created;

-- name: UpdatePostOrderInCollection :exec
UPDATE collection_posts 
SET "order" = ? 
WHERE collection_id = ? AND post_id = ?;
