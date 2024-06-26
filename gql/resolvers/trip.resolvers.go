package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"

	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/gql/model"
)

// Trip is the resolver for the trip field.
func (r *recipientResolver) Trip(ctx context.Context, obj *model.Recipient) (*model.Trip, error) {
	return r.tripController.GetTripDetails(obj.TripID)
}

// Courier is the resolver for the courier field.
func (r *tripResolver) Courier(ctx context.Context, obj *model.Trip) (*model.Courier, error) {
	return r.GetCourierByID(*obj.CourierID)
}

// Recipient is the resolver for the recipient field.
func (r *tripResolver) Recipient(ctx context.Context, obj *model.Trip) (*model.Recipient, error) {
	return r.tripController.GetTripRecipient(obj.ID)
}

// Recipient returns gql.RecipientResolver implementation.
func (r *Resolver) Recipient() gql.RecipientResolver { return &recipientResolver{r} }

// Trip returns gql.TripResolver implementation.
func (r *Resolver) Trip() gql.TripResolver { return &tripResolver{r} }

type recipientResolver struct{ *Resolver }
type tripResolver struct{ *Resolver }
