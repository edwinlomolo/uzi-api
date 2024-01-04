-- name: CreateUser :one
INSERT INTO users (
  first_name, last_name, phone
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: FindByPhone :one
SELECT * FROM users
WHERE phone = $1
LIMIT 1;

-- name: IsUserOnboarding :one
SELECT onboarding FROM
users
WHERE id = $1
LIMIT 1;

-- name: SetOnboardingStatus :one
UPDATE users
SET onboarding = $1
WHERE phone = $2
RETURNING *;

-- name: UpdateUserName :one
UPDATE users
SET first_name = COALESCE($1, first_name), last_name = COALESCE($2, last_name)
WHERE phone = $3
RETURNING *;

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

-- name: CreateVehicle :one
INSERT INTO vehicles (
  product_id
) VALUES (
  $1
)
RETURNING *;

-- name: UpdateProductLocation :one
UPDATE products
SET location = $1
WHERE id = $2
RETURNING *;

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

-- name: CreateRoute :one
INSERT INTO routes (
  distance, eta, polyline
) VALUES (
  $1, $2, ST_GeographyFromText(sqlc.arg(polyline))
)
RETURNING *;
