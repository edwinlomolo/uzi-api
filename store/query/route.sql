-- name: CreateRoute :one
INSERT INTO routes (
  distance, trip_id, eta, state, polyline
) VALUES (
  $1, $2, $3, $4, ST_GeographyFromText(sqlc.arg(polyline))
)
RETURNING *;
