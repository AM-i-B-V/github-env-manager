package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v74/github"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
)

// GitHub API response structures
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Owner       Owner  `json:"owner"`
}

type Owner struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

type Variable struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Secret struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Production-ready GitHub Environment Manager - No mock data

// Store authenticated users in memory (in production, use a proper session store)
var authenticatedUsers = make(map[string]User)

type User struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Token     string `json:"-"` // Don't expose token in JSON
}

func getAuthURL(c *gin.Context) {
	// For PAT authentication, we don't need a URL - just instructions
	c.JSON(http.StatusOK, gin.H{
		"type":         "pat",
		"instructions": "Please create a GitHub Personal Access Token with 'repo' and 'workflow' scopes",
		"url":          "https://github.com/settings/tokens/new",
	})
}

func validateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}

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
	authenticatedUsers[sessionID] = User{
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		AvatarURL: user.GetAvatarURL(),
		Token:     req.Token,
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"user": gin.H{
			"login":     user.GetLogin(),
			"name":      user.GetName(),
			"avatarUrl": user.GetAvatarURL(),
		},
	})
}

func handleAuthCallback(c *gin.Context) {
	// This endpoint is not used for PAT authentication
	c.JSON(http.StatusBadRequest, gin.H{"error": "PAT authentication doesn't use callbacks"})
}

func authenticateWithPAT(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}

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
	user := User{
		Login:     userData.Login,
		Name:      userData.Name,
		AvatarURL: userData.AvatarURL,
		Token:     req.Token,
	}

	// Generate a simple session ID (in production, use proper session management)
	sessionID := fmt.Sprintf("session_%s", userData.Login)
	authenticatedUsers[sessionID] = user

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

func getAuthStatus(c *gin.Context) {
	// Check if user is authenticated via session
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
		return
	}

	user, exists := authenticatedUsers[sessionID]
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

func getRepositories(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25 // Limit to 25 repositories per page
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	query := c.Query("q")

	var repos []Repository

	// If there's a search query, use the fallback method which is more reliable
	if query != "" {
		fmt.Printf("Searching for query: %s\n", query)
		repos, totalCount := fallbackRepositorySearch(client, ctx, query, page, perPage)
		fmt.Printf("Found %d repositories matching query, total count: %d\n", len(repos), totalCount)

		totalPages := (totalCount + perPage - 1) / perPage
		hasNext := page < totalPages
		hasPrev := page > 1

		c.JSON(http.StatusOK, gin.H{
			"repositories": repos,
			"pagination": gin.H{
				"page":        page,
				"per_page":    perPage,
				"total_count": totalCount,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
				"total_pages": totalPages,
			},
		})
		return
	}

	// For regular browsing (no search query), fetch more repositories to ensure we get all accessible ones
	allRepos := fetchAllRepositories(client, ctx)

	// Apply pagination
	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(allRepos) {
		start = len(allRepos)
	}
	if end > len(allRepos) {
		end = len(allRepos)
	}

	// Get paginated results
	for i := start; i < end; i++ {
		repos = append(repos, allRepos[i])
	}

	// Calculate pagination info
	totalPages := (len(allRepos) + perPage - 1) / perPage
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"pagination": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total_count": len(allRepos),
			"has_next":    hasNext,
			"has_prev":    hasPrev,
			"total_pages": totalPages,
		},
	})
}

