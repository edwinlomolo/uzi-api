package uzi

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services"
	"github.com/3dw1nM0535/uzi-api/store"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	userService services.UserService
}

func New(store *store.Queries, redis *redis.Client, logger *logrus.Logger) Config {
	c := Config{Resolvers: &Resolver{
		services.NewUserService(store, redis, logger),
	}}

	return c
}

func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, world", nil
}

func (r *mutationResolver) SignIn(ctx context.Context, input model.SigninInput) (*model.User, error) {
	newUser, err := r.userService.FindOrCreate(input)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
