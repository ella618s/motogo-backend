package models

// TDX Token 回應結構
type TDXTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// 靜態停車場資料 (CarPark)
type CarPark struct {
	CarParkID   string `json:"CarParkID"`
	CarParkName struct {
		ZhTw string `json:"Zh_tw"`
	} `json:"CarParkName"`
	CarParkPosition struct {
		PositionLat float64 `json:"PositionLat"`
		PositionLon float64 `json:"PositionLon"`
	} `json:"CarParkPosition"`
	Address string `json:"Address"`
}

type TDXCarParkResponse struct {
	CarParks []CarPark `json:"CarParks"`
}

// 動態剩餘車位 (ParkingAvailability)
type ParkingAvailability struct {
	CarParkID       string `json:"CarParkID"`
	AvailableSpaces int    `json:"AvailableSpaces"`
}

type TDXAvailabilityResponse struct {
	ParkingAvailabilities []ParkingAvailability `json:"ParkingAvailabilities"`
}

// 對外統一輸出的 API 格式（提供給 Mobile App 與 Web 使用）
type UnifiedParking struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	Address         string  `json:"address"`
	AvailableSpaces int     `json:"available_spaces"`
}