// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: courier.sql

package sqlc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const assignTripToCourier = `-- name: AssignTripToCourier :one
UPDATE couriers
SET trip_id = $1
WHERE id = $2
RETURNING id, verified, status, location, rating, points, user_id, trip_id, created_at, updated_at
`

type AssignTripToCourierParams struct {
	TripID uuid.NullUUID `json:"trip_id"`
	ID     uuid.UUID     `json:"id"`
}

func (q *Queries) AssignTripToCourier(ctx context.Context, arg AssignTripToCourierParams) (Courier, error) {
	row := q.db.QueryRowContext(ctx, assignTripToCourier, arg.TripID, arg.ID)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Rating,
		&i.Points,
		&i.UserID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createCourier = `-- name: CreateCourier :one
INSERT INTO couriers (
  user_id
) VALUES (
  $1
)
RETURNING id, verified, status, location, rating, points, user_id, trip_id, created_at, updated_at
`

func (q *Queries) CreateCourier(ctx context.Context, userID uuid.NullUUID) (Courier, error) {
	row := q.db.QueryRowContext(ctx, createCourier, userID)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Rating,
		&i.Points,
		&i.UserID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourier = `-- name: GetCourier :one
SELECT id, verified, status, location, rating, points, user_id, trip_id, created_at, updated_at FROM
couriers
WHERE user_id = $1
LIMIT 1
`

func (q *Queries) GetCourier(ctx context.Context, userID uuid.NullUUID) (Courier, error) {
	row := q.db.QueryRowContext(ctx, getCourier, userID)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Rating,
		&i.Points,
		&i.UserID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourierStatus = `-- name: GetCourierStatus :one
SELECT status FROM
couriers
WHERE user_id = $1
LIMIT 1
`

func (q *Queries) GetCourierStatus(ctx context.Context, userID uuid.NullUUID) (string, error) {
	row := q.db.QueryRowContext(ctx, getCourierStatus, userID)
	var status string
	err := row.Scan(&status)
	return status, err
}

const isCourier = `-- name: IsCourier :one
SELECT verified FROM
couriers
WHERE user_id = $1
LIMIT 1
`

func (q *Queries) IsCourier(ctx context.Context, userID uuid.NullUUID) (sql.NullBool, error) {
	row := q.db.QueryRowContext(ctx, isCourier, userID)
	var verified sql.NullBool
	err := row.Scan(&verified)
	return verified, err
}

const setCourierStatus = `-- name: SetCourierStatus :one
UPDATE couriers
SET status = $1
WHERE user_id = $2
RETURNING id, verified, status, location, rating, points, user_id, trip_id, created_at, updated_at
`

type SetCourierStatusParams struct {
	Status string        `json:"status"`
	UserID uuid.NullUUID `json:"user_id"`
}

func (q *Queries) SetCourierStatus(ctx context.Context, arg SetCourierStatusParams) (Courier, error) {
	row := q.db.QueryRowContext(ctx, setCourierStatus, arg.Status, arg.UserID)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Rating,
		&i.Points,
		&i.UserID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
