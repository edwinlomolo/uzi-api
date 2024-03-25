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

-- name: CreateSession :one
INSERT INTO sessions (
  id, ip, user_agent, phone
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1
LIMIT 1;

-- name: CreateTrip :one
INSERT INTO trips (
  user_id, product_id, confirmed_pickup, start_location, end_location
) VALUES (
  $1, $2, $3, sqlc.arg(start_location), sqlc.arg(end_location)
)
RETURNING *;

-- name: GetNearbyAvailableCourierProducts :many
SELECT c.id, c.product_id, p.* FROM couriers c
JOIN products p
ON ST_DWithin(c.location, sqlc.arg(point)::geography, sqlc.arg(radius))
WHERE c.product_id = p.id AND c.verified = 'true'
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
SELECT id, status, courier_id, cost, product_id, ST_AsGeoJSON(confirmed_pickup) AS confirmed_pickup, ST_AsGeoJSON(start_location) AS start_location, ST_AsGeoJSON(end_location) AS end_location FROM trips
WHERE id = $1
LIMIT 1;

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
  name, building, unit, phone, trip_id, trip_note
) VALUES (
  $1, $2, $3, $4, $5, $6
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

-- name: CreateCourierUpload :one
INSERT INTO uploads (
  type, uri, courier_id, verification
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetCourierUpload :one
SELECT * FROM
uploads
WHERE courier_id = $1 AND type = $2
LIMIT 1;

-- name: GetCourierUploads :many
SELECT * FROM uploads
WHERE courier_id = $1;

-- name: UpdateUpload :one
UPDATE uploads
SET uri = COALESCE(sqlc.narg(uri), uri), verification = COALESCE(sqlc.narg(verification), verification)
WHERE id = $1
RETURNING *;

-- name: CreateUserUpload :one
INSERT INTO uploads (
  type, uri, user_id
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserUpload :one
SELECT * FROM uploads
WHERE user_id = $1 AND type = $2
LIMIT 1;

-- name: GetCourierAvatar :one
SELECT id, uri FROM uploads
WHERE courier_id = $1 AND type = 'DP';

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

-- name: FindUserByID :one
SELECT * FROM users
WHERE id = $1
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
