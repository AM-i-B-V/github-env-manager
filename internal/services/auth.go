package services

import (
	"sync"
	"time"

	"github-env-manager/internal/models"
)

// Store authenticated users in memory (in production, use a proper session store)
var (
	authenticatedUsers = make(map[string]*models.User)
	userMutex          sync.RWMutex
)

// StoreUser stores a user in the session store
func StoreUser(sessionID string, user *models.User) {
	userMutex.Lock()
	defer userMutex.Unlock()
	authenticatedUsers[sessionID] = user
}

// GetUser retrieves a user from the session store
func GetUser(sessionID string) (*models.User, bool) {
	userMutex.RLock()
	defer userMutex.RUnlock()
	user, exists := authenticatedUsers[sessionID]
	return user, exists
}

// RemoveUser removes a user from the session store
func RemoveUser(sessionID string) {
	userMutex.Lock()
	defer userMutex.Unlock()
	delete(authenticatedUsers, sessionID)
}

// CleanupExpiredSessions removes expired sessions
// This is a placeholder for proper session management
func CleanupExpiredSessions() {
	// In a real implementation, you would check session expiration
	// For now, we'll just keep all sessions in memory
	time.Sleep(1 * time.Hour) // Placeholder
}