// Helper function to fetch all repositories a user has access to
func fetchAllRepositories(client *github.Client, ctx context.Context) []Repository {
	var allRepos []Repository

	// Fetch repositories with different types to ensure we get all accessible ones
	// Try different combinations to get all repositories the user has access to
	types := []string{"owner", "member", "all"}

	fmt.Printf("Fetching repositories for types: %v\n", types)

	for _, repoType := range types {
		page := 1
		for {
			opt := &github.RepositoryListOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: 100, // Maximum per page
				},
				Type: repoType,
			}

			repos, resp, err := client.Repositories.List(ctx, "", opt)
			if err != nil {
				fmt.Printf("Failed to fetch %s repositories: %v\n", repoType, err)
				break
			}

			// Convert repositories
			for _, repo := range repos {
				allRepos = append(allRepos, Repository{
					ID:          repo.GetID(),
					Name:        repo.GetName(),
					FullName:    repo.GetFullName(),
					Description: repo.GetDescription(),
					Private:     repo.GetPrivate(),
					Owner: Owner{
						Login: repo.Owner.GetLogin(),
						ID:    repo.Owner.GetID(),
					},
				})
			}

			// Check if there are more pages
			if resp.NextPage == 0 {
				break
			}
			page = resp.NextPage
		}
	}

	// Remove duplicates based on repository ID
	seen := make(map[int64]bool)
	var uniqueRepos []Repository
	for _, repo := range allRepos {
		if !seen[repo.ID] {
			seen[repo.ID] = true
			uniqueRepos = append(uniqueRepos, repo)
		}
	}

	fmt.Printf("Total repositories fetched: %d, after deduplication: %d\n", len(allRepos), len(uniqueRepos))
	return uniqueRepos
}

