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

-- name: GetCourierProductByID :one
SELECT id, icon, name FROM products
WHERE id = $1
LIMIT 1;
