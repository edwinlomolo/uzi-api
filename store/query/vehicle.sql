-- name: CreateVehicle :one
INSERT INTO vehicles (
  product_id, courier_id, mass
) VALUES (
  $1, $2, $3
)
RETURNING *;
