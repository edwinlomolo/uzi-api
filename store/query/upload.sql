-- name: CreateCourierUpload :one
INSERT INTO uploads (
  type, uri, courier_id
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetCourierUpload :one
SELECT * FROM
uploads
WHERE courier_id = $1 AND type = $2
LIMIT 1;

-- name: GetCourierUploads :many
SELECT * FROM uploads
WHERE courier_id = $1;

-- name: UpdateUpload :one
UPDATE uploads
SET uri = COALESCE($1, uri)
WHERE id = $2
RETURNING *;

-- name: CreateUserUpload :one
INSERT INTO uploads (
  type, uri, user_id
) VALUES (
  $1, $2, $3
)
RETURNING *;
