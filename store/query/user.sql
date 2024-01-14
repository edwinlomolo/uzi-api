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
