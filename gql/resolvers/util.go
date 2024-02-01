package resolvers

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/services/courier"
	"github.com/google/uuid"
)

func GetCourierIDFromRequestContext(ctx context.Context, courier courier.CourierService) uuid.UUID {
	userID := ctx.Value("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		panic(err)
	}

	c, _ := courier.GetCourier(uid)
	return c.ID
}
