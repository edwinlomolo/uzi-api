package model

type RoutePoint struct {
	Location LatLng `json:"latLng"`
}

type Origin struct {
	RoutePoint `json:"location"`
}

type Destination struct {
	RoutePoint `json:"location"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RouteRequest struct {
	Origin                 `json:"origin"`
	Destination            `json:"destination"`
	TravelMode             string `json:"travelMode"`
	RoutePreference        string `json:"routePreference"`
	ComputeAlternateRoutes bool   `json:"computeAlternativeRoutes"`
	RouteModifiers         struct {
		AvoidTolls    bool `json:"avoidTolls"`
		AvoidHighways bool `json:"avoidHighways"`
		AvoidFerries  bool `json:"avoidFerries"`
	} `json:"routeModifiers"`
	Language string `json:"language"`
	Units    string `json:"units"`
}

type RouteResponse struct {
	Routes []struct {
		Distance int    `json:"distance"`
		Duration string `json:"duration"`
		Polyline struct {
			EncodedPolyline string `json:"encodedPolyline"`
		} `json:"polyline"`
	} `json:"routes"`
}
