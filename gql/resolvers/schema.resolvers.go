package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.41

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/gql"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/google/uuid"
)

// CreateCourierDocument is the resolver for the createCourierDocument field.
func (r *mutationResolver) CreateCourierDocument(ctx context.Context, input model.CourierUploadInput) (bool, error) {
	courierID := GetCourierIDFromRequestContext(ctx, r)

	err := r.CreateCourierUpload(input.Type.String(), input.URI, courierID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// TrackCourierGps is the resolver for the trackCourierGps field.
func (r *mutationResolver) TrackCourierGps(ctx context.Context, input model.GpsInput) (bool, error) {
	userID := ctx.Value("userID").(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		panic(err)
	}

	return r.TrackCourierLocation(uid, input)
}

// SetCourierStatus is the resolver for the setCourierStatus field.
func (r *mutationResolver) SetCourierStatus(ctx context.Context, status string) (bool, error) {
	userID := ctx.Value("userID").(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		panic(err)
	}

	s := model.CourierStatus(status)
	return r.UpdateCourierStatus(uid, s)
}

// Hello is the resolver for the hello field.
func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, world!", nil
}

// GetCourierDocuments is the resolver for the getCourierDocuments field.
func (r *queryResolver) GetCourierDocuments(ctx context.Context) ([]*model.Uploads, error) {
	courierID := GetCourierIDFromRequestContext(ctx, r)

	uploads, err := r.GetCourierUploads(courierID)
	if err != nil {
		return nil, err
	}

	return uploads, nil
}

// SearchPlace is the resolver for the searchPlace field.
func (r *queryResolver) SearchPlace(ctx context.Context, textQuery string) ([]*model.Place, error) {
	return r.AutocompletePlace(textQuery)
}

// ReverseGeocode is the resolver for the reverseGeocode field.
func (r *queryResolver) ReverseGeocode(ctx context.Context, place model.GpsInput) (*model.Geocode, error) {
	return r.GeocodeLatLng(place)
}

// GetRoute is the resolver for the getRoute field.
func (r *queryResolver) MakeTripRoute(ctx context.Context, input model.TripRouteInput) (*model.TripRoute, error) {
	return r.GetTripRoute(input)
}

// Mutation returns gql.MutationResolver implementation.
func (r *Resolver) Mutation() gql.MutationResolver { return &mutationResolver{r} }

// Query returns gql.QueryResolver implementation.
func (r *Resolver) Query() gql.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
