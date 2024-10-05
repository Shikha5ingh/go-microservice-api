package main

import (
	"log"
	"os"

	"go-rest-api/handlers"
	"go-rest-api/utils"

	"github.com/gin-gonic/gin"
)

func main() {

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize Redis client
	redisClient := utils.InitRedisClient()

	router := gin.Default()

	validator, err := utils.NewValidator("validation/schema.json")
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	// Set up routes
	api := router.Group("/api/v1")
	{
		plans := api.Group("/plans")
		{
			plans.POST("", handlers.CreatePlan(redisClient, validator))
			plans.GET("", handlers.GetAllPlans(redisClient))
			plans.GET("/:id", handlers.GetPlanByID(redisClient))
			plans.DELETE("/:id", handlers.DeletePlan(redisClient))
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
