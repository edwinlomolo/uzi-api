package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CreateCourierDocument is the resolver for the createCourierDocument field.
func (r *mutationResolver) CreateCourierDocument(ctx context.Context, input model.CourierUploadInput) (bool, error) {
	courierID := getCourierIDFromResolverContext(ctx)

	err := r.CreateCourierUpload(input.Type.String(), input.URI, courierID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// TrackCourierGps is the resolver for the trackCourierGps field.
func (r *mutationResolver) TrackCourierGps(ctx context.Context, input model.GpsInput) (bool, error) {
	userID := stringToUUID(ctx.Value("userID").(string))

	go r.TrackCourierLocation(userID, input)
	return true, nil
}

// SetCourierStatus is the resolver for the setCourierStatus field.
func (r *mutationResolver) SetCourierStatus(ctx context.Context, status string) (bool, error) {
	userID := stringToUUID(ctx.Value("userID").(string))

	s := model.CourierStatus(status)
	return r.UpdateCourierStatus(userID, s)
}

// CreateTrip is the resolver for the createTrip field.
func (r *mutationResolver) CreateTrip(ctx context.Context, input model.CreateTripInput) (*model.Trip, error) {
	userID := stringToUUID(ctx.Value("userID").(string))

	params := sqlc.CreateTripParams{
		UserID:    userID,
		ProductID: stringToUUID(input.TripProductID),
		StartLocation: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			input.TripInput.Pickup.Location.Lng,
			input.TripInput.Pickup.Location.Lat,
		),
		EndLocation: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			input.TripInput.Dropoff.Location.Lng,
			input.TripInput.Dropoff.Location.Lat,
		),
		ConfirmedPickup: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			input.ConfirmedPickup.Location.Lng,
			input.ConfirmedPickup.Location.Lat,
		),
	}
	pickup, pickupErr := r.tripService.ParsePickupDropoff(*input.TripInput.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}
	dropoff, dropErr := r.tripService.ParsePickupDropoff(*input.TripInput.Dropoff)
	if dropErr != nil {
		return nil, dropErr
	}
	params.StartLocation = fmt.Sprintf(
		"SRID=4326;POINT(%.8f %.8f)",
		pickup.Location.Lng,
		pickup.Location.Lat,
	)
	params.EndLocation = fmt.Sprintf(
		"SRID=4326;POINT(%.8f %.8f)",
		dropoff.Location.Lng,
		dropoff.Location.Lat,
	)

	trip, err := r.tripService.CreateTrip(params)

	go func() {
		err := r.tripService.CreateTripRecipient(trip.ID, *input.Recipient)
		if err != nil {
			return
		}
	}()

	r.tripService.MatchCourier(trip.ID, *input.TripInput.Pickup)

	return trip, err
}

// ReportTripStatus is the resolver for the reportTripStatus field.
func (r *mutationResolver) ReportTripStatus(ctx context.Context, tripID uuid.UUID, status model.TripStatus) (bool, error) {
	err := r.tripService.ReportTripStatus(tripID, status)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Hello is the resolver for the hello field.
func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, world!", nil
}

// GetCourierDocuments is the resolver for the getCourierDocuments field.
func (r *queryResolver) GetCourierDocuments(ctx context.Context) ([]*model.Uploads, error) {
	courierID := getCourierIDFromResolverContext(ctx)

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
func (r *queryResolver) ComputeTripRoute(ctx context.Context, input model.TripRouteInput) (*model.TripRoute, error) {
	return r.tripService.ComputeTripRoute(input)
}

// GetCourierNearPickupPoint is the resolver for the getCourierNearPickupPoint field.
func (r *queryResolver) GetCourierNearPickupPoint(ctx context.Context, point model.GpsInput) ([]*model.Courier, error) {
	return r.tripService.GetCourierNearPickupPoint(point)
}

// GetTripDetails is the resolver for the getTripDetails field.
func (r *queryResolver) GetTripDetails(ctx context.Context, tripID uuid.UUID) (*model.Trip, error) {
	return r.tripService.GetTripDetails(tripID)
}

// TripUpdates is the resolver for the tripUpdates field.
func (r *subscriptionResolver) TripUpdates(ctx context.Context, tripID uuid.UUID) (<-chan *model.TripUpdate, error) {
	pubsub := r.redisClient.Subscribe(context.Background(), internal.TRIP_UPDATES_CHANNEL)

	ch := make(chan *model.TripUpdate)

	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				log.WithFields(logrus.Fields{
					"error":   err,
					"trip_id": tripID,
				}).Errorf("receive trip update")
				close(ch)
				return
			}

			var update *model.TripUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				log.WithError(err).Errorf("unmarshal redis trip update payload")
				return
			}
			if update.ID == tripID {
				ch <- update
			}
		}
	}()

	return ch, nil
}

// AssignTrip is the resolver for the assignTrip field.
func (r *subscriptionResolver) AssignTrip(ctx context.Context, userID uuid.UUID) (<-chan *model.TripUpdate, error) {
	c, err := r.GetCourierByUserID(userID)
	if err != nil {
		return nil, err
	}

	pubsub := r.redisClient.Subscribe(context.Background(), internal.ASSIGN_TRIP_CHANNEL)

	ch := make(chan *model.TripUpdate)

	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				log.WithFields(logrus.Fields{
					"error":   err,
					"user_id": userID,
				}).Errorf("trip assignment update")
				close(ch)
				return
			}

			var update *model.TripUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				log.WithError(err).Errorf("unmarshal redis trip assignment update payload")
				return
			}
			if *update.CourierID == c.ID {
				ch <- update
			}
		}
	}()

	return ch, nil
}

// Mutation returns gql.MutationResolver implementation.
func (r *Resolver) Mutation() gql.MutationResolver { return &mutationResolver{r} }

// Query returns gql.QueryResolver implementation.
func (r *Resolver) Query() gql.QueryResolver { return &queryResolver{r} }

// Subscription returns gql.SubscriptionResolver implementation.
func (r *Resolver) Subscription() gql.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
