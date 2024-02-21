// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: trip.sql

package sqlc

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const assignCourierToTrip = `-- name: AssignCourierToTrip :one
UPDATE couriers
SET trip_id = $1
WHERE id = $2
RETURNING id, verified, status, location, ratings, points, user_id, product_id, trip_id, created_at, updated_at
`

type AssignCourierToTripParams struct {
	TripID uuid.NullUUID `json:"trip_id"`
	ID     uuid.UUID     `json:"id"`
}

func (q *Queries) AssignCourierToTrip(ctx context.Context, arg AssignCourierToTripParams) (Courier, error) {
	row := q.db.QueryRowContext(ctx, assignCourierToTrip, arg.TripID, arg.ID)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Ratings,
		&i.Points,
		&i.UserID,
		&i.ProductID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const assignRouteToTrip = `-- name: AssignRouteToTrip :one
UPDATE trips
SET route_id = $1
WHERE id = $2
RETURNING id, start_location, end_location, confirmed_pickup, courier_id, user_id, route_id, product_id, cost, status, created_at, updated_at
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
		&i.ConfirmedPickup,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.ProductID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const assignTripToCourier = `-- name: AssignTripToCourier :one
UPDATE trips
SET courier_id = $1
WHERE id = $2
RETURNING id, start_location, end_location, confirmed_pickup, courier_id, user_id, route_id, product_id, cost, status, created_at, updated_at
`

type AssignTripToCourierParams struct {
	CourierID uuid.NullUUID `json:"courier_id"`
	ID        uuid.UUID     `json:"id"`
}

func (q *Queries) AssignTripToCourier(ctx context.Context, arg AssignTripToCourierParams) (Trip, error) {
	row := q.db.QueryRowContext(ctx, assignTripToCourier, arg.CourierID, arg.ID)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.ConfirmedPickup,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.ProductID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createRecipient = `-- name: CreateRecipient :one
INSERT INTO recipients (
  name, building, unit, phone, trip_id, trip_note
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, name, building, unit, phone, trip_note, trip_id, created_at, updated_at
`

type CreateRecipientParams struct {
	Name     string         `json:"name"`
	Building sql.NullString `json:"building"`
	Unit     sql.NullString `json:"unit"`
	Phone    string         `json:"phone"`
	TripID   uuid.NullUUID  `json:"trip_id"`
	TripNote string         `json:"trip_note"`
}

func (q *Queries) CreateRecipient(ctx context.Context, arg CreateRecipientParams) (Recipient, error) {
	row := q.db.QueryRowContext(ctx, createRecipient,
		arg.Name,
		arg.Building,
		arg.Unit,
		arg.Phone,
		arg.TripID,
		arg.TripNote,
	)
	var i Recipient
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Building,
		&i.Unit,
		&i.Phone,
		&i.TripNote,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createTrip = `-- name: CreateTrip :one
INSERT INTO trips (
  user_id, product_id, confirmed_pickup, start_location, end_location
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id, start_location, end_location, confirmed_pickup, courier_id, user_id, route_id, product_id, cost, status, created_at, updated_at
`

type CreateTripParams struct {
	UserID          uuid.UUID   `json:"user_id"`
	ProductID       uuid.UUID   `json:"product_id"`
	ConfirmedPickup interface{} `json:"confirmed_pickup"`
	StartLocation   interface{} `json:"start_location"`
	EndLocation     interface{} `json:"end_location"`
}

func (q *Queries) CreateTrip(ctx context.Context, arg CreateTripParams) (Trip, error) {
	row := q.db.QueryRowContext(ctx, createTrip,
		arg.UserID,
		arg.ProductID,
		arg.ConfirmedPickup,
		arg.StartLocation,
		arg.EndLocation,
	)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.ConfirmedPickup,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.ProductID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createTripCost = `-- name: CreateTripCost :one
UPDATE trips
SET cost = $1
WHERE id = $2
RETURNING id, start_location, end_location, confirmed_pickup, courier_id, user_id, route_id, product_id, cost, status, created_at, updated_at
`

type CreateTripCostParams struct {
	Cost sql.NullString `json:"cost"`
	ID   uuid.UUID      `json:"id"`
}

func (q *Queries) CreateTripCost(ctx context.Context, arg CreateTripCostParams) (Trip, error) {
	row := q.db.QueryRowContext(ctx, createTripCost, arg.Cost, arg.ID)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.ConfirmedPickup,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.ProductID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findAvailableCourier = `-- name: FindAvailableCourier :one
SELECT id, user_id, product_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE ST_DWithin(location, $1::geography, $2) AND status = 'ONLINE' AND verified = 'true' AND trip_id IS null
LIMIT 1
`

type FindAvailableCourierParams struct {
	Point  interface{} `json:"point"`
	Radius interface{} `json:"radius"`
}

type FindAvailableCourierRow struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.NullUUID `json:"user_id"`
	ProductID uuid.NullUUID `json:"product_id"`
	Location  interface{}   `json:"location"`
}

func (q *Queries) FindAvailableCourier(ctx context.Context, arg FindAvailableCourierParams) (FindAvailableCourierRow, error) {
	row := q.db.QueryRowContext(ctx, findAvailableCourier, arg.Point, arg.Radius)
	var i FindAvailableCourierRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.Location,
	)
	return i, err
}

const getCourierAssignedTrip = `-- name: GetCourierAssignedTrip :one
SELECT id, verified, status, location, ratings, points, user_id, product_id, trip_id, created_at, updated_at FROM couriers
WHERE id = $1 AND trip_id = null
LIMIT 1
`

func (q *Queries) GetCourierAssignedTrip(ctx context.Context, id uuid.UUID) (Courier, error) {
	row := q.db.QueryRowContext(ctx, getCourierAssignedTrip, id)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Ratings,
		&i.Points,
		&i.UserID,
		&i.ProductID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourierNearPickupPoint = `-- name: GetCourierNearPickupPoint :many
SELECT id, product_id, ST_AsGeoJSON(location) AS location FROM
couriers
WHERE ST_DWithin(location, $1::geography, $2) AND status = 'ONLINE' AND verified = 'true'
`

type GetCourierNearPickupPointParams struct {
	Point  interface{} `json:"point"`
	Radius interface{} `json:"radius"`
}

type GetCourierNearPickupPointRow struct {
	ID        uuid.UUID     `json:"id"`
	ProductID uuid.NullUUID `json:"product_id"`
	Location  interface{}   `json:"location"`
}

func (q *Queries) GetCourierNearPickupPoint(ctx context.Context, arg GetCourierNearPickupPointParams) ([]GetCourierNearPickupPointRow, error) {
	rows, err := q.db.QueryContext(ctx, getCourierNearPickupPoint, arg.Point, arg.Radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetCourierNearPickupPointRow{}
	for rows.Next() {
		var i GetCourierNearPickupPointRow
		if err := rows.Scan(&i.ID, &i.ProductID, &i.Location); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCourierTrip = `-- name: GetCourierTrip :one
SELECT id, start_location, end_location, confirmed_pickup, courier_id, user_id, route_id, product_id, cost, status, created_at, updated_at FROM trips
WHERE courier_id = $1
LIMIT 1
`

func (q *Queries) GetCourierTrip(ctx context.Context, courierID uuid.NullUUID) (Trip, error) {
	row := q.db.QueryRowContext(ctx, getCourierTrip, courierID)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.ConfirmedPickup,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.ProductID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNearbyAvailableCourierProducts = `-- name: GetNearbyAvailableCourierProducts :many
SELECT c.id, c.product_id, p.id, p.name, p.description, p.weight_class, p.icon, p.relevance, p.created_at, p.updated_at FROM couriers c
JOIN products p
ON ST_DWithin(c.location, $1::geography, $2)
WHERE c.product_id = p.id AND c.status = 'OFFLINE' AND c.verified = 'true'
ORDER BY p.relevance ASC
`

type GetNearbyAvailableCourierProductsParams struct {
	Point  interface{} `json:"point"`
	Radius interface{} `json:"radius"`
}

type GetNearbyAvailableCourierProductsRow struct {
	ID          uuid.UUID     `json:"id"`
	ProductID   uuid.NullUUID `json:"product_id"`
	ID_2        uuid.UUID     `json:"id_2"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	WeightClass int32         `json:"weight_class"`
	Icon        string        `json:"icon"`
	Relevance   int32         `json:"relevance"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (q *Queries) GetNearbyAvailableCourierProducts(ctx context.Context, arg GetNearbyAvailableCourierProductsParams) ([]GetNearbyAvailableCourierProductsRow, error) {
	rows, err := q.db.QueryContext(ctx, getNearbyAvailableCourierProducts, arg.Point, arg.Radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetNearbyAvailableCourierProductsRow{}
	for rows.Next() {
		var i GetNearbyAvailableCourierProductsRow
		if err := rows.Scan(
			&i.ID,
			&i.ProductID,
			&i.ID_2,
			&i.Name,
			&i.Description,
			&i.WeightClass,
			&i.Icon,
			&i.Relevance,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTrip = `-- name: GetTrip :one
SELECT id, status, courier_id, ST_AsGeoJSON(start_location) AS start_location FROM trips
WHERE id = $1
LIMIT 1
`

type GetTripRow struct {
	ID            uuid.UUID     `json:"id"`
	Status        string        `json:"status"`
	CourierID     uuid.NullUUID `json:"courier_id"`
	StartLocation interface{}   `json:"start_location"`
}

func (q *Queries) GetTrip(ctx context.Context, id uuid.UUID) (GetTripRow, error) {
	row := q.db.QueryRowContext(ctx, getTrip, id)
	var i GetTripRow
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CourierID,
		&i.StartLocation,
	)
	return i, err
}

const getTripRecipient = `-- name: GetTripRecipient :one
SELECT id, name, building, unit, phone, trip_note, trip_id, created_at, updated_at FROM recipients
WHERE trip_id = $1
LIMIT 1
`

func (q *Queries) GetTripRecipient(ctx context.Context, tripID uuid.NullUUID) (Recipient, error) {
	row := q.db.QueryRowContext(ctx, getTripRecipient, tripID)
	var i Recipient
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Building,
		&i.Unit,
		&i.Phone,
		&i.TripNote,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const setTripStatus = `-- name: SetTripStatus :one
UPDATE trips
SET status = $1
WHERE id = $2
RETURNING id, start_location, end_location, confirmed_pickup, courier_id, user_id, route_id, product_id, cost, status, created_at, updated_at
`

type SetTripStatusParams struct {
	Status string    `json:"status"`
	ID     uuid.UUID `json:"id"`
}

func (q *Queries) SetTripStatus(ctx context.Context, arg SetTripStatusParams) (Trip, error) {
	row := q.db.QueryRowContext(ctx, setTripStatus, arg.Status, arg.ID)
	var i Trip
	err := row.Scan(
		&i.ID,
		&i.StartLocation,
		&i.EndLocation,
		&i.ConfirmedPickup,
		&i.CourierID,
		&i.UserID,
		&i.RouteID,
		&i.ProductID,
		&i.Cost,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const unassignCourierTrip = `-- name: UnassignCourierTrip :one
UPDATE couriers
SET trip_id = null
WHERE id = $1
RETURNING id, verified, status, location, ratings, points, user_id, product_id, trip_id, created_at, updated_at
`

func (q *Queries) UnassignCourierTrip(ctx context.Context, id uuid.UUID) (Courier, error) {
	row := q.db.QueryRowContext(ctx, unassignCourierTrip, id)
	var i Courier
	err := row.Scan(
		&i.ID,
		&i.Verified,
		&i.Status,
		&i.Location,
		&i.Ratings,
		&i.Points,
		&i.UserID,
		&i.ProductID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
