package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v74/github"

	"github-env-manager/internal/models"
	"github-env-manager/internal/services"
)

// GetAuthURL returns the GitHub OAuth URL
func GetAuthURL(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"type":         "pat",
		"instructions": "Please create a GitHub Personal Access Token with 'repo' and 'workflow' scopes",
		"url":          "https://github.com/settings/tokens/new",
	})
}

// ValidateToken validates a GitHub Personal Access Token
func ValidateToken(c *gin.Context) {
	var req models.AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client to validate token
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(req.Token)

	// Try to get the authenticated user to validate the token
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Generate a new session ID
	sessionID := fmt.Sprintf("session_%s", user.GetLogin())

	// Store the authenticated user
	services.StoreUser(sessionID, &models.User{
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		AvatarURL: user.GetAvatarURL(),
		Token:     req.Token,
	})

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"user": gin.H{
			"login":     user.GetLogin(),
			"name":      user.GetName(),
			"avatarUrl": user.GetAvatarURL(),
		},
	})
}

// AuthenticateWithPAT handles Personal Access Token authentication
func AuthenticateWithPAT(c *gin.Context) {
	var req models.AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	// Validate the token by making a request to GitHub API
	client := &http.Client{}
	req2, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req2.Header.Set("Authorization", "token "+req.Token)
	req2.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Parse user info
	var userData struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	// Store user info
	user := &models.User{
		Login:     userData.Login,
		Name:      userData.Name,
		AvatarURL: userData.AvatarURL,
		Token:     req.Token,
	}

	// Generate a simple session ID (in production, use proper session management)
	sessionID := fmt.Sprintf("session_%s", userData.Login)
	services.StoreUser(sessionID, user)

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"user": gin.H{
			"login":     user.Login,
			"name":      user.Name,
			"avatarUrl": user.AvatarURL,
		},
		"sessionId": sessionID,
	})
}

// HandleAuthCallback handles OAuth callback (not used for PAT authentication)
func HandleAuthCallback(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "PAT authentication doesn't use callbacks"})
}

// GetAuthStatus checks if the user is authenticated
func GetAuthStatus(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
		return
	}

	user, exists := services.GetUser(sessionID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"login":     user.Login,
		"name":      user.Name,
		"avatarUrl": user.AvatarURL,
	})
}
