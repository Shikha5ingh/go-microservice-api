package handlers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go-rest-api/models"
	"go-rest-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// CreatePlan handles POST /api/v1/plans
func CreatePlan(redisClient *redis.Client, validator *utils.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var jsonData []byte
		if c.Request.Body != nil {
			var err error
			jsonData, err = io.ReadAll(c.Request.Body)
			if err != nil {
				log.Println("Error reading request body:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
				return
			}
			if len(jsonData) == 0 {
				log.Println("JSON Data is empty")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
				return
			}
			log.Println("JSON Data:", string(jsonData))
			// Reset the body for further use
			c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonData))
		} else {
			log.Println("Empty request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty request body"})
			return
		}

		// Bind JSON to Plan struct using the stored body
		var plan models.Plan
		if err := c.ShouldBindJSON(&plan); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
			return
		}

		// Validate JSON data against the schema
		if err := validator.Validate(jsonData); err != nil {
			log.Println("Validation failed:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		// Set the CreationDate to the current date and time
		plan.CreationDate = time.Now().Format("01-02-2006")
		// Generate a unique numeric ID for the plan
		// id, err := redisClient.Incr(ctx, "plan_id_counter").Result()
		// if err != nil {
		// 	log.Println("Error generating plan ID:", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate plan ID"})
		// 	return
		// }
		// plan.ID = strconv.FormatInt(id, 10)

		// Serialize to JSON
		data, err := json.Marshal(plan)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Serialization error", "details": err.Error()})
			return
		}

		// Generate ETag
		hasher := sha1.New()
		hasher.Write(data)
		etag := fmt.Sprintf(`"%x"`, hasher.Sum(nil))

		// Set ETag header
		c.Header("ETag", etag)

		// Store in Redis
		if err := redisClient.Set(ctx, plan.ObjectID, data, 0).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error", "details": err.Error()})
			return
		}

		// Return response
		c.JSON(http.StatusCreated, plan)
	}
}

// GetPlan handles GET /api/v1/plans/:id
func GetPlan(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		objectId := c.Param("id")
		log.Println("Object ID:", objectId)

		planData, err := redisClient.Get(ctx, objectId).Result()
		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		} else if err != nil {
			log.Println("Error retrieving plan from Redis:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plan"})
			return
		}

		// Generate ETag
		hasher := sha1.New()
		hasher.Write([]byte(planData))
		etag := fmt.Sprintf(`"%x"`, hasher.Sum(nil))
		log.Println("ETag:", etag)
		ifNoneMatch := c.GetHeader("If-None-Match")
		log.Println("If-None-Match:", ifNoneMatch)
		log.Println(c.Request.Header)
		// Check If-None-Match header
		if match := c.GetHeader("If-None-Match"); match != "" {
			matchArray := c.Request.Header["If-None-Match"]
			match := matchArray[0]
			etag = strings.Trim(etag, "\"") // Removing quotes from the ETag
			log.Println("Match:", match)
			log.Println("ETag:", etag)
			log.Println(match == etag)
			if match == etag {
				c.Status(http.StatusNotModified)
				return
			}
		}

		// Set ETag header
		c.Header("ETag", etag)

		// Return data
		var plan models.Plan
		if err := json.Unmarshal([]byte(planData), &plan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Deserialization error", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, plan)
	}
}

// DeletePlan handles DELETE /api/v1/plans/:id
func DeletePlan(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		objectId := c.Param("id")

		if err := redisClient.Del(ctx, objectId).Err(); err != nil {
			if err == redis.Nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
				return
			}
			log.Println("Error deleting plan from Redis:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}