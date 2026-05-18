package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
)

type ImportHandler struct{}

func NewImportHandler() *ImportHandler {
	return &ImportHandler{}
}

func (h *ImportHandler) ImportFromURL(c *gin.Context) {
	var req struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	resp, err := http.Get(req.URL)
	if err != nil {
		log.Printf("Import failed for %s: %v", req.URL, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var products []map[string]interface{}
	if err := json.Unmarshal(body, &products); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format. Expected array of products."})
		return
	}

	imported := 0
	for _, p := range products {
		name, _ := p["name"].(string)
		description, _ := p["description"].(string)
		price, _ := p["price"].(float64)
		imageURL, _ := p["image_url"].(string)
		category, _ := p["category"].(string)

		stock := 0
		if s, ok := p["stock"].(float64); ok {
			stock = int(s)
		}

		if name != "" && price > 0 {
			_, err := database.DB.Exec(
				"INSERT INTO products (name, description, price, image_url, category, stock) VALUES (?, ?, ?, ?, ?, ?)",
				name, description, price, imageURL, category, stock,
			)
			if err != nil {
				log.Printf("Failed to import product '%s': %v", name, err)
			} else {
				imported++
			}
		}
	}

	log.Printf("Imported %d products from %s", imported, req.URL)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Import completed",
		"products_found": len(products),
		"products_imported": imported,
	})
}
