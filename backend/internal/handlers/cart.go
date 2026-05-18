package handlers

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/middleware"
	"shopvault/internal/models"
)

type CartHandler struct{}

func NewCartHandler() *CartHandler {
	return &CartHandler{}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	cartCookie, err := c.Cookie("cart_data")
	if err != nil || cartCookie == "" {
		c.JSON(http.StatusOK, gin.H{"items": []models.CartItem{}, "total": 0.0})
		return
	}

	var items []models.CartItem
	decoder := gob.NewDecoder(bytes.NewBuffer([]byte(cartCookie)))
	if err := decoder.Decode(&items); err != nil {
		c.JSON(http.StatusOK, gin.H{"items": []models.CartItem{}, "total": 0.0})
		return
	}

	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

type CheckoutRequest struct {
	Items           []models.CartItem `json:"items"`
	ShippingAddress string            `json:"shipping_address"`
	CCNumber        string            `json:"cc_number"`
	CCExpiry        string            `json:"cc_expiry"`
	CCCVV           string            `json:"cc_cvv"`
	CouponCode      string            `json:"coupon_code"`
}

func (h *CartHandler) Checkout(c *gin.Context) {
	userID, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	total := 0.0
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)
	}

	discountPercent := 0.0
	if req.CouponCode != "" {
		var coupon models.Coupon
		row := database.DB.QueryRow(
			"SELECT id, code, discount_percent, max_uses, used_count FROM coupons WHERE code = ?",
			req.CouponCode,
		)
		if err := row.Scan(&coupon.ID, &coupon.Code, &coupon.DiscountPercent, &coupon.MaxUses, &coupon.UsedCount); err == nil {
			if coupon.UsedCount < coupon.MaxUses {
				database.DB.Exec("UPDATE coupons SET used_count = used_count + 1 WHERE id = ?", coupon.ID)
				discountPercent = coupon.DiscountPercent
			}
		}
	}

	if discountPercent > 0 {
		total = total * (1 - discountPercent/100)
	}

	result, err := database.DB.Exec(
		"INSERT INTO orders (user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		userID, total, "pending", req.ShippingAddress, req.CCNumber, req.CCExpiry, req.CCCVV, req.CouponCode, discountPercent,
	)
	if err != nil {
		log.Printf("Checkout error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	orderID, _ := result.LastInsertId()

	for _, item := range req.Items {
		database.DB.Exec(
			"INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)",
			orderID, item.ProductID, item.Quantity, item.Price,
		)
	}

	log.Printf("Order #%d created for user %d, total: %.2f", orderID, userID, total)

	c.JSON(http.StatusCreated, gin.H{
		"id":               orderID,
		"total":            total,
		"status":           "pending",
		"shipping_address": req.ShippingAddress,
		"discount_percent": discountPercent,
		"message":          "Order placed successfully",
	})
}

// UpdateCartCookie creates a gob-encoded cart cookie
func UpdateCartCookie(c *gin.Context, items []models.CartItem) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(items); err != nil {
		return
	}
	c.SetCookie("cart_data", string(buf.Bytes()), int(time.Hour*24*7), "/", "", false, false)
}
