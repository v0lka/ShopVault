package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/middleware"
)

type ReviewHandler struct{}

func NewReviewHandler() *ReviewHandler {
	return &ReviewHandler{}
}

func (h *ReviewHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		ProductID int64  `json:"product_id"`
		Rating    int    `json:"rating"`
		Comment   string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.ProductID == 0 || req.Rating < 1 || req.Rating > 5 || req.Comment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	_, err := database.DB.Exec(
		"INSERT INTO reviews (product_id, user_id, rating, comment) VALUES (?, ?, ?, ?)",
		req.ProductID, userID, req.Rating, req.Comment,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Review submitted successfully",
		"product_id": req.ProductID,
		"rating":     req.Rating,
		"comment":    req.Comment,
	})
}
