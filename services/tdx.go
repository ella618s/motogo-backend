package services

import (
	"encoding/json"
	"fmt"
	"log" 
	"sync"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"motogo-backend/models"
)

type CacheItem struct {
	Data      []models.UnifiedParking
	ExpiresAt time.Time
}

type TDXService struct {
	client *http.Client
	cache  sync.Map // 👈 記憶體快取
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
	// 1. 檢查快取是否存在且未過期 (有效期設定為 10 分鐘)
	if val, ok := s.cache.Load(city); ok {
		item := val.(CacheItem)
		if time.Now().Before(item.ExpiresAt) && len(item.Data) > 0 {
			log.Printf("⚡ [TDXService] Returning valid cached data for city: %s (count: %d)", city, len(item.Data))
			return item.Data, nil
		}
	}

	carParkURL := fmt.Sprintf("https://tdx.transportdata.tw/api/basic/v1/Parking/OffStreet/CarPark/City/%s?format=JSON", city)
	log.Printf("📌 [TDXService] Requesting CarParks URL: %s", carParkURL)

	carParks, err := s.fetchCarParks(carParkURL, token)
	if err != nil {
		log.Printf("❌ [TDXService] fetchCarParks Error: %v", err)
		// 🎯 如果抓取失敗，但快取裡面有舊資料，優先退回使用舊快取避免崩潰
		if val, ok := s.cache.Load(city); ok {
			log.Printf("⚠️ [TDXService] API error occurred, falling back to cached data for %s", city)
			return val.(CacheItem).Data, nil
		}
		return nil, err
	}
	log.Printf("📦 [TDXService] Fetched CarParks count: %d", len(carParks))

	availURL := fmt.Sprintf("https://tdx.transportdata.tw/api/basic/v1/Parking/OffStreet/ParkingAvailability/City/%s?format=JSON", city)
	log.Printf("📌 [TDXService] Requesting AvailURL: %s", availURL)

	availMap, err := s.fetchAvailabilities(availURL, token)
	if err != nil {
		log.Printf("❌ [TDXService] fetchAvailabilities Error: %v", err)
		// 同樣在即時車位失敗時，若有快取則退回快取
		if val, ok := s.cache.Load(city); ok {
			log.Printf("⚠️ [TDXService] Avail API error, falling back to cached data for %s", city)
			return val.(CacheItem).Data, nil
		}
		return nil, err
	}
	log.Printf("📦 [TDXService] Fetched Availabilities count: %d", len(availMap))

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

	// 🎯 關鍵防禦：如果這次 TDX 突然回傳 0 筆資料，但我們手上有舊快取，就保留舊快取不覆蓋
	if len(unifiedList) == 0 {
		if val, ok := s.cache.Load(city); ok {
			log.Printf("⚠️ [TDXService] TDX returned 0 results, keeping previous valid cache for %s", city)
			return val.(CacheItem).Data, nil
		}
	}

	// 4. 成功取得正常資料，寫入快取（有效期限 10 分鐘）
	if len(unifiedList) > 0 {
		s.cache.Store(city, CacheItem{
			Data:      unifiedList,
			ExpiresAt: time.Now().Add(10 * time.Minute),
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