-- name: CreateRoute :one
INSERT INTO routes (
  distance, eta, polyline
) VALUES (
  $1, $2, ST_GeographyFromText(sqlc.arg(polyline))
)
RETURNING *;
