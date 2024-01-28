// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: trip.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const assignRouteToTrip = `-- name: AssignRouteToTrip :one
UPDATE trips
SET route_id = $1
WHERE id = $2
RETURNING id, start_location, end_location, courier_id, user_id, route_id, cost, status, created_at, updated_at
`

type AssignRouteToTripParams struct {
	RouteID uuid.NullUUID `json:"route_id"`
	ID      uuid.UUID     `json:"id"`
}

func (q *Queries) AssignRouteToTrip(ctx context.Context, arg AssignRouteToTripParams) (Trip, error) {
	row := q.db.QueryRowContext(ctx, assignRouteToTrip, arg.RouteID, arg.ID)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createTrip = `-- name: CreateTrip :one
INSERT INTO trips (
  user_id, start_location, end_location
) VALUES (
  $1, $2, $3
)
RETURNING id, start_location, end_location, courier_id, user_id, route_id, cost, status, created_at, updated_at
`

type CreateTripParams struct {
	UserID        uuid.NullUUID `json:"user_id"`
	StartLocation interface{}   `json:"start_location"`
	EndLocation   interface{}   `json:"end_location"`
}

func (q *Queries) CreateTrip(ctx context.Context, arg CreateTripParams) (Trip, error) {
	row := q.db.QueryRowContext(ctx, createTrip, arg.UserID, arg.StartLocation, arg.EndLocation)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}