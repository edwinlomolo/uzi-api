-- name: CreateTrip :one
INSERT INTO trips (
  user_id, product_id, courier_id, route_id, start_location, end_location
) VALUES (
  $1, $2, $3, $4, sqlc.arg(start_location), sqlc.arg(end_location)
)
RETURNING *;

-- name: AssignRouteToTrip :one
UPDATE trips
SET route_id = $1
WHERE id = $2
RETURNING *;

-- name: AssignTripToCourier :one
UPDATE couriers
SET trip_id = $1
WHERE id = $2
RETURNING *;

-- name: UnassignTripToCourier :one
UPDATE couriers
SET trip_id = null
WHERE id = $1
RETURNING *;

-- name: CreateTripCost :one
UPDATE trips
SET cost = $1
WHERE id = $2
RETURNING *;

-- name: SetTripStatus :one
UPDATE trips
SET status = $1
WHERE id = $2
RETURNING *;