// Fallback search function when GitHub Search API fails
func fallbackRepositorySearch(client *github.Client, ctx context.Context, query string, page, perPage int) ([]Repository, int) {
	var repos []Repository

	// Fetch all repositories and filter them
	allRepos := fetchAllRepositories(client, ctx)

	// Filter repositories based on search query
	var filteredRepos []Repository
	queryLower := strings.ToLower(query)

	fmt.Printf("Filtering %d repositories with query: %s\n", len(allRepos), query)

	for _, repo := range allRepos {
		repoName := strings.ToLower(repo.Name)
		repoFullName := strings.ToLower(repo.FullName)
		repoDescription := strings.ToLower(repo.Description)
		ownerLogin := strings.ToLower(repo.Owner.Login)

		// Check if query matches any part of the repository
		if strings.Contains(repoName, queryLower) ||
			strings.Contains(repoFullName, queryLower) ||
			strings.Contains(repoDescription, queryLower) ||
			strings.Contains(ownerLogin, queryLower) {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	fmt.Printf("Found %d repositories matching the query\n", len(filteredRepos))

	// Apply pagination to filtered results
	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(filteredRepos) {
		start = len(filteredRepos)
	}
	if end > len(filteredRepos) {
		end = len(filteredRepos)
	}

	// Return paginated results
	for i := start; i < end; i++ {
		repos = append(repos, filteredRepos[i])
	}

	return repos, len(filteredRepos)
}

func getEnvironments(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get environments
	environments, _, err := client.Repositories.ListEnvironments(ctx, owner, repo, nil)
	if err != nil {
		// Check if it's a 404 error (repository doesn't exist or no access)
		if strings.Contains(err.Error(), "404") {
			c.JSON(http.StatusOK, []string{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environments from GitHub"})
		return
	}

	// Extract environment names
	envs := make([]string, 0, len(environments.Environments))
	for _, env := range environments.Environments {
		envs = append(envs, env.GetName())
	}

	c.JSON(http.StatusOK, envs)
}

func createEnvironment(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate environment name
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment name is required"})
		return
	}

	// Validate environment name format (GitHub requirements)
	if !isValidEnvironmentName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment name can only contain lowercase letters, numbers, and hyphens"})
		return
	}

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// GitHub environments are created automatically when first referenced
	// We'll create a simple environment by creating a deployment
	deploymentReq := &github.DeploymentRequest{
		Ref:         github.String("main"),
		Environment: &req.Name,
		Description: &req.Description,
	}

	_, _, err = client.Repositories.CreateDeployment(ctx, owner, repo, deploymentReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create environment: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"name":        req.Name,
		"description": req.Description,
		"created_at":  time.Now().Format("2006-01-02T15:04:05Z"),
	})
}

func isValidEnvironmentName(name string) bool {
	// GitHub environment names can only contain lowercase letters, numbers, and hyphens
	// and must be between 1 and 255 characters
	if len(name) < 1 || len(name) > 255 {
		return false
	}

	// Check if name contains only valid characters
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	// Name cannot start or end with a hyphen
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}

	return true
}

func getVariables(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get all Actions variables using pagination
	var allVariables []Variable
	page := 1
	for {
		opt := &github.ListOptions{
			Page:    page,
			PerPage: 100, // Maximum per page
		}

		variables, resp, err := client.Actions.ListRepoVariables(ctx, owner, repo, opt)
		if err != nil {
			// Check if it's a 404 error
			if strings.Contains(err.Error(), "404") {
				c.JSON(http.StatusOK, []Variable{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch variables from GitHub"})
			return
		}

		// Convert to our Variable struct
		for _, variable := range variables.Variables {
			allVariables = append(allVariables, Variable{
				Name:      variable.Name,
				Value:     variable.Value,
				CreatedAt: variable.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: variable.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	c.JSON(http.StatusOK, allVariables)
}

func getSecrets(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get all Actions secrets using pagination
	var allSecrets []Secret
	page := 1
	for {
		opt := &github.ListOptions{
			Page:    page,
			PerPage: 100, // Maximum per page
		}

		secrets, resp, err := client.Actions.ListRepoSecrets(ctx, owner, repo, opt)
		if err != nil {
			// Check if it's a 404 error
			if strings.Contains(err.Error(), "404") {
				c.JSON(http.StatusOK, []Secret{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch secrets from GitHub"})
			return
		}

		// Convert to our Secret struct
		for _, secret := range secrets.Secrets {
			allSecrets = append(allSecrets, Secret{
				Name:      secret.Name,
				CreatedAt: secret.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: secret.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	c.JSON(http.StatusOK, allSecrets)
}

func getEnvironmentVariables(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")

	// Get all environment variables using pagination
	var allVariables []Variable
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/environments/%s/variables?page=%d&per_page=100", owner, repo, env, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		req.Header.Set("Authorization", "token "+user.Token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environment variables from GitHub"})
			return
		}

		if resp.StatusCode == 404 {
			c.JSON(http.StatusOK, []Variable{})
			resp.Body.Close()
			return
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environment variables from GitHub"})
			return
		}

		var response struct {
			Variables []struct {
				Name      string    `json:"name"`
				Value     string    `json:"value"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			} `json:"variables"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}
		resp.Body.Close()

		// Convert to our Variable struct
		for _, variable := range response.Variables {
			allVariables = append(allVariables, Variable{
				Name:      variable.Name,
				Value:     variable.Value,
				CreatedAt: variable.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: variable.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		// Check if there are more pages (if we got less than 100 items, we're done)
		if len(response.Variables) < 100 {
			break
		}
		page++
	}

	c.JSON(http.StatusOK, allVariables)
}

func getEnvironmentSecrets(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")

	// Get all environment secrets using pagination
	var allSecrets []Secret
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/environments/%s/secrets?page=%d&per_page=100", owner, repo, env, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		req.Header.Set("Authorization", "token "+user.Token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environment secrets from GitHub"})
			return
		}

		if resp.StatusCode == 404 {
			c.JSON(http.StatusOK, []Secret{})
			resp.Body.Close()
			return
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environment secrets from GitHub"})
			return
		}

		var response struct {
			Secrets []struct {
				Name      string    `json:"name"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			} `json:"secrets"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}
		resp.Body.Close()

		// Convert to our Secret struct
		for _, secret := range response.Secrets {
			allSecrets = append(allSecrets, Secret{
				Name:      secret.Name,
				CreatedAt: secret.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: secret.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		// Check if there are more pages (if we got less than 100 items, we're done)
		if len(response.Secrets) < 100 {
			break
		}
		page++
	}

	c.JSON(http.StatusOK, allSecrets)
}

func createEnvironmentVariable(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")

	var req struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create environment variable using direct HTTP call
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/environments/%s/variables", owner, repo, env)

	payload := map[string]interface{}{
		"name":  req.Name,
		"value": req.Value,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	request.Header.Set("Authorization", "token "+user.Token)
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create environment variable"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		c.JSON(response.StatusCode, gin.H{"error": fmt.Sprintf("Failed to create environment variable: %s", string(body))})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Environment variable created successfully"})
}

func updateEnvironmentVariable(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")
	name := c.Param("name")

	var req struct {
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Delete the existing environment variable
	_, err = client.Actions.DeleteEnvVariable(ctx, owner, repo, env, name)
	if err != nil {
		fmt.Printf("Failed to delete environment variable: %v\n", err)
		// Continue anyway, the variable might not exist
	}

	// Create the new environment variable
	variable := &github.ActionsVariable{
		Name:  name,
		Value: req.Value,
	}

	_, err = client.Actions.CreateEnvVariable(ctx, owner, repo, env, variable)
	if err != nil {
		fmt.Printf("Failed to create environment variable: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update environment variable: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment variable updated successfully"})
}

func deleteEnvironmentVariable(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")
	name := c.Param("name")

	// Delete environment variable using direct HTTP call
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/environments/%s/variables/%s", owner, repo, env, name)

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	request.Header.Set("Authorization", "token "+user.Token)
	request.Header.Set("Accept", "application/vnd.github.v3+json")

	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete environment variable"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(response.Body)
		c.JSON(response.StatusCode, gin.H{"error": fmt.Sprintf("Failed to delete environment variable: %s", string(body))})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment variable deleted successfully"})
}

func createEnvironmentSecret(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")

	var req struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client using go-github library
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get the repository to find the repo ID (required for environment secrets)
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		fmt.Printf("Failed to get repository: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository: %v", err)})
		return
	}

	// Get the public key for the environment
	publicKey, _, err := client.Actions.GetEnvPublicKey(ctx, int(repository.GetID()), env)
	if err != nil {
		fmt.Printf("Failed to get environment public key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get environment public key: %v", err)})
		return
	}

	// Encrypt the secret value
	encryptedValue, err := encryptSecret(publicKey.GetKey(), req.Value)
	if err != nil {
		fmt.Printf("Failed to encrypt secret: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to encrypt secret: %v", err)})
		return
	}

	// Create environment secret using go-github library
	secret := &github.EncryptedSecret{
		Name:           req.Name,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedValue,
	}

	_, err = client.Actions.CreateOrUpdateEnvSecret(ctx, int(repository.GetID()), env, secret)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("GitHub API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create environment secret: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Environment secret created successfully"})
}

func updateEnvironmentSecret(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")
	name := c.Param("name")

	var req struct {
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client using go-github library
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get the repository to find the repo ID (required for environment secrets)
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		fmt.Printf("Failed to get repository: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository: %v", err)})
		return
	}

	// Get the public key for the environment
	publicKey, _, err := client.Actions.GetEnvPublicKey(ctx, int(repository.GetID()), env)
	if err != nil {
		fmt.Printf("Failed to get environment public key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get environment public key: %v", err)})
		return
	}

	// Encrypt the secret value
	encryptedValue, err := encryptSecret(publicKey.GetKey(), req.Value)
	if err != nil {
		fmt.Printf("Failed to encrypt secret: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to encrypt secret: %v", err)})
		return
	}

	// Update environment secret using go-github library
	secret := &github.EncryptedSecret{
		Name:           name,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedValue,
	}

	_, err = client.Actions.CreateOrUpdateEnvSecret(ctx, int(repository.GetID()), env, secret)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("GitHub API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update environment secret: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment secret updated successfully"})
}

