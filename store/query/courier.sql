-- name: CreateCourier :one
INSERT INTO couriers (
  user_id
) VALUES (
  $1
)
RETURNING *;

-- name: SetCourierStatus :one
UPDATE couriers
SET status = $1
WHERE user_id = $2
RETURNING *;

-- name: GetCourierStatus :one
SELECT status FROM
couriers
WHERE user_id = $1
LIMIT 1;

-- name: AssignTripToCourier :one
UPDATE couriers
SET trip_id = $1
WHERE id = $2
RETURNING *;

-- name: IsCourier :one
SELECT verified FROM
couriers
WHERE user_id = $1
LIMIT 1;

-- name: GetCourier :one
SELECT * FROM
couriers
WHERE user_id = $1
LIMIT 1;

-- name: TrackCourierLocation :one
UPDATE couriers
SET location = sqlc.arg(location)
WHERE user_id = $1
RETURNING *;

-- name: GetNearbyAvailableCourierProducts :many
SELECT c.id, c.product_id, p.* FROM couriers c
JOIN products p
ON ST_DWithin(c.location, sqlc.arg(point)::geography, sqlc.arg(radius))
WHERE c.product_id = p.id AND c.status = 'ONLINE' AND c.verified = 'true'
ORDER BY p.relevance ASC;

-- name: GetCourierNearPickupPoint :many
SELECT id, product_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE ST_DWithin(location, sqlc.arg(point)::geography, sqlc.arg(radius)) AND status = 'ONLINE' AND verified = 'true';

-- name: GetCourierProductByID :one
SELECT id, icon FROM products
WHERE id = $1
LIMIT 1;
