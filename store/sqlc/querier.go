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
	AssignCourierToTrip(ctx context.Context, arg AssignCourierToTripParams) (Courier, error)
	AssignRouteToTrip(ctx context.Context, arg AssignRouteToTripParams) (Trip, error)
	AssignTripToCourier(ctx context.Context, arg AssignTripToCourierParams) (Trip, error)
	CreateCourier(ctx context.Context, userID uuid.NullUUID) (Courier, error)
	CreateCourierUpload(ctx context.Context, arg CreateCourierUploadParams) (Upload, error)
	CreateRecipient(ctx context.Context, arg CreateRecipientParams) (Recipient, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateTrip(ctx context.Context, arg CreateTripParams) (Trip, error)
	CreateTripCost(ctx context.Context, arg CreateTripCostParams) (Trip, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateUserUpload(ctx context.Context, arg CreateUserUploadParams) (Upload, error)
	FindAvailableCourier(ctx context.Context, arg FindAvailableCourierParams) (FindAvailableCourierRow, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetCourierAssignedTrip(ctx context.Context, id uuid.UUID) (Courier, error)
	GetCourierAvatar(ctx context.Context, courierID uuid.NullUUID) (GetCourierAvatarRow, error)
	GetCourierByID(ctx context.Context, id uuid.UUID) (GetCourierByIDRow, error)
	GetCourierByUserID(ctx context.Context, userID uuid.NullUUID) (GetCourierByUserIDRow, error)
	GetCourierLocation(ctx context.Context, id uuid.UUID) (interface{}, error)
	GetCourierNearPickupPoint(ctx context.Context, arg GetCourierNearPickupPointParams) ([]GetCourierNearPickupPointRow, error)
	GetCourierProductByID(ctx context.Context, id uuid.UUID) (GetCourierProductByIDRow, error)
	GetCourierStatus(ctx context.Context, userID uuid.NullUUID) (string, error)
	GetCourierTrip(ctx context.Context, courierID uuid.NullUUID) (Trip, error)
	GetCourierUpload(ctx context.Context, arg GetCourierUploadParams) (Upload, error)
	GetCourierUploads(ctx context.Context, courierID uuid.NullUUID) ([]Upload, error)
	GetNearbyAvailableCourierProducts(ctx context.Context, arg GetNearbyAvailableCourierProductsParams) ([]GetNearbyAvailableCourierProductsRow, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	GetTrip(ctx context.Context, id uuid.UUID) (GetTripRow, error)
	GetTripRecipient(ctx context.Context, tripID uuid.NullUUID) (Recipient, error)
	GetUserUpload(ctx context.Context, arg GetUserUploadParams) (Upload, error)
	IsCourier(ctx context.Context, userID uuid.NullUUID) (sql.NullBool, error)
	IsUserOnboarding(ctx context.Context, id uuid.UUID) (bool, error)
	SetCourierStatus(ctx context.Context, arg SetCourierStatusParams) (Courier, error)
	SetOnboardingStatus(ctx context.Context, arg SetOnboardingStatusParams) (User, error)
	SetTripStatus(ctx context.Context, arg SetTripStatusParams) (Trip, error)
	TrackCourierLocation(ctx context.Context, arg TrackCourierLocationParams) (Courier, error)
	UnassignCourierTrip(ctx context.Context, id uuid.UUID) (Courier, error)
	UpdateUpload(ctx context.Context, arg UpdateUploadParams) (Upload, error)
	UpdateUserName(ctx context.Context, arg UpdateUserNameParams) (User, error)
}

var _ Querier = (*Queries)(nil)
