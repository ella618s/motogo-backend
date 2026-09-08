package handlers

import (
	"net/http"

	"motogo-backend/services"

	"github.com/gin-gonic/gin"
)

type ParkingHandler struct {
	tdxService *services.TDXService
}

func NewParkingHandler(tdxService *services.TDXService) *ParkingHandler {
	return &ParkingHandler{tdxService: tdxService}
}

func (h *ParkingHandler) GetParkingList(c *gin.Context) {
	// 從網址路由中取得 :city 參數
	city := c.Param("city")

	token, err := h.tdxService.GetAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with TDX", "details": err.Error()})
		return
	}

	// 將取得的 city 傳入 GetParkingDataByCity
	data, err := h.tdxService.GetParkingDataByCity(city, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(data),
		"data":   data,
	})
}