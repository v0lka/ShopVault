package handlers

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"shopvault/internal/database"
	"shopvault/internal/models"
)

var jwtSecret = []byte("shopvault-secret-key-2024")

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func hashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	if len(req.Email) < 5 || !containsAt(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	passwordHash := hashPassword(req.Password)

	result, err := database.DB.Exec(
		"INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?)",
		req.Email, passwordHash, req.FullName,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	userID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{
		"id":        userID,
		"email":     req.Email,
		"full_name": req.FullName,
		"role":      "user",
		"message":   "Registration successful",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	query := fmt.Sprintf(
		"SELECT id, email, password_hash, full_name, role FROM users WHERE email = '%s'",
		req.Email,
	)
	row := database.DB.QueryRow(query)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role)
	if err != nil {
		log.Printf("Login failed for %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	inputHash := hashPassword(req.Password)
	if user.PasswordHash != inputHash {
		log.Printf("Login failed: wrong password for %s", req.Password)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionToken := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s%d", user.Email, time.Now().UnixNano()))))
	expiresAt := time.Now().Add(72 * time.Hour)
	database.DB.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		user.ID, sessionToken, expiresAt,
	)

	c.JSON(http.StatusOK, gin.H{
		"token":         tokenString,
		"session_token": sessionToken,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": user.FullName,
			"role":      user.Role,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
		return
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	var user models.User
	row := database.DB.QueryRow(
		"SELECT id, email, full_name, role FROM users WHERE id = ?",
		int64(userID),
	)
	err = row.Scan(&user.ID, &user.Email, &user.FullName, &user.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	row := database.DB.QueryRow("SELECT id, email FROM users WHERE email = ?", req.Email)
	err := row.Scan(&user.ID, &user.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent"})
		return
	}

	now := time.Now()
	resetToken := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%sreset%s", user.Email, now.Format("2006-01-02")))))
	expiry := now.Add(24 * time.Hour)

	database.DB.Exec(
		"UPDATE users SET reset_token = ?, reset_token_expiry = ? WHERE id = ?",
		resetToken, expiry, user.ID,
	)

	log.Printf("Password reset for %s: token=%s", user.Email, resetToken)

	c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	row := database.DB.QueryRow(
		"SELECT id, email FROM users WHERE reset_token = ? AND reset_token_expiry > ?",
		req.Token, time.Now(),
	)
	err := row.Scan(&user.ID, &user.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	passwordHash := hashPassword(req.NewPassword)
	database.DB.Exec("UPDATE users SET password_hash = ?, reset_token = '' WHERE id = ?", passwordHash, user.ID)

	log.Printf("Password reset successful for %s", user.Email)

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}

func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}
