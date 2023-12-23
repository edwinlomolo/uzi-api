package uzi

import (
	"context"
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

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation
//func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
