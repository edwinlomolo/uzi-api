package resolvers

import (
	"context"

	"github.com/edwinlomolo/uzi-api/services/courier"
	"github.com/google/uuid"
)

func getCourierIDFromResolverContext(ctx context.Context, courier courier.CourierService) uuid.UUID {
	userID := ctx.Value("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		panic(err)
	}

	c, _ := courier.GetCourierByUserID(uid)
	return c.ID
}

func stringToUUID(id string) uuid.UUID {
	uid, err := uuid.Parse(id)
	if err != nil {
		panic(err)
	}

	return uid
}
