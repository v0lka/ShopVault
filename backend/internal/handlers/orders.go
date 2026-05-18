package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/middleware"
	"shopvault/internal/models"
)

type OrderHandler struct{}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

func (h *OrderHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent, created_at FROM orders WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.ShippingAddress, &o.CCNumber, &o.CCExpiry, &o.CCCVV, &o.CouponCode, &o.DiscountPercent, &o.CreatedAt); err != nil {
			continue
		}

		itemRows, err := database.DB.Query(
			"SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = ?",
			o.ID,
		)
		if err == nil {
			items := make([]models.OrderItem, 0)
			for itemRows.Next() {
				var item models.OrderItem
				if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
					continue
				}
				items = append(items, item)
			}
			itemRows.Close()
			o.Items = items
		}

		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var o models.Order
	row := database.DB.QueryRow(
		"SELECT id, user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent, created_at FROM orders WHERE id = ?",
		id,
	)
	err = row.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.ShippingAddress, &o.CCNumber, &o.CCExpiry, &o.CCCVV, &o.CouponCode, &o.DiscountPercent, &o.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	itemRows, err := database.DB.Query(
		"SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = ?",
		o.ID,
	)
	if err == nil {
		defer itemRows.Close()
		items := make([]models.OrderItem, 0)
		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
				continue
			}
			items = append(items, item)
		}
		o.Items = items
	}

	c.JSON(http.StatusOK, gin.H{"order": o})
}
