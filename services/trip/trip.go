package trip

import "github.com/3dw1nM0535/uzi-api/model"

type Trip interface {
	ComputeTrip(input model.TripRouteInput) (*model.TripRoute, error)
}