func deleteEnvironmentSecret(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	env := c.Param("env")
	name := c.Param("name")

	// Create GitHub client using go-github library
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get the repository to find the repo ID (required for environment secrets)
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		fmt.Printf("Failed to get repository: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository: %v", err)})
		return
	}

	// Delete environment secret using go-github library
	_, err = client.Actions.DeleteEnvSecret(ctx, int(repository.GetID()), env, name)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("GitHub API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete environment secret: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment secret deleted successfully"})
}

func createVariable(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	var req struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Create Actions variable
	variable := &github.ActionsVariable{
		Name:  req.Name,
		Value: req.Value,
	}

	_, err = client.Actions.CreateRepoVariable(ctx, owner, repo, variable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create variable in GitHub"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Variable created successfully"})
}

func updateVariable(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	name := c.Param("name")

	var req struct {
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Update Actions variable
	variable := &github.ActionsVariable{
		Name:  name,
		Value: req.Value,
	}

	_, err = client.Actions.UpdateRepoVariable(ctx, owner, repo, variable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update variable in GitHub"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Variable updated successfully"})
}

func deleteVariable(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	name := c.Param("name")

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Delete Actions variable
	_, err = client.Actions.DeleteRepoVariable(ctx, owner, repo, name)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Variable not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete variable in GitHub"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Variable deleted successfully"})
}

func createSecret(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	var req struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client using go-github library
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get the public key for the repository
	publicKey, _, err := client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		fmt.Printf("Failed to get repository public key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository public key: %v", err)})
		return
	}

	// Encrypt the secret value
	encryptedValue, err := encryptSecret(publicKey.GetKey(), req.Value)
	if err != nil {
		fmt.Printf("Failed to encrypt secret: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to encrypt secret: %v", err)})
		return
	}

	// Create repository secret using go-github library
	secret := &github.EncryptedSecret{
		Name:           req.Name,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedValue,
	}

	_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, secret)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("GitHub API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create repository secret: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Repository secret created successfully"})
}

func updateSecret(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	name := c.Param("name")

	var req struct {
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create GitHub client using go-github library
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Get the public key for the repository
	publicKey, _, err := client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		fmt.Printf("Failed to get repository public key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get repository public key: %v", err)})
		return
	}

	// Encrypt the secret value
	encryptedValue, err := encryptSecret(publicKey.GetKey(), req.Value)
	if err != nil {
		fmt.Printf("Failed to encrypt secret: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to encrypt secret: %v", err)})
		return
	}

	// Update repository secret using go-github library
	secret := &github.EncryptedSecret{
		Name:           name,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedValue,
	}

	_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, secret)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("GitHub API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update repository secret: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository secret updated successfully"})
}

func deleteSecret(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	name := c.Param("name")

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(user.Token)

	// Delete Actions secret
	_, err = client.Actions.DeleteRepoSecret(ctx, owner, repo, name)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret in GitHub"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
}

