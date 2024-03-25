package controllers

type routepoint struct {
	Location latlng `json:"latLng"`
}

type origin struct {
	routepoint `json:"location"`
}

type destination struct {
	routepoint `json:"location"`
}

type latlng struct {
	Lat float64 `json:"latitude"`
	Lng float64 `json:"longitude"`
}

type routemodifiers struct {
	AvoidTolls    bool `json:"avoidTolls"`
	AvoidHighways bool `json:"avoidHighways"`
	AvoidFerries  bool `json:"avoidFerries"`
}

type routerequest struct {
	origin                 `json:"origin"`
	destination            `json:"destination"`
	TravelMode             string         `json:"travelMode"`
	RoutePreference        string         `json:"routingPreference"`
	ComputeAlternateRoutes bool           `json:"computeAlternativeRoutes"`
	RouteModifiers         routemodifiers `json:"routeModifiers"`
	PolylineQuality        string         `json:"polylineQuality"`
	Language               string         `json:"languageCode"`
	Units                  string         `json:"units"`
	RegionCode             string         `json:"regionCode"`
}

type polyline struct {
	EncodedPolyline string `json:"encodedPolyline"`
}

type routes struct {
	Distance       int      `json:"distanceMeters"`
	Duration       string   `json:"duration"`
	StaticDuration string   `json:"staticDuration"`
	Polyline       polyline `json:"polyline"`
}

type routeresponse struct {
	Error  routeerror `json:"error"`
	Routes []routes   `json:"routes"`
}

type routeerror struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
