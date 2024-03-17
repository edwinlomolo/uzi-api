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

-- name: IsCourier :one
SELECT verified FROM
couriers
WHERE user_id = $1
LIMIT 1;

-- name: GetCourierByUserID :one
SELECT id, user_id, product_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE user_id = $1
LIMIT 1;

-- name: GetCourierByID :one
SELECT id, trip_id, product_id, user_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE id = $1
LIMIT 1;

-- name: TrackCourierLocation :one
UPDATE couriers
SET location = sqlc.arg(location)
WHERE user_id = $1
RETURNING *;

-- name: GetProductByID :one
SELECT id, icon, name, weight_class FROM products
WHERE id = $1
LIMIT 1;

-- name: GetCourierLocation :one
SELECT ST_AsGeoJSON(location) AS location FROM
couriers
WHERE id = $1
LIMIT 1;