func syncVariables(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		SourceRepo    string   `json:"source_repo"`
		SourceEnv     string   `json:"source_env"`
		TargetRepos   []string `json:"target_repos"`
		TargetEnvs    []string `json:"target_envs"`
		VariableNames []string `json:"variable_names"`
		Overwrite     bool     `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse source repo (format: "owner/repo")
	parts := strings.Split(req.SourceRepo, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source repo format. Use 'owner/repo'"})
		return
	}
	sourceOwner, sourceRepo := parts[0], parts[1]

	// Get source variables - Use Actions variables endpoint
	client := &http.Client{}
	sourceURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/variables", sourceOwner, sourceRepo)

	sourceReq, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	sourceReq.Header.Set("Authorization", "token "+user.Token)
	sourceReq.Header.Set("Accept", "application/vnd.github.v3+json")

	sourceResp, err := client.Do(sourceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch source variables"})
		return
	}
	defer sourceResp.Body.Close()

	if sourceResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch source variables from GitHub"})
		return
	}

	var sourceVariablesResponse struct {
		Variables []Variable `json:"variables"`
	}
	if err := json.NewDecoder(sourceResp.Body).Decode(&sourceVariablesResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse source variables"})
		return
	}
	sourceVariables := sourceVariablesResponse.Variables

	syncedCount := 0
	errors := []string{}

	// Sync to each target
	for _, targetRepo := range req.TargetRepos {
		targetParts := strings.Split(targetRepo, "/")
		if len(targetParts) != 2 {
			errors = append(errors, fmt.Sprintf("Invalid target repo format: %s", targetRepo))
			continue
		}
		targetOwner, targetRepoName := targetParts[0], targetParts[1]

		for _, targetEnv := range req.TargetEnvs {
			for _, variable := range sourceVariables {
				// Check if variable should be synced
				if len(req.VariableNames) == 0 || contains(req.VariableNames, variable.Name) {
					// Create/update variable in target - Use Actions variables endpoint
					targetURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/variables", targetOwner, targetRepoName)

					payload := map[string]string{
						"name":  variable.Name,
						"value": variable.Value,
					}

					jsonPayload, _ := json.Marshal(payload)
					targetReq, err := http.NewRequest("PUT", targetURL, strings.NewReader(string(jsonPayload)))
					if err != nil {
						errors = append(errors, fmt.Sprintf("Failed to create request for %s/%s/%s", targetRepo, targetEnv, variable.Name))
						continue
					}

					targetReq.Header.Set("Authorization", "token "+user.Token)
					targetReq.Header.Set("Accept", "application/vnd.github.v3+json")
					targetReq.Header.Set("Content-Type", "application/json")

					targetResp, err := client.Do(targetReq)
					if err != nil {
						errors = append(errors, fmt.Sprintf("Failed to sync %s to %s/%s/%s", variable.Name, targetRepo, targetEnv, variable.Name))
						continue
					}
					targetResp.Body.Close()

					if targetResp.StatusCode == http.StatusCreated || targetResp.StatusCode == http.StatusNoContent {
						syncedCount++
					} else {
						errors = append(errors, fmt.Sprintf("Failed to sync %s to %s/%s (status: %d)", variable.Name, targetRepo, targetEnv, targetResp.StatusCode))
					}
				}
			}
		}
	}

	response := gin.H{
		"message":      fmt.Sprintf("Successfully synced %d variables", syncedCount),
		"synced_count": syncedCount,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, response)
}

func exportVariables(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		Repos []string `json:"repos"`
		Envs  []string `json:"envs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	exportData := make(map[string]interface{})
	client := &http.Client{}

	for _, repo := range req.Repos {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			continue
		}
		owner, repoName := parts[0], parts[1]

		repoData := make(map[string]string)

		// Get variables for this repository (Actions variables are repo-wide)
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/variables", owner, repoName)

		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		request.Header.Set("Authorization", "token "+user.Token)
		request.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(request)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		var variablesResponse struct {
			Variables []Variable `json:"variables"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&variablesResponse); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Add variables to export data
		for _, variable := range variablesResponse.Variables {
			repoData[variable.Name] = variable.Value
		}

		exportData[repo] = repoData
	}

	c.JSON(http.StatusOK, exportData)
}

func importVariables(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		Repo      string            `json:"repo"`
		Variables map[string]string `json:"variables"`
		Overwrite bool              `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	parts := strings.Split(req.Repo, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repo format. Use 'owner/repo'"})
		return
	}
	owner, repoName := parts[0], parts[1]

	// Get existing environments for this repo
	client := &http.Client{}
	envsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/environments", owner, repoName)

	envsReq, err := http.NewRequest("GET", envsURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	envsReq.Header.Set("Authorization", "token "+user.Token)
	envsReq.Header.Set("Accept", "application/vnd.github.v3+json")

	envsResp, err := client.Do(envsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environments"})
		return
	}
	defer envsResp.Body.Close()

	if envsResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch environments from GitHub"})
		return
	}

	var envsResponse struct {
		Environments []struct {
			Name string `json:"name"`
		} `json:"environments"`
	}

	if err := json.NewDecoder(envsResp.Body).Decode(&envsResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse environments response"})
		return
	}

	// Import variables to the repository (Actions variables are repo-wide, not environment-specific)
	importedCount := 0
	errors := []string{}

	for name, value := range req.Variables {
		// Create/update variable
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/variables", owner, repoName)

		payload := map[string]string{
			"name":  name,
			"value": value,
		}

		jsonPayload, _ := json.Marshal(payload)
		request, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to create request for %s", name))
			continue
		}

		request.Header.Set("Authorization", "token "+user.Token)
		request.Header.Set("Accept", "application/vnd.github.v3+json")
		request.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(request)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to import %s", name))
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
			importedCount++
		} else {
			errors = append(errors, fmt.Sprintf("Failed to import %s (status: %d)", name, resp.StatusCode))
		}
	}

	response := gin.H{
		"message":        fmt.Sprintf("Successfully imported %d variables", importedCount),
		"imported_count": importedCount,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, response)
}

