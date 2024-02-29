package resolvers

import (
	"context"

	"github.com/edwinlomolo/uzi-api/courier"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/google/uuid"
)

func getCourierIDFromResolverContext(ctx context.Context, courier courier.CourierService) uuid.UUID {
	userID := ctx.Value("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		logger.Logger.WithError(err).Errorf("get courier id from resolver ctx")
	}

	c, _ := courier.GetCourierByUserID(uid)
	return c.ID
}

func stringToUUID(id string) uuid.UUID {
	uid, err := uuid.Parse(id)
	if err != nil {
		logger.Logger.WithError(err).Errorf("parse valid string uuid")
	}

	return uid
}
