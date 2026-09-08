package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"motogo-backend/models"
	"motogo-backend/services"
)

type RouteHandler struct {
	routeService *services.RouteService
}

func NewRouteHandler(routeService *services.RouteService) *RouteHandler {
	return &RouteHandler{routeService: routeService}
}

func (h *RouteHandler) GetScooterRoute(c *gin.Context) {
	var req models.ScooterRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	points, err := h.routeService.GetScooterRoute(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ScooterRouteResponse{
		Status: "success",
		Points: points,
	})
}