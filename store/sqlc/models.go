// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package sqlc

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Courier struct {
	ID        uuid.UUID     `json:"id"`
	Verified  sql.NullBool  `json:"verified"`
	Status    string        `json:"status"`
	Location  interface{}   `json:"location"`
	Ratings   int32         `json:"ratings"`
	Points    int32         `json:"points"`
	UserID    uuid.NullUUID `json:"user_id"`
	ProductID uuid.NullUUID `json:"product_id"`
	TripID    uuid.NullUUID `json:"trip_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	WeightClass int32     `json:"weight_class"`
	Icon        string    `json:"icon"`
	Relevance   int32     `json:"relevance"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Recipient struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Building  sql.NullString `json:"building"`
	Unit      sql.NullString `json:"unit"`
	Phone     string         `json:"phone"`
	TripID    uuid.NullUUID  `json:"trip_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Route struct {
	ID        uuid.UUID     `json:"id"`
	Distance  string        `json:"distance"`
	Polyline  interface{}   `json:"polyline"`
	Eta       time.Time     `json:"eta"`
	State     string        `json:"state"`
	TripID    uuid.NullUUID `json:"trip_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Session struct {
	ID        uuid.UUID `json:"id"`
	Ip        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Trip struct {
	ID            uuid.UUID      `json:"id"`
	StartLocation interface{}    `json:"start_location"`
	EndLocation   interface{}    `json:"end_location"`
	CourierID     uuid.NullUUID  `json:"courier_id"`
	UserID        uuid.UUID      `json:"user_id"`
	RouteID       uuid.NullUUID  `json:"route_id"`
	ProductID     uuid.UUID      `json:"product_id"`
	Cost          sql.NullString `json:"cost"`
	Status        string         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type Upload struct {
	ID           uuid.UUID     `json:"id"`
	Type         string        `json:"type"`
	Uri          string        `json:"uri"`
	Verification string        `json:"verification"`
	CourierID    uuid.NullUUID `json:"courier_id"`
	UserID       uuid.NullUUID `json:"user_id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type User struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Phone      string    `json:"phone"`
	Onboarding bool      `json:"onboarding"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Vehicle struct {
	ID        uuid.UUID     `json:"id"`
	Mass      float64       `json:"mass"`
	ProductID uuid.UUID     `json:"product_id"`
	CourierID uuid.NullUUID `json:"courier_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
