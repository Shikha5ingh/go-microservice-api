// auth.go
package auth

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	key    = "sherkaghar@xdfgsgg"
	MaxAge = 86400 * 30
	IsProd = false
)

// Initialize authentication
func NewAuth1() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, proceeding with environment variables")
	}
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleSecret := os.Getenv("GOOGLE_SECRET")
	googleCallback := os.Getenv("GOOGLE_CALLBACK_URL")

	goth.UseProviders(
		google.New(googleClientID, googleSecret, googleCallback, "email", "profile"),
	)

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd
	store.Options.Path = "/"
	gothic.Store = store
}

// BeginAuthHandler starts the authentication process
func BeginAuthHandler(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// AuthCallbackHandler handles the callback from the provider
func AuthCallbackHandler(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed", "details": err.Error()})
		return
	}
	// Store user in session or context as needed
	c.Set("user", user)
	c.JSON(http.StatusOK, gin.H{"message": "Authentication successful", "user": user})
}

// AuthMiddleware protects routes and ensures user is authenticated
func AuthMiddleware1() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			// Redirect to authentication if not authenticated
			c.Redirect(http.StatusFound, "/auth/google")
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}
