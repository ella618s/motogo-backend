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
	token, err := h.tdxService.GetAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with TDX", "details": err.Error()})
		return
	}

	data, err := h.tdxService.GetTaipeiParkingData(token)
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