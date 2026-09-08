package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"motogo-backend/models"
)

type RouteService struct {
	client *http.Client
}

func NewRouteService() *RouteService {
	return &RouteService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetScooterRoute 計算避開快速道路的機車路徑
func (s *RouteService) GetScooterRoute(req models.ScooterRouteRequest) ([]models.Coordinate, error) {
	// 範例使用 OSRM 公共 API（實際部署可替換為自架 OSRM 服務並調整 profile）
	// OSRM 格式: /route/v1/driving/{lng1},{lat1};{lng2},{lat2}?overview=full&geometries=geojson
	baseURL := fmt.Sprintf("https://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f",
		req.OriginLng, req.OriginLat, req.DestLng, req.DestLat,
	)

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	query := parsedURL.Query()
	query.Set("overview", "full")
	query.Set("geometries", "geojson")
	// 註：若使用自架 OSRM，可在此加入過濾或自定義 scooter profile 參數
	parsedURL.RawQuery = query.Encode()

	log.Printf("📌 [RouteService] Requesting OSRM URL: %s", parsedURL.String())

	resp, err := s.client.Get(parsedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var osrmResp struct {
		Code   string `json:"code"`
		Routes []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"` // [lng, lat]
			} `json:"geometry"`
		} `json:"routes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return nil, err
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("failed to find route, code: %s", osrmResp.Code)
	}

	var coordinates []models.Coordinate
	for _, coord := range osrmResp.Routes[0].Geometry.Coordinates {
		if len(coord) >= 2 {
			coordinates = append(coordinates, models.Coordinate{
				Lng: coord[0],
				Lat: coord[1],
			})
		}
	}

	return coordinates, nil
}