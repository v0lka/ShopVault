package models

import "time"

type User struct {
	ID                int64     `json:"id"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"password_hash"`
	FullName          string    `json:"full_name"`
	Role              string    `json:"role"`
	ResetToken        string    `json:"reset_token,omitempty"`
	ResetTokenExpiry  time.Time `json:"reset_token_expiry,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
}

type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	Total           float64     `json:"total"`
	Status          string      `json:"status"`
	ShippingAddress string      `json:"shipping_address"`
	CCNumber        string      `json:"cc_number"`
	CCExpiry        string      `json:"cc_expiry"`
	CCCVV           string      `json:"cc_cvv"`
	CouponCode      string      `json:"coupon_code,omitempty"`
	DiscountPercent float64     `json:"discount_percent"`
	CreatedAt       time.Time   `json:"created_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Review struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	UserID    int64     `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type Coupon struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`
	DiscountPercent float64   `json:"discount_percent"`
	MaxUses         int       `json:"max_uses"`
	UsedCount       int       `json:"used_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CartItem struct {
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}
