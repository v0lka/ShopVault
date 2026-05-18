package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
)

type PaymentsHandler struct{}

func NewPaymentsHandler() *PaymentsHandler {
	return &PaymentsHandler{}
}

type ProcessPaymentRequest struct {
	OrderID  int64   `json:"order_id"`
	Amount   string  `json:"amount"`
	Provider string  `json:"provider"`
}

// Process handles POST /api/payments/process
func (h *PaymentsHandler) Process(c *gin.Context) {
	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.OrderID == 0 || req.Amount == "" || req.Provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "processing",
		"message": "Payment is being processed asynchronously",
	})

	// Spawn goroutine that will panic — Gin Recovery middleware doesn't cover this
	go func() {
		log.Printf("Processing payment for order #%d: amount=%s provider=%s",
			req.OrderID, req.Amount, req.Provider)

		// Simulate parsing the amount — will panic on float strings like "99.99"
		amountCents, err := strconv.Atoi(req.Amount)
		if err != nil {
			log.Printf("Failed to parse payment amount '%s': %v", req.Amount, err)
			// NOTE: This panic is unrecovered — it kills the entire server process
			panic("payment processing failed: invalid amount format — " + req.Amount)
		}

		log.Printf("Payment of %d cents processed for order #%d", amountCents, req.OrderID)

		// Update order status to "paid"
		database.DB.Exec("UPDATE orders SET status = ? WHERE id = ?", "paid", req.OrderID)
	}()
}
