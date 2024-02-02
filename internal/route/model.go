package route

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
	Lat float64 `json:"latitude"`
	Lng float64 `json:"longitude"`
}

type RouteModifiers struct {
	AvoidTolls    bool `json:"avoidTolls"`
	AvoidHighways bool `json:"avoidHighways"`
	AvoidFerries  bool `json:"avoidFerries"`
}

type RouteRequest struct {
	Origin                 `json:"origin"`
	Destination            `json:"destination"`
	TravelMode             string         `json:"travelMode"`
	RoutePreference        string         `json:"routingPreference"`
	ComputeAlternateRoutes bool           `json:"computeAlternativeRoutes"`
	RouteModifiers         RouteModifiers `json:"routeModifiers"`
	PolylineQuality        string         `json:"polylineQuality"`
	Language               string         `json:"languageCode"`
	Units                  string         `json:"units"`
	RegionCode             string         `json:"regionCode"`
}

type Polyline struct {
	EncodedPolyline string `json:"encodedPolyline"`
}

type Routes struct {
	Distance       int      `json:"distanceMeters"`
	Duration       string   `json:"duration"`
	StaticDuration string   `json:"staticDuration"`
	Polyline       Polyline `json:"polyline"`
}

type RouteResponse struct {
	Error  RouteError `json:"error"`
	Routes []Routes   `json:"routes"`
}

type RouteError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
