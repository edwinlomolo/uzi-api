// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: queries.sql

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const assignRouteToTrip = `-- name: AssignRouteToTrip :one
UPDATE trips
SET route_id = $1
WHERE id = $2
RETURNING id, start_location, end_location, courier_id, user_id, route_id, cost, status, created_at, updated_at
`

type AssignRouteToTripParams struct {
	RouteID uuid.NullUUID
	ID      uuid.UUID
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

const assignTripToCourier = `-- name: AssignTripToCourier :one
UPDATE couriers
SET trip_id = $1
WHERE id = $2
RETURNING id, verified, status, location, rating, points, vehicle_id, user_id, trip_id, created_at, updated_at
`

type AssignTripToCourierParams struct {
	TripID uuid.NullUUID
	ID     uuid.UUID
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
		&i.VehicleID,
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
RETURNING id, verified, status, location, rating, points, vehicle_id, user_id, trip_id, created_at, updated_at
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
		&i.VehicleID,
		&i.UserID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createCourierUpload = `-- name: CreateCourierUpload :one
INSERT INTO uploads (
  type, uri, courier_id
) VALUES (
  $1, $2, $3
)
RETURNING id, type, uri, verified, courier_id, user_id, created_at, updated_at
`

type CreateCourierUploadParams struct {
	Type      string
	Uri       string
	CourierID uuid.NullUUID
}

func (q *Queries) CreateCourierUpload(ctx context.Context, arg CreateCourierUploadParams) (Upload, error) {
	row := q.db.QueryRowContext(ctx, createCourierUpload, arg.Type, arg.Uri, arg.CourierID)
	var i Upload
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Uri,
		&i.Verified,
		&i.CourierID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createRoute = `-- name: CreateRoute :one
INSERT INTO routes (
  distance, eta, polyline
) VALUES (
  $1, $2, ST_GeographyFromText($3)
)
RETURNING id, distance, polyline, eta, created_at, updated_at
`

type CreateRouteParams struct {
	Distance string
	Eta      time.Time
	Polyline interface{}
}

func (q *Queries) CreateRoute(ctx context.Context, arg CreateRouteParams) (Route, error) {
	row := q.db.QueryRowContext(ctx, createRoute, arg.Distance, arg.Eta, arg.Polyline)
	var i Route
	err := row.Scan(
		&i.ID,
		&i.Distance,
		&i.Polyline,
		&i.Eta,
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
	UserID        uuid.NullUUID
	StartLocation interface{}
	EndLocation   interface{}
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

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  first_name, last_name, phone
) VALUES (
  $1, $2, $3
)
RETURNING id, first_name, last_name, phone, onboarding, created_at, updated_at
`

type CreateUserParams struct {
	FirstName string
	LastName  string
	Phone     string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.FirstName, arg.LastName, arg.Phone)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Phone,
		&i.Onboarding,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createUserUpload = `-- name: CreateUserUpload :one
INSERT INTO uploads (
  type, uri, user_id
) VALUES (
  $1, $2, $3
)
RETURNING id, type, uri, verified, courier_id, user_id, created_at, updated_at
`

type CreateUserUploadParams struct {
	Type   string
	Uri    string
	UserID uuid.NullUUID
}

func (q *Queries) CreateUserUpload(ctx context.Context, arg CreateUserUploadParams) (Upload, error) {
	row := q.db.QueryRowContext(ctx, createUserUpload, arg.Type, arg.Uri, arg.UserID)
	var i Upload
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Uri,
		&i.Verified,
		&i.CourierID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createVehicle = `-- name: CreateVehicle :one
INSERT INTO vehicles (
  product_id
) VALUES (
  $1
)
RETURNING id, product_id, created_at, updated_at
`

func (q *Queries) CreateVehicle(ctx context.Context, productID uuid.UUID) (Vehicle, error) {
	row := q.db.QueryRowContext(ctx, createVehicle, productID)
	var i Vehicle
	err := row.Scan(
		&i.ID,
		&i.ProductID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findByPhone = `-- name: FindByPhone :one
SELECT id, first_name, last_name, phone, onboarding, created_at, updated_at FROM users
WHERE phone = $1
LIMIT 1
`

func (q *Queries) FindByPhone(ctx context.Context, phone string) (User, error) {
	row := q.db.QueryRowContext(ctx, findByPhone, phone)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Phone,
		&i.Onboarding,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourier = `-- name: GetCourier :one
SELECT id, verified, status, location, rating, points, vehicle_id, user_id, trip_id, created_at, updated_at FROM
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
		&i.VehicleID,
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

const getCourierUpload = `-- name: GetCourierUpload :one
SELECT id, type, uri, verified, courier_id, user_id, created_at, updated_at FROM
uploads
WHERE courier_id = $1 AND type = $2
LIMIT 1
`

type GetCourierUploadParams struct {
	CourierID uuid.NullUUID
	Type      string
}

func (q *Queries) GetCourierUpload(ctx context.Context, arg GetCourierUploadParams) (Upload, error) {
	row := q.db.QueryRowContext(ctx, getCourierUpload, arg.CourierID, arg.Type)
	var i Upload
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Uri,
		&i.Verified,
		&i.CourierID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourierUploads = `-- name: GetCourierUploads :many
SELECT id, type, uri, verified, courier_id, user_id, created_at, updated_at FROM uploads
WHERE courier_id = $1
`

func (q *Queries) GetCourierUploads(ctx context.Context, courierID uuid.NullUUID) ([]Upload, error) {
	rows, err := q.db.QueryContext(ctx, getCourierUploads, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Upload
	for rows.Next() {
		var i Upload
		if err := rows.Scan(
			&i.ID,
			&i.Type,
			&i.Uri,
			&i.Verified,
			&i.CourierID,
			&i.UserID,
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

const isUserOnboarding = `-- name: IsUserOnboarding :one
SELECT onboarding FROM
users
WHERE id = $1
LIMIT 1
`

func (q *Queries) IsUserOnboarding(ctx context.Context, id uuid.UUID) (bool, error) {
	row := q.db.QueryRowContext(ctx, isUserOnboarding, id)
	var onboarding bool
	err := row.Scan(&onboarding)
	return onboarding, err
}

const setCourierStatus = `-- name: SetCourierStatus :one
UPDATE couriers
SET status = $1
WHERE user_id = $2
RETURNING id, verified, status, location, rating, points, vehicle_id, user_id, trip_id, created_at, updated_at
`

type SetCourierStatusParams struct {
	Status string
	UserID uuid.NullUUID
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
		&i.VehicleID,
		&i.UserID,
		&i.TripID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const setOnboardingStatus = `-- name: SetOnboardingStatus :one
UPDATE users
SET onboarding = $1
WHERE phone = $2
RETURNING id, first_name, last_name, phone, onboarding, created_at, updated_at
`

type SetOnboardingStatusParams struct {
	Onboarding bool
	Phone      string
}

func (q *Queries) SetOnboardingStatus(ctx context.Context, arg SetOnboardingStatusParams) (User, error) {
	row := q.db.QueryRowContext(ctx, setOnboardingStatus, arg.Onboarding, arg.Phone)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Phone,
		&i.Onboarding,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateProductLocation = `-- name: UpdateProductLocation :one
UPDATE products
SET location = $1
WHERE id = $2
RETURNING id, name, description, location, created_at, updated_at
`

type UpdateProductLocationParams struct {
	Location interface{}
	ID       uuid.UUID
}

func (q *Queries) UpdateProductLocation(ctx context.Context, arg UpdateProductLocationParams) (Product, error) {
	row := q.db.QueryRowContext(ctx, updateProductLocation, arg.Location, arg.ID)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Location,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUpload = `-- name: UpdateUpload :one
UPDATE uploads
SET uri = COALESCE($1, uri)
WHERE id = $2
RETURNING id, type, uri, verified, courier_id, user_id, created_at, updated_at
`

type UpdateUploadParams struct {
	Uri string
	ID  uuid.UUID
}

func (q *Queries) UpdateUpload(ctx context.Context, arg UpdateUploadParams) (Upload, error) {
	row := q.db.QueryRowContext(ctx, updateUpload, arg.Uri, arg.ID)
	var i Upload
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Uri,
		&i.Verified,
		&i.CourierID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserName = `-- name: UpdateUserName :one
UPDATE users
SET first_name = COALESCE($1, first_name), last_name = COALESCE($2, last_name)
WHERE phone = $3
RETURNING id, first_name, last_name, phone, onboarding, created_at, updated_at
`

type UpdateUserNameParams struct {
	FirstName string
	LastName  string
	Phone     string
}

func (q *Queries) UpdateUserName(ctx context.Context, arg UpdateUserNameParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUserName, arg.FirstName, arg.LastName, arg.Phone)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Phone,
		&i.Onboarding,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
