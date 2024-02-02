package route

import "github.com/3dw1nM0535/uzi-api/gql/model"

type Route interface {
	ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error)
}
