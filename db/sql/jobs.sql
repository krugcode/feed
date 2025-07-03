
-- name: CreateJob :one
INSERT INTO jobs (
    id, job_name, status, result_message, completed, created
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = ?;

-- name: UpdateJobStatus :one
UPDATE jobs 
SET 
    status = ?,
    result_message = COALESCE(?, result_message),
    completed = COALESCE(?, completed)
WHERE id = ?
RETURNING *;

-- name: ListJobs :many
SELECT * FROM jobs 
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: ListJobsByStatus :many
SELECT * FROM jobs 
WHERE status = ?
ORDER BY created DESC 
LIMIT ? OFFSET ?;

-- name: GetJobsByName :many
SELECT * FROM jobs 
WHERE job_name = ?
ORDER BY created DESC 
LIMIT ?;

-- Crosspost Jobs
-- name: CreateCrosspostJob :one
INSERT INTO crosspost_jobs (
    id, platform, post_id, created
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: GetCrosspostJobsByPost :many
SELECT * FROM crosspost_jobs 
WHERE post_id = ?
ORDER BY created DESC;

-- name: GetCrosspostJobsByPlatform :many
SELECT * FROM crosspost_jobs 
WHERE platform = ?
ORDER BY created DESC 
LIMIT ? OFFSET ?;
