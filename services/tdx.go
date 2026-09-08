package services

import (
	"encoding/json"
	"fmt"
	"log" // 👈 確保有引入 log 套件
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
// 改成接收 city 參數，並支援動態代入網址
func (s *TDXService) GetParkingDataByCity(city string, token string) ([]models.UnifiedParking, error) {
	carParkURL := fmt.Sprintf("https://tdx.transportdata.tw/api/basic/v1/Parking/OffStreet/CarPark/City/%s?format=JSON", city)
	log.Printf("📌 [TDXService] Requesting CarParks URL: %s", carParkURL) // 👈 印出組好的網址

	carParks, err := s.fetchCarParks(carParkURL, token)
	if err != nil {
		log.Printf("❌ [TDXService] fetchCarParks Error: %v", err)
		return nil, err
	}
	log.Printf("📦 [TDXService] Fetched CarParks count: %d", len(carParks)) // 👈 印出抓到的靜態車位數量

	availURL := fmt.Sprintf("https://tdx.transportdata.tw/api/basic/v1/Parking/OffStreet/ParkingAvailability/City/%s?format=JSON", city)
	log.Printf("📌 [TDXService] Requesting AvailURL: %s", availURL)

	availMap, err := s.fetchAvailabilities(availURL, token)
	if err != nil {
		log.Printf("❌ [TDXService] fetchAvailabilities Error: %v", err)
		return nil, err
	}
	log.Printf("📦 [TDXService] Fetched Availabilities count: %d", len(availMap)) // 👈 印出即時車位對應數量

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