-- name: CreateTrip :one
INSERT INTO trips (
  user_id, start_location, end_location
) VALUES (
  $1, sqlc.arg(start_location), sqlc.arg(end_location)
)
RETURNING *;

-- name: AssignRouteToTrip :one
UPDATE trips
SET route_id = $1
WHERE id = $2
RETURNING *;
