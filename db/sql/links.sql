-- name: CreateLink :one
INSERT INTO links (
    id, title, is_local_href, href, image_url, find_out_more_href, 
    click_count, is_visible, "order", created
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetLink :one
SELECT * FROM links WHERE id = ?;

-- name: UpdateLink :one
UPDATE links 
SET 
    title = COALESCE(?, title),
    is_local_href = COALESCE(?, is_local_href),
    href = COALESCE(?, href),
    image_url = COALESCE(?, image_url),
    find_out_more_href = COALESCE(?, find_out_more_href),
    click_count = COALESCE(?, click_count),
    is_visible = COALESCE(?, is_visible),
    "order" = COALESCE(?, "order")
WHERE id = ?
RETURNING *;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = ?;

-- name: ListVisibleLinks :many
SELECT * FROM links 
WHERE is_visible = true 
ORDER BY "order", created DESC;

-- name: ListAllLinks :many
SELECT * FROM links 
ORDER BY "order", created DESC 
LIMIT ? OFFSET ?;

-- name: IncrementLinkClicks :exec
UPDATE links 
SET click_count = click_count + 1 
WHERE id = ?;