func compareEnvironments(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	repos := c.QueryArray("repos")
	if len(repos) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 2 repositories required for comparison"})
		return
	}

	comparison := make(map[string]map[string]string)
	client := &http.Client{}

	for _, repo := range repos {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			continue
		}
		owner, repoName := parts[0], parts[1]

		repoVars := make(map[string]string)

		// Get environments for this repo
		envsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/environments", owner, repoName)

		envsReq, err := http.NewRequest("GET", envsURL, nil)
		if err != nil {
			continue
		}

		envsReq.Header.Set("Authorization", "token "+user.Token)
		envsReq.Header.Set("Accept", "application/vnd.github.v3+json")

		envsResp, err := client.Do(envsReq)
		if err != nil || envsResp.StatusCode != http.StatusOK {
			if envsResp != nil {
				envsResp.Body.Close()
			}
			continue
		}

		var envsResponse struct {
			Environments []struct {
				Name string `json:"name"`
			} `json:"environments"`
		}

		if err := json.NewDecoder(envsResp.Body).Decode(&envsResponse); err != nil {
			envsResp.Body.Close()
			continue
		}
		envsResp.Body.Close()

		// Get repository variables (Actions variables are repo-wide)
		varsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/variables", owner, repoName)

		varsReq, err := http.NewRequest("GET", varsURL, nil)
		if err != nil {
			continue
		}

		varsReq.Header.Set("Authorization", "token "+user.Token)
		varsReq.Header.Set("Accept", "application/vnd.github.v3+json")

		varsResp, err := client.Do(varsReq)
		if err != nil || varsResp.StatusCode != http.StatusOK {
			if varsResp != nil {
				varsResp.Body.Close()
			}
			continue
		}

		var variablesResponse struct {
			Variables []Variable `json:"variables"`
		}
		if err := json.NewDecoder(varsResp.Body).Decode(&variablesResponse); err != nil {
			varsResp.Body.Close()
			continue
		}
		varsResp.Body.Close()

		for _, variable := range variablesResponse.Variables {
			repoVars[variable.Name] = variable.Value
		}

		comparison[repo] = repoVars
	}

	c.JSON(http.StatusOK, comparison)
}

func handleWebSocket(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "WebSocket endpoint"})
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getAuthenticatedUser(c *gin.Context) (*User, error) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		return nil, fmt.Errorf("no session")
	}

	user, exists := authenticatedUsers[sessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	return &user, nil
}

// encryptSecret encrypts a secret value using GitHub's public key for sealed box encryption
func encryptSecret(publicKey string, secretValue string) (string, error) {
	// Decode the base64-encoded public key
	decodedKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %v", err)
	}

	var recipientKey [32]byte
	copy(recipientKey[:], decodedKey)

	// Generate an ephemeral key pair
	ephemeralPub, ephemeralPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Create a nonce using blake2b
	hash, _ := blake2b.New(24, nil)
	hash.Write(ephemeralPub[:])
	hash.Write(recipientKey[:])
	var nonce [24]byte
	copy(nonce[:], hash.Sum(nil))

	// Encrypt the secret
	encrypted := box.Seal(nil, []byte(secretValue), &nonce, &recipientKey, ephemeralPriv)

	// Prepend the ephemeral public key to the encrypted message
	encryptedMessage := append(ephemeralPub[:], encrypted...)

	// Encode the result to base64
	return base64.StdEncoding.EncodeToString(encryptedMessage), nil
}
