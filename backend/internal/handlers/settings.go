package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// RuntimeConfig holds dynamically-configurable application settings.
var RuntimeConfig = make(map[string]interface{})
var configMu sync.RWMutex

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// Update handles POST /api/admin/settings
func (h *SettingsHandler) Update(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	configMu.Lock()
	for key, value := range req {
		RuntimeConfig[key] = value
	}
	configMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
		"keys":    len(req),
	})
}

// GetRuntimeConfig returns the current runtime config (for internal use)
func GetRuntimeConfig(key string) (interface{}, bool) {
	configMu.RLock()
	defer configMu.RUnlock()
	val, ok := RuntimeConfig[key]
	return val, ok
}
