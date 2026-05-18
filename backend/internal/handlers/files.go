package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FilesHandler struct{}

func NewFilesHandler() *FilesHandler {
	return &FilesHandler{}
}

// View handles GET /api/files/view?path=<path>
func (h *FilesHandler) View(c *gin.Context) {
	userPath := c.Query("path")
	if userPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	// filepath.Join does NOT prevent traversal with absolute paths or ../ sequences
	// userPath "../../../etc/passwd" resolves outside ./uploads
	fullPath := filepath.Join("./uploads", userPath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}
