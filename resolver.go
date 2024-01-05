package uzi

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services"
	"github.com/google/uuid"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	services.Upload
}

func New() Config {
	c := Config{Resolvers: &Resolver{
		services.GetUploadService(),
	}}

	return c
}

func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, world", nil
}

func (r *queryResolver) GetCourierDocuments(ctx context.Context) ([]*model.Uploads, error) {
	courierID := GetCourierIDFromRequestContext(ctx)

	uploads, err := r.GetCourierUploads(courierID)
	if err != nil {
		return nil, err
	}

	return uploads, nil
}

func (r *mutationResolver) CreateCourierDocument(ctx context.Context, doc model.CourierUploadInput) (bool, error) {
	courierID := GetCourierIDFromRequestContext(ctx)

	err := r.CreateCourierUpload(doc.Type.String(), doc.URI, courierID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetCourierIDFromRequestContext(ctx context.Context) uuid.UUID {
	userID := ctx.Value("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		panic(err)
	}

	courier, _ := services.GetCourierService().GetCourier(uid)
	return courier.ID
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
