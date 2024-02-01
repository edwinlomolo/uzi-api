// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package sqlc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Querier interface {
	AssignRouteToTrip(ctx context.Context, arg AssignRouteToTripParams) (Trip, error)
	AssignTripToCourier(ctx context.Context, arg AssignTripToCourierParams) (Courier, error)
	CreateCourier(ctx context.Context, userID uuid.NullUUID) (Courier, error)
	CreateCourierUpload(ctx context.Context, arg CreateCourierUploadParams) (Upload, error)
	CreateRoute(ctx context.Context, arg CreateRouteParams) (Route, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateTrip(ctx context.Context, arg CreateTripParams) (Trip, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateUserUpload(ctx context.Context, arg CreateUserUploadParams) (Upload, error)
	CreateVehicle(ctx context.Context, arg CreateVehicleParams) (Vehicle, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetCourier(ctx context.Context, userID uuid.NullUUID) (Courier, error)
	GetCourierNearPickupPoint(ctx context.Context, arg GetCourierNearPickupPointParams) ([]GetCourierNearPickupPointRow, error)
	GetCourierProductByID(ctx context.Context, id uuid.UUID) (GetCourierProductByIDRow, error)
	GetCourierStatus(ctx context.Context, userID uuid.NullUUID) (string, error)
	GetCourierUpload(ctx context.Context, arg GetCourierUploadParams) (Upload, error)
	GetCourierUploads(ctx context.Context, courierID uuid.NullUUID) ([]Upload, error)
	GetNearbyAvailableCourierProducts(ctx context.Context, arg GetNearbyAvailableCourierProductsParams) ([]GetNearbyAvailableCourierProductsRow, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	IsCourier(ctx context.Context, userID uuid.NullUUID) (sql.NullBool, error)
	IsUserOnboarding(ctx context.Context, id uuid.UUID) (bool, error)
	SetCourierStatus(ctx context.Context, arg SetCourierStatusParams) (Courier, error)
	SetOnboardingStatus(ctx context.Context, arg SetOnboardingStatusParams) (User, error)
	TrackCourierLocation(ctx context.Context, arg TrackCourierLocationParams) (Courier, error)
	UpdateUpload(ctx context.Context, arg UpdateUploadParams) (Upload, error)
	UpdateUserName(ctx context.Context, arg UpdateUserNameParams) (User, error)
}

var _ Querier = (*Queries)(nil)
