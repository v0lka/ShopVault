package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/middleware"
	"shopvault/internal/models"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) GetOrders(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")

	var rows *sql.Rows
	var err error

	if userID != "" {
		query := fmt.Sprintf(
			"SELECT id, user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent, created_at FROM orders WHERE user_id = %s",
			userID,
		)
		rows, err = database.DB.Query(query)
	} else if status != "" {
		query := fmt.Sprintf(
			"SELECT id, user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent, created_at FROM orders WHERE status = '%s'",
			status,
		)
		rows, err = database.DB.Query(query)
	} else {
		rows, err = database.DB.Query("SELECT id, user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent, created_at FROM orders ORDER BY created_at DESC")
	}

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

		itemRows, _ := database.DB.Query("SELECT id, product_id, quantity, price FROM order_items WHERE order_id = ?", o.ID)
		if itemRows != nil {
			items := make([]models.OrderItem, 0)
			for itemRows.Next() {
				var item models.OrderItem
				if err := itemRows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
					continue
				}
				item.OrderID = o.ID
				items = append(items, item)
			}
			itemRows.Close()
			o.Items = items
		}

		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, email, password_hash, full_name, role, reset_token, created_at FROM users ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		var resetToken sql.NullString
		var resetExpiry sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &resetToken, &u.CreatedAt); err != nil {
			continue
		}
		if resetToken.Valid {
			u.ResetToken = resetToken.String
		}
		if resetExpiry.Valid {
			u.ResetTokenExpiry = resetExpiry.Time
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, description, price, image_url, category, stock FROM products ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock); err != nil {
			continue
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO products (name, description, price, image_url, category, stock) VALUES (?, ?, ?, ?, ?, ?)",
		p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.Stock,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	p.ID = id

	c.JSON(http.StatusCreated, gin.H{"product": p})
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	_, err = database.DB.Exec(
		"UPDATE products SET name = ?, description = ?, price = ?, image_url = ?, category = ?, stock = ? WHERE id = ?",
		p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.Stock, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *AdminHandler) WebhookCallback(c *gin.Context) {
	var req struct {
		OrderID     string `json:"order_id"`
		Status      string `json:"status"`
		CallbackURL string `json:"callback_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.OrderID != "" {
		database.DB.Exec("UPDATE orders SET status = ? WHERE id = ?", req.Status, req.OrderID)
	}

	if req.CallbackURL != "" {
		resp, err := http.Get(req.CallbackURL)
		if err != nil {
			log.Printf("Webhook callback failed for %s: %v", req.CallbackURL, err)
		} else {
			resp.Body.Close()
			log.Printf("Webhook callback sent to %s", req.CallbackURL)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed"})
}

func GetUserID(c *gin.Context) (int64, bool) {
	return middleware.GetUserFromContext(c)
}
