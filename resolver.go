package uzi

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/model"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func New() Config {
	c := Config{Resolvers: &Resolver{}}

	return c
}

func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, world", nil
}

func (r *queryResolver) GetCourierDocuments(ctx context.Context) ([]*model.Uploads, error) {
	panic("not implemented")
}

func (r *mutationResolver) CreateCourierDocument(ctx context.Context) (*model.Uploads, error) {
	panic("not implemented")
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
