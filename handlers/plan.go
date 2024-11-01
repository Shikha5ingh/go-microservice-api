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

		decoder := json.NewDecoder(bytes.NewReader(jsonData))
		decoder.DisallowUnknownFields()

		// Bind JSON to Plan struct using the stored body
		var plan models.Plan
		if err := decoder.Decode(&plan); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
			return
		}
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

		exists, err := redisClient.Exists(ctx, plan.ObjectID).Result()
		if err != nil {
			log.Println("Error checking if plan exists in Redis:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if plan exists"})
			return
		}
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Plan already exists"})
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

// Get All Plan handles GET /api/v1/plans
func GetAllPlans(redisClient *redis.Client) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		// Get all keys from redis
		planKeys, err := redisClient.Keys(ctx, "*").Result()
		if err != nil {
			log.Println("Error retrieving plan keys from Redis:", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plans"})
			return
		}
		if len(planKeys) == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "No plans found"})
			return
		}

		var plans []models.Plan
		for _, key := range planKeys {
			planData, err := redisClient.Get(ctx, key).Result()
			if err != nil {
				log.Println("Error retrieving plan from Redis:", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plan"})
				return
			}
			var plan models.Plan
			if err := json.Unmarshal([]byte(planData), &plan); err != nil {
				log.Println("Error unmarshalling plan data:", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal plan data"})
				return
			}
			plans = append(plans, plan)

		}
		plansData, err := json.Marshal(plans)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Serialization error", "details": err.Error()})
			return
		}
		// Generate ETag
		hasher := sha1.New()
		hasher.Write(plansData)
		etag := fmt.Sprintf(`"%x"`, hasher.Sum(nil))

		etag = strings.Trim(etag, "\"") // Removing quotes from the ETag

		log.Println("ETag:", etag)
		ifMatch := ctx.GetHeader("If-Match")
		if ifMatch != "" && ifMatch != etag {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": "Precondition failed"})
			return
		}

		ifNoneMatch := ctx.GetHeader("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			ctx.Status(http.StatusNotModified)
			return
		}

		ctx.Header("ETag", etag)
		ctx.JSON(http.StatusOK, plans)
	}
}

// GetPlan handles GET /api/v1/plans/:id
func GetPlanByID(redisClient *redis.Client) gin.HandlerFunc {
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
		etag = strings.Trim(etag, "\"") // Removing quotes from the ETag

		log.Println("ETag:", etag)
		ifMatch := c.GetHeader("If-Match")
		if ifMatch != "" && ifMatch != etag {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Precondition failed"})
			return
		}

		ifNoneMatch := c.GetHeader("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			c.Status(http.StatusNotModified)
			return
		}

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
		log.Println("Object ID:", objectId)
		planKeys, err := redisClient.Keys(ctx, objectId).Result()
		if err != nil {
			log.Println("Error retrieving plan keys from Redis:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plans"})
			return
		}
		if len(planKeys) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No plans found for the given ID"})
			return
		}
		error := redisClient.Del(ctx, objectId).Err()
		if error != nil {
			log.Println("Error deleting plan from Redis:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Plan deleted successfully"})
	}
}

// UpdatePlan handles PATCH /api/v1/plans/:id
func UpdatePlan(redisClient *redis.Client, validator *utils.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		objectId := c.Param("id")
		log.Println("Object ID:", objectId)

		// Read the request body and parse the JSON data into a map of partial updates
		var patchData map[string]interface{}
		if err := c.ShouldBindJSON(&patchData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
			return
		}
		if len(patchData) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
			return
		}
		planData, err := redisClient.Get(ctx, objectId).Result()
		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		if err != nil {
			log.Println("Error retrieving plan from Redis:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plan"})
			return
		}

		var exisingPlan map[string]interface{}
		if err := json.Unmarshal([]byte(planData), &exisingPlan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Deserialization error", "details": err.Error()})
			return
		}
		hasher := sha1.New()
		hasher.Write([]byte(planData))
		currentETag := fmt.Sprintf(`"%x"`, hasher.Sum(nil))
		currentETag = strings.Trim(currentETag, "\"") // Remove quotes
		println("Current ETag:", currentETag)

		ifMatch := c.GetHeader("If-Match")
		println("If-Match:", ifMatch)
		if ifMatch != "" && ifMatch != currentETag {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Precondition failed"})
			return
		}
		ifNoneMatch := c.GetHeader("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == currentETag {
			c.Status(http.StatusNotModified)
			return
		}

		for key, value := range patchData {
			exisingPlan[key] = value
		}
		exisingPlan["CreationData"] = time.Now().Format("01-02-2006")
		updatedPlanJSON, err := json.Marshal(exisingPlan)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Serialization error", "details": err.Error()})
			return
		}
		if err := validator.Validate(updatedPlanJSON); err != nil {
			log.Println("Validation failed:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}
		if err := redisClient.Set(ctx, objectId, updatedPlanJSON, 0).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error", "details": err.Error()})
			return
		}
		// Generate ETag
		hasher = sha1.New()
		hasher.Write(updatedPlanJSON)
		newETag := fmt.Sprintf(`"%x"`, hasher.Sum(nil))
		// newETag = strings.Trim(newETag, "\"") // Removing quotes from the ETag
		c.Header("ETag", newETag)
		c.JSON(http.StatusOK, exisingPlan)

	}
}
