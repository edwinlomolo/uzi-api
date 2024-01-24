//go:generate go run github.com/99designs/gqlgen generate --verbose
package resolvers

import (
	"github.com/3dw1nM0535/uzi-api/gql"
	"github.com/3dw1nM0535/uzi-api/services/courier"
	"github.com/3dw1nM0535/uzi-api/services/location"
	"github.com/3dw1nM0535/uzi-api/services/route"
	"github.com/3dw1nM0535/uzi-api/services/upload"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	upload.Upload
	courier.Courier
	location.Location
	route.Route
}

func New() gql.Config {
	c := gql.Config{Resolvers: &Resolver{
		upload.GetUploadService(),
		courier.GetCourierService(),
		location.GetLocationService(),
		route.GetRouteService(),
	}}

	return c
}
