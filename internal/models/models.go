package models

import "time"

// Repository represents a GitHub repository
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Owner       Owner  `json:"owner"`
}

// Owner represents a GitHub repository owner
type Owner struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// Variable represents a GitHub repository or environment variable
type Variable struct {
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Secret represents a GitHub repository or environment secret
type Secret struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents an authenticated GitHub user
type User struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Token     string `json:"-"` // Don't expose token in JSON
}

// Environment represents a GitHub environment
type Environment struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// SyncRequest represents a request to sync variables between environments
type SyncRequest struct {
	SourceRepo    string   `json:"source_repo" binding:"required"`
	SourceEnv     string   `json:"source_env" binding:"required"`
	TargetRepos   []string `json:"target_repos" binding:"required"`
	TargetEnvs    []string `json:"target_envs" binding:"required"`
	VariableNames []string `json:"variable_names"`
	Overwrite     bool     `json:"overwrite"`
}

// ExportRequest represents a request to export variables
type ExportRequest struct {
	Repos []string `json:"repos" binding:"required"`
	Envs  []string `json:"envs" binding:"required"`
}

// ImportRequest represents a request to import variables
type ImportRequest struct {
	Repo      string            `json:"repo" binding:"required"`
	Variables map[string]string `json:"variables" binding:"required"`
	Overwrite bool              `json:"overwrite"`
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Token string `json:"token" binding:"required"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	TotalCount int  `json:"total_count"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
	TotalPages int  `json:"total_pages"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RepositoryListResponse represents a paginated repository list response
type RepositoryListResponse struct {
	Repositories []Repository `json:"repositories"`
	Pagination   Pagination   `json:"pagination"`
}
