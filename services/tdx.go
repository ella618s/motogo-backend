package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"motogo-backend/models"
)

type TDXService struct {
	client *http.Client
}

func NewTDXService() *TDXService {
	return &TDXService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// 取得 TDX Access Token
func (s *TDXService) GetAccessToken() (string, error) {
	clientID := os.Getenv("TDX_CLIENT_ID")
	clientSecret := os.Getenv("TDX_CLIENT_SECRET")

	tokenURL := "https://tdx.transportdata.tw/auth/realms/TDXConnect/protocol/openid-connect/token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tdx auth failed with status: %d", resp.StatusCode)
	}

	var tokenResp models.TDXTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// 取得台北市整合停車位資料
func (s *TDXService) GetTaipeiParkingData(token string) ([]models.UnifiedParking, error) {
	// 1. 取得靜態資料
	carParkURL := "https://tdx.transportdata.tw/api/basic/v1/Parking/OffStreet/CarPark/City/Taipei?$format=JSON"
	carParks, err := s.fetchCarParks(carParkURL, token)
	if err != nil {
		return nil, err
	}

	// 2. 取得動態剩餘車位
	availURL := "https://tdx.transportdata.tw/api/basic/v1/Parking/OffStreet/ParkingAvailability/City/Taipei?$format=JSON"
	availMap, err := s.fetchAvailabilities(availURL, token)
	if err != nil {
		return nil, err
	}

	// 3. 資料組合 (Data Fusion)
	var unifiedList []models.UnifiedParking
	for _, cp := range carParks {
		avail := availMap[cp.CarParkID]
		unifiedList = append(unifiedList, models.UnifiedParking{
			ID:              cp.CarParkID,
			Name:            cp.CarParkName.ZhTw,
			Lat:             cp.CarParkPosition.PositionLat,
			Lng:             cp.CarParkPosition.PositionLon,
			Address:         cp.Address,
			AvailableSpaces: avail,
		})
	}

	return unifiedList, nil
}

func (s *TDXService) fetchCarParks(apiURL, token string) ([]models.CarPark, error) {
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res models.TDXCarParkResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.CarParks, nil
}

func (s *TDXService) fetchAvailabilities(apiURL, token string) (map[string]int, error) {
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res models.TDXAvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	availMap := make(map[string]int)
	for _, a := range res.ParkingAvailabilities {
		availMap[a.CarParkID] = a.AvailableSpaces
	}
	return availMap, nil
}