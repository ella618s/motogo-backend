package main

import (
	"log"

	"motogo-backend/handlers"
	"motogo-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 載入 .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, relying on system environment variables")
	}

	tdxService := services.NewTDXService()
	parkingHandler := handlers.NewParkingHandler(tdxService)

	r := gin.Default()

	// 註冊 API 路由
	api := r.Group("/api/v1")
	{
		api.GET("/parking/taipei", parkingHandler.GetParkingList)
	}

	log.Println("Server running on http://localhost:8080")
	r.Run(":8080")
}