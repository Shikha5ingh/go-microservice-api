package main

import (
	"go-rest-api/auth"
	"go-rest-api/handlers"
	"go-rest-api/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize authentication
	auth.NewAuth()
	// Initialize Redis client
	redisClient := utils.InitRedisClient()
	// Initialize JSON schema validator
	validator, err := utils.NewValidator("validation/schema.json")
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}
	// Initialize Gin router
	router := gin.Default()
	// Set up routes
	api := router.Group("/api/v1")
	api.Use(auth.AuthMiddleware())
	{
		plans := api.Group("/plans")
		{
			plans.POST("", handlers.CreatePlan(redisClient, validator))
			plans.GET("", handlers.GetAllPlans(redisClient))
			plans.GET("/:id", handlers.GetPlanByID(redisClient))
			plans.DELETE("/:id", handlers.DeletePlan(redisClient))
			plans.PATCH("/:id", handlers.UpdatePlan(redisClient, validator))
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
