package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/models"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

func (h *ProductHandler) List(c *gin.Context) {
	category := c.Query("category")

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = database.DB.Query("SELECT id, name, description, price, image_url, category, stock FROM products WHERE category = ?", category)
	} else {
		rows, err = database.DB.Query("SELECT id, name, description, price, image_url, category, stock FROM products")
	}

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

func (h *ProductHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var p models.Product
	row := database.DB.QueryRow(
		"SELECT id, name, description, price, image_url, category, stock FROM products WHERE id = ?",
		id,
	)
	err = row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	revRows, err := database.DB.Query(
		"SELECT r.id, r.rating, r.comment, r.created_at, u.full_name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = ? ORDER BY r.created_at DESC",
		id,
	)
	if err == nil {
		defer revRows.Close()
		type ReviewWithUser struct {
			ID        int64  `json:"id"`
			Rating    int    `json:"rating"`
			Comment   string `json:"comment"`
			CreatedAt string `json:"created_at"`
			UserName  string `json:"user_name"`
		}
		reviews := make([]ReviewWithUser, 0)
		for revRows.Next() {
			var r ReviewWithUser
			if err := revRows.Scan(&r.ID, &r.Rating, &r.Comment, &r.CreatedAt, &r.UserName); err != nil {
				continue
			}
			reviews = append(reviews, r)
		}
		c.JSON(http.StatusOK, gin.H{"product": p, "reviews": reviews})
	} else {
		c.JSON(http.StatusOK, gin.H{"product": p, "reviews": []interface{}{}})
	}
}

func (h *ProductHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	q := fmt.Sprintf(
		"SELECT id, name, description, price, image_url, category, stock FROM products WHERE name LIKE '%%%s%%' OR description LIKE '%%%s%%'",
		query, query,
	)

	rows, err := database.DB.Query(q)
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

	c.JSON(http.StatusOK, gin.H{"products": products, "query": query})
}

func (h *ProductHandler) ImageProxy(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Header("Content-Type", contentType)
	io.Copy(c.Writer, resp.Body)
}
