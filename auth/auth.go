// auth.go
package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var googleClientID string

// Initialize authentication
func NewAuth() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, proceeding with environment variables")
	}
	googleClientID = os.Getenv("GOOGLE_CLIENT_KEY")
	if googleClientID == "" {
		fmt.Println("GOOGLE_CLIENT_ID is not set")
		os.Exit(1)
	}
}

// AuthMiddleware validates the bearer token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token missing"})
			c.Abort()
			return
		}

		token, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		// Token is valid
		c.Set("user", token.Claims)
		c.Next()
	}
}

func validateToken(idToken string) (*jwt.Token, error) {
	// Get Google's public keys
	jwksURL := "https://www.googleapis.com/oauth2/v3/certs"

	// Create the keyfunc options
	options := keyfunc.Options{
		// Refresh the JWKS when a token signed by an unknown KID is found
		RefreshUnknownKID: true,
	}

	// Get the JWKS from the remote server
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get the JWKS from the given URL: %v", err)
	}

	// Parse and validate the ID token
	token, err := jwt.Parse(idToken, jwks.Keyfunc)
	if err != nil {
		return nil, err
	}

	// Verify audience
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		aud, ok := claims["aud"].(string)
		if !ok || aud != googleClientID {
			return nil, errors.New("invalid audience")
		}
	} else {
		return nil, errors.New("invalid token claims")
	}

	return token, nil
}
