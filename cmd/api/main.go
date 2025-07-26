package main

import (
	"log/slog"
	"os"

	"github.com/Utsav7428/ChronoXchange/internal/api"
	"github.com/Utsav7428/ChronoXchange/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to the database
	database.Connect()

	// Set up the web server
	router := gin.Default()

	// Group routes under /api/v1
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", api.Signup)
			auth.POST("/login", api.Login)
		}
		orders := v1.Group("/orders")
		orders.Use(api.AuthMiddleware())
		{
			orders.POST("", api.CreateOrder)
		}
	}

	slog.Info("API server starting on :8080")
	router.Run(":8080")
}
