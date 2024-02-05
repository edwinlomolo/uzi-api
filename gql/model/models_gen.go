// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Courier struct {
	ID             uuid.UUID     `json:"id"`
	UserID         uuid.UUID     `json:"user_id"`
	Verified       bool          `json:"verified"`
	Status         CourierStatus `json:"status"`
	Rating         float64       `json:"rating"`
	Location       *Gps          `json:"location"`
	TripID         *uuid.UUID    `json:"trip_id,omitempty"`
	ProductID      uuid.UUID     `json:"product_id"`
	Product        *Product      `json:"product,omitempty"`
	CompletedTrips int           `json:"completedTrips"`
	Points         int           `json:"points"`
	UploadID       *uuid.UUID    `json:"upload_id,omitempty"`
	CreatedAt      *time.Time    `json:"created_at,omitempty"`
	UpdatedAt      *time.Time    `json:"updated_at,omitempty"`
}

type CourierUploadInput struct {
	Type UploadFile `json:"type"`
	URI  string     `json:"uri"`
}

type CreateTripInput struct {
	TripInput   *TripInput `json:"tripInput"`
	TripProduct string     `json:"tripProduct"`
}

type Gps struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type GpsInput struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Place struct {
	ID            string `json:"id"`
	MainText      string `json:"mainText"`
	SecondaryText string `json:"secondaryText"`
}

type Product struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IconURL     string     `json:"icon_url"`
	Price       int        `json:"price"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type Route struct {
	ID        uuid.UUID  `json:"id"`
	Distance  string     `json:"distance"`
	Eta       time.Time  `json:"eta"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type Session struct {
	ID            uuid.UUID      `json:"id"`
	IP            string         `json:"ip"`
	UserAgent     string         `json:"userAgent"`
	Token         string         `json:"token"`
	IsCourier     bool           `json:"isCourier"`
	Onboarding    bool           `json:"onboarding"`
	Phone         string         `json:"phone"`
	CourierStatus *CourierStatus `json:"courierStatus,omitempty"`
}

type Trip struct {
	ID            uuid.UUID  `json:"id"`
	CourierID     uuid.UUID  `json:"courier_id"`
	UserID        uuid.UUID  `json:"user_id"`
	StartLocation *Gps       `json:"start_location"`
	EndLocation   *Gps       `json:"end_location"`
	Status        TripStatus `json:"status"`
	Cost          *string    `json:"cost,omitempty"`
	RouteID       *uuid.UUID `json:"route_id,omitempty"`
	Route         *Route     `json:"route,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type TripInput struct {
	PlaceID          string    `json:"placeId"`
	FormattedAddress string    `json:"formattedAddress"`
	Location         *GpsInput `json:"location"`
}

type TripRoute struct {
	Polyline          string     `json:"polyline"`
	Distance          int        `json:"distance"`
	AvailableProducts []*Product `json:"availableProducts"`
}

type TripRouteInput struct {
	Pickup  *TripInput `json:"pickup"`
	Dropoff *TripInput `json:"dropoff"`
}

type TripUpdates struct {
	Status TripStatus `json:"status"`
}

type UnixTime struct {
	UnixTime  int    `json:"unixTime"`
	TimeStamp string `json:"timeStamp"`
}

type Uploads struct {
	ID           uuid.UUID                `json:"ID"`
	Type         string                   `json:"type"`
	URI          string                   `json:"uri"`
	Verification UploadVerificationStatus `json:"verification"`
	CourierID    *uuid.UUID               `json:"courier_id,omitempty"`
	UserID       *uuid.UUID               `json:"user_id,omitempty"`
	CreatedAt    *time.Time               `json:"created_at,omitempty"`
	UpdatedAt    *time.Time               `json:"updated_at,omitempty"`
}

type User struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Phone     string     `json:"phone"`
	CourierID *uuid.UUID `json:"courier_id,omitempty"`
	Courier   *Courier   `json:"courier,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CourierStatus string

const (
	CourierStatusOffline    CourierStatus = "OFFLINE"
	CourierStatusOnline     CourierStatus = "ONLINE"
	CourierStatusOnboarding CourierStatus = "ONBOARDING"
)

var AllCourierStatus = []CourierStatus{
	CourierStatusOffline,
	CourierStatusOnline,
	CourierStatusOnboarding,
}

func (e CourierStatus) IsValid() bool {
	switch e {
	case CourierStatusOffline, CourierStatusOnline, CourierStatusOnboarding:
		return true
	}
	return false
}

func (e CourierStatus) String() string {
	return string(e)
}

func (e *CourierStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = CourierStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid CourierStatus", str)
	}
	return nil
}

func (e CourierStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TripStatus string

const (
	TripStatusArriving TripStatus = "ARRIVING"
	TripStatusEnRoute  TripStatus = "EN_ROUTE"
	TripStatusComplete TripStatus = "COMPLETE"
)

var AllTripStatus = []TripStatus{
	TripStatusArriving,
	TripStatusEnRoute,
	TripStatusComplete,
}

func (e TripStatus) IsValid() bool {
	switch e {
	case TripStatusArriving, TripStatusEnRoute, TripStatusComplete:
		return true
	}
	return false
}

func (e TripStatus) String() string {
	return string(e)
}

func (e *TripStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TripStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TripStatus", str)
	}
	return nil
}

func (e TripStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type UploadFile string

const (
	UploadFileDp  UploadFile = "DP"
	UploadFileMcr UploadFile = "MCR"
	UploadFileID  UploadFile = "ID"
	UploadFilePc  UploadFile = "PC"
	UploadFileLb  UploadFile = "LB"
	UploadFileVi  UploadFile = "VI"
)

var AllUploadFile = []UploadFile{
	UploadFileDp,
	UploadFileMcr,
	UploadFileID,
	UploadFilePc,
	UploadFileLb,
	UploadFileVi,
}

func (e UploadFile) IsValid() bool {
	switch e {
	case UploadFileDp, UploadFileMcr, UploadFileID, UploadFilePc, UploadFileLb, UploadFileVi:
		return true
	}
	return false
}

func (e UploadFile) String() string {
	return string(e)
}

func (e *UploadFile) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UploadFile(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UploadFile", str)
	}
	return nil
}

func (e UploadFile) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type UploadVerificationStatus string

const (
	UploadVerificationStatusVerifying  UploadVerificationStatus = "VERIFYING"
	UploadVerificationStatusVerified   UploadVerificationStatus = "VERIFIED"
	UploadVerificationStatusRejected   UploadVerificationStatus = "REJECTED"
	UploadVerificationStatusOnboarding UploadVerificationStatus = "ONBOARDING"
)

var AllUploadVerificationStatus = []UploadVerificationStatus{
	UploadVerificationStatusVerifying,
	UploadVerificationStatusVerified,
	UploadVerificationStatusRejected,
	UploadVerificationStatusOnboarding,
}

func (e UploadVerificationStatus) IsValid() bool {
	switch e {
	case UploadVerificationStatusVerifying, UploadVerificationStatusVerified, UploadVerificationStatusRejected, UploadVerificationStatusOnboarding:
		return true
	}
	return false
}

func (e UploadVerificationStatus) String() string {
	return string(e)
}

func (e *UploadVerificationStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UploadVerificationStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UploadVerificationStatus", str)
	}
	return nil
}

func (e UploadVerificationStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
