// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Courier struct {
	ID        uuid.UUID
	Verified  sql.NullBool
	Status    string
	Location  interface{}
	Rating    float64
	Points    int32
	VehicleID uuid.NullUUID
	UserID    uuid.NullUUID
	TripID    uuid.NullUUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Location    interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Route struct {
	ID        uuid.UUID
	Distance  string
	Polyline  interface{}
	Eta       time.Time
	TripID    uuid.NullUUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Session struct {
	ID        uuid.UUID
	Ip        string
	Token     string
	Expires   time.Time
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Trip struct {
	ID            uuid.UUID
	StartLocation interface{}
	EndLocation   interface{}
	CourierID     uuid.NullUUID
	UserID        uuid.NullUUID
	RouteID       uuid.NullUUID
	Cost          sql.NullString
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Upload struct {
	ID        uuid.UUID
	Type      sql.NullString
	Uri       string
	CourierID uuid.NullUUID
	UserID    uuid.NullUUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Vehicle struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	CourierID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
