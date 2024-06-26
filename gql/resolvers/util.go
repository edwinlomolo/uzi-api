package resolvers

import (
	"context"

	"github.com/edwinlomolo/uzi-api/controllers"
	"github.com/google/uuid"
)

func getCourierIDFromResolverContext(ctx context.Context) uuid.UUID {
	userID := ctx.Value("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		log.WithError(err).Errorf("get courier id from resolver ctx")
	}

	c, _ := controllers.GetCourierController().GetCourierByUserID(uid)
	return c.ID
}

func stringToUUID(id string) uuid.UUID {
	uid, err := uuid.Parse(id)
	if err != nil {
		log.WithError(err).Errorf("parse valid string uuid")
	}

	return uid
}
