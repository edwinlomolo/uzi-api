package uzi

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services"
	"github.com/3dw1nM0535/uzi-api/store"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	userService    services.UserService
	sessionService services.Session
}

func New(store *store.Queries, redis *redis.Client, logger *logrus.Logger, config *config.Configuration) Config {
	c := Config{Resolvers: &Resolver{
		services.NewUserService(store, redis, logger),
		services.NewSessionService(store, logger, config.Jwt),
	}}

	return c
}

func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, world", nil
}

func (r *mutationResolver) SignIn(ctx context.Context, input model.SigninInput) (*model.Session, error) {
	newUser, newUserErr := r.userService.FindOrCreate(input)
	if newUserErr != nil {
		return nil, newUserErr
	}

	ip, ok := netip.AddrFromSlice([]byte(ctx.Value("ip").(string)))
	if !ok {
		return nil, fmt.Errorf("Error parsing ip from context")
	}

	newSession, newSessionErr := r.sessionService.FindOrCreate(newUser.ID, ip)
	if newSessionErr != nil {
		return nil, newSessionErr
	}

	return newSession, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
