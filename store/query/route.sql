-- name: CreateRoute :one
INSERT INTO routes (
  distance, eta, state, polyline
) VALUES (
  $1, $2, $3, ST_GeographyFromText(sqlc.arg(polyline))
)
RETURNING *;
