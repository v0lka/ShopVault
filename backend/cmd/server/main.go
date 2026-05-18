package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"shopvault/internal/database"
	"shopvault/internal/handlers"
	"shopvault/internal/middleware"
)

func main() {
	database.Init()
	database.Seed()

	os.MkdirAll("./uploads", 0755)

	authHandler := handlers.NewAuthHandler()
	productHandler := handlers.NewProductHandler()
	reviewHandler := handlers.NewReviewHandler()
	cartHandler := handlers.NewCartHandler()
	orderHandler := handlers.NewOrderHandler()
	couponHandler := handlers.NewCouponHandler()
	uploadHandler := handlers.NewUploadHandler()
	importHandler := handlers.NewImportHandler()
	adminHandler := handlers.NewAdminHandler()
	paymentsHandler := handlers.NewPaymentsHandler()
	filesHandler := handlers.NewFilesHandler()
	settingsHandler := handlers.NewSettingsHandler()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  fmt.Sprintf("Internal server error: %v", err),
					"panic":  err,
					"stack":  string(debug.Stack()),
					"env":    os.Environ(),
					"cwd":    func() string { d, _ := os.Getwd(); return d }(),
					"pid":    os.Getpid(),
				})
				c.Abort()
			}
		}()
		c.Next()
	})

	r.StaticFS("/uploads", http.Dir("./uploads"))
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authHandler.Me)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		api.PUT("/profile", middleware.AuthMiddleware(), authHandler.UpdateProfile)

		api.GET("/products", productHandler.List)
		api.GET("/products/search", productHandler.Search)
		api.GET("/products/image-proxy", productHandler.ImageProxy)
		api.GET("/products/:id", productHandler.Get)

		api.GET("/cart", cartHandler.GetCart)
		api.POST("/cart/checkout", middleware.AuthMiddleware(), cartHandler.Checkout)

		api.GET("/files/view", filesHandler.View)

		api.POST("/payments/process", paymentsHandler.Process)

		api.POST("/payments/capture", middleware.AuthMiddleware(), cartHandler.CapturePayment)

		api.GET("/orders", middleware.AuthMiddleware(), orderHandler.List)
		api.GET("/orders/:id", middleware.AuthMiddleware(), orderHandler.Get)

		api.POST("/coupons/validate", couponHandler.Validate)
		api.POST("/reviews", middleware.AuthMiddleware(), reviewHandler.Create)
		api.POST("/upload", middleware.AuthMiddleware(), middleware.AdminMiddleware(), uploadHandler.Upload)

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			admin.GET("/orders", adminHandler.GetOrders)
			admin.GET("/users", middleware.AdminMiddleware(), adminHandler.GetUsers)
			admin.GET("/products", middleware.AdminMiddleware(), adminHandler.ListProducts)
			admin.POST("/products", middleware.AdminMiddleware(), adminHandler.CreateProduct)
			admin.PUT("/products/:id", middleware.AdminMiddleware(), adminHandler.UpdateProduct)
			admin.DELETE("/products/:id", middleware.AdminMiddleware(), adminHandler.DeleteProduct)
			admin.POST("/import", middleware.AdminMiddleware(), importHandler.ImportFromURL)
			admin.POST("/settings", middleware.AdminMiddleware(), settingsHandler.Update)
		}

		api.POST("/webhook/payment-callback", adminHandler.WebhookCallback)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}
