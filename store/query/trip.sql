-- name: CreateTrip :one
INSERT INTO trips (
  user_id, product_id, start_location, end_location
) VALUES (
  $1, $2, sqlc.arg(start_location), sqlc.arg(end_location)
)
RETURNING *;

-- name: GetNearbyAvailableCourierProducts :many
SELECT c.id, c.product_id, p.* FROM couriers c
JOIN products p
ON ST_DWithin(c.location, sqlc.arg(point)::geography, sqlc.arg(radius))
WHERE c.product_id = p.id AND c.status = 'OFFLINE' AND c.verified = 'true'
ORDER BY p.relevance ASC;

-- name: FindAvailableCourier :one
SELECT id, user_id, product_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE ST_DWithin(location, sqlc.arg(point)::geography, sqlc.arg(radius)) AND status = 'ONLINE' AND verified = 'true' AND trip_id IS null
LIMIT 1;

-- name: GetCourierNearPickupPoint :many
SELECT id, product_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE ST_DWithin(location, sqlc.arg(point)::geography, sqlc.arg(radius)) AND status = 'ONLINE' AND verified = 'true';

-- name: GetTrip :one
SELECT id, status, courier_id, ST_AsGeoJSON(start_location) AS start_location FROM trips
WHERE id = $1
LIMIT 1;

-- name: AssignRouteToTrip :one
UPDATE trips
SET route_id = $1
WHERE id = $2
RETURNING *;

-- name: AssignCourierToTrip :one
UPDATE couriers
SET trip_id = $1
WHERE id = $2
RETURNING *;

-- name: AssignTripToCourier :one
UPDATE trips
SET courier_id = $1
WHERE id = $2
RETURNING *;

-- name: UnassignCourierTrip :one
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

-- name: GetCourierAssignedTrip :one
SELECT * FROM couriers
WHERE id = $1 AND trip_id = null
LIMIT 1;

-- name: CreateRecipient :one
INSERT INTO recipients (
  name, building, unit, phone, trip_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetTripRecipient :one
SELECT * FROM recipients
WHERE trip_id = $1
LIMIT 1;

-- name: GetCourierTrip :one
SELECT * FROM trips
WHERE courier_id = $1
LIMIT 1;

-- name: AssignTripRoute :one
UPDATE trips
SET route_id = $1
RETURNING *;