package models

type ScooterRouteRequest struct {
	OriginLat float64 `json:"origin_lat"`
	OriginLng float64 `json:"origin_lng"`
	DestLat   float64 `json:"dest_lat"`
	DestLng   float64 `json:"dest_lng"`
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ScooterRouteResponse struct {
	Status string       `json:"status"`
	Points []Coordinate `json:"points"`
}