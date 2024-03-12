package handler

import (
	"context"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/edwinlomolo/uzi-api/internal"
)

func WsInit(
	ctx context.Context,
	initPayload transport.InitPayload,
) (context.Context, *transport.InitPayload, error) {
	log := internal.GetLogger()
	any, ok := initPayload["payload"].(map[string]interface{})
	if !ok {
		log.Warnln("payload not passed in request")
	}

	payload := any["headers"].(map[string]interface{})
	auth := payload["Authorization"].(string)

	ctxNew := context.WithValue(ctx, "auth", auth)

	return ctxNew, nil, nil
}
