package route

import "github.com/3dw1nM0535/uzi-api/model"

type Route interface {
	GetTripRoute(input model.TripRouteInput) (*model.TripRoute, error)
}
