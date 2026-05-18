package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/models"
)

type CouponHandler struct{}

func NewCouponHandler() *CouponHandler {
	return &CouponHandler{}
}

func (h *CouponHandler) Validate(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon code is required"})
		return
	}

	var coupon models.Coupon
	row := database.DB.QueryRow(
		"SELECT id, code, discount_percent, max_uses, used_count FROM coupons WHERE code = ?",
		req.Code,
	)
	err := row.Scan(&coupon.ID, &coupon.Code, &coupon.DiscountPercent, &coupon.MaxUses, &coupon.UsedCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid coupon code"})
		return
	}

	if coupon.UsedCount >= coupon.MaxUses {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon has reached maximum uses"})
		return
	}

	database.DB.Exec("UPDATE coupons SET used_count = used_count + 1 WHERE id = ?", coupon.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":             coupon.Code,
		"discount_percent": coupon.DiscountPercent,
		"message":          "Coupon applied successfully",
	})
}
