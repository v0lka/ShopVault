package database

import (
	"crypto/md5"
	"fmt"
	"log"
	"time"
)

func Seed() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return
	}

	log.Println("Seeding database...")

	adminPassword := fmt.Sprintf("%x", md5.Sum([]byte("admin123")))

	_, err := DB.Exec(
		"INSERT INTO users (email, password_hash, full_name, role) VALUES (?, ?, ?, ?)",
		"admin@shopvault.com",
		adminPassword,
		"Admin User",
		"admin",
	)
	if err != nil {
		log.Printf("Failed to seed admin user: %v", err)
	}

	customerPassword := fmt.Sprintf("%x", md5.Sum([]byte("customer123")))
	_, err = DB.Exec(
		"INSERT INTO users (email, password_hash, full_name, role) VALUES (?, ?, ?, ?)",
		"customer@example.com",
		customerPassword,
		"Jane Customer",
		"user",
	)
	if err != nil {
		log.Printf("Failed to seed customer user: %v", err)
	}

	products := []struct {
		Name        string
		Description string
		Price       float64
		ImageURL    string
		Category    string
		Stock       int
	}{
		{"Wireless Headphones", "Premium Bluetooth headphones with noise cancellation and 30-hour battery life", 149.99, "/uploads/headphones.jpg", "Electronics", 50},
		{"Mechanical Keyboard", "RGB backlit mechanical keyboard with Cherry MX switches", 89.99, "/uploads/keyboard.jpg", "Electronics", 35},
		{"USB-C Hub", "7-in-1 USB-C hub with HDMI, SD card reader, and 100W power delivery", 45.99, "/uploads/usbhub.jpg", "Electronics", 100},
		{"Running Shoes", "Lightweight running shoes with responsive cushioning for daily training", 129.99, "/uploads/shoes.jpg", "Sports", 25},
		{"Yoga Mat", "Extra thick 6mm yoga mat with non-slip surface and carrying strap", 34.99, "/uploads/yogamat.jpg", "Sports", 60},
		{"Stainless Steel Water Bottle", "750ml insulated water bottle, keeps drinks cold for 24 hours", 24.99, "/uploads/bottle.jpg", "Sports", 80},
		{"Leather Wallet", "Genuine leather bifold wallet with RFID blocking technology", 49.99, "/uploads/wallet.jpg", "Fashion", 40},
		{"Canvas Backpack", "Vintage style canvas backpack with padded laptop compartment", 59.99, "/uploads/backpack.jpg", "Fashion", 30},
		{"Desk Lamp", "LED desk lamp with adjustable brightness, color temperature, and USB charging port", 39.99, "/uploads/desklamp.jpg", "Home", 45},
		{"Plant Pot Set", "Set of 3 ceramic plant pots with drainage holes and bamboo trays", 29.99, "/uploads/plantpots.jpg", "Home", 55},
		{"French Press Coffee Maker", "8-cup French press made of borosilicate glass and stainless steel", 27.99, "/uploads/frenchpress.jpg", "Home", 70},
		{"Notebook Set", "Pack of 3 hardcover dotted notebooks, 240 pages each, A5 size", 19.99, "/uploads/notebooks.jpg", "Office", 90},
	}

	for _, p := range products {
		_, err := DB.Exec(
			"INSERT INTO products (name, description, price, image_url, category, stock) VALUES (?, ?, ?, ?, ?, ?)",
			p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.Stock,
		)
		if err != nil {
			log.Printf("Failed to seed product '%s': %v", p.Name, err)
		}
	}

	_, err = DB.Exec(
		"INSERT INTO coupons (code, discount_percent, max_uses) VALUES (?, ?, ?)",
		"WELCOME10", 10.0, 100,
	)
	if err != nil {
		log.Printf("Failed to seed coupon: %v", err)
	}

	// Seed some review and order data for the demo customer
	var customerID int64
	DB.QueryRow("SELECT id FROM users WHERE email = ?", "customer@example.com").Scan(&customerID)
	if customerID > 0 {
		now := time.Now()
		orderDate := now.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")

		result, err := DB.Exec(
			"INSERT INTO orders (user_id, total, status, shipping_address, cc_number, cc_expiry, cc_cvv, coupon_code, discount_percent, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			customerID, 239.98, "delivered", "123 Main St, Springfield, IL 62701", "4111111111111111", "12/27", "123", "WELCOME10", 10.0, orderDate,
		)
		if err != nil {
			log.Printf("Failed to seed order: %v", err)
		} else {
			orderID, _ := result.LastInsertId()
			DB.Exec("INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)", orderID, 1, 1, 149.99)
			DB.Exec("INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)", orderID, 3, 2, 45.99)

			DB.Exec("INSERT INTO reviews (product_id, user_id, rating, comment) VALUES (?, ?, ?, ?)", 1, customerID, 5, "Great sound quality! The noise cancellation works perfectly on my commute.")
			DB.Exec("INSERT INTO reviews (product_id, user_id, rating, comment) VALUES (?, ?, ?, ?)", 3, customerID, 4, "Works well with my MacBook. Wish the cable was a bit longer though.")
		}
	}

	log.Println("Database seeded successfully")
}
