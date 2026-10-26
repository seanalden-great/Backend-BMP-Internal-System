// api/index.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"accounting-be/config"
	"accounting-be/controllers"
)

var app *gin.Engine

func init() {
	// Inisialisasi Database MySQL saat mesin Vercel menyala
	config.ConnectDatabase()

	gin.SetMode(gin.ReleaseMode)
	app = gin.Default()

	app.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Inisialisasi Controller dengan injeksi koneksi Database
	accountController := controllers.AccountController{DB: config.DB}
	journalController := controllers.JournalController{DB: config.DB}
	productController := controllers.ProductController{DB: config.DB}
	categoryController := controllers.CategoryController{DB: config.DB}
	contactController := controllers.ContactController{DB: config.DB}
	purchaseController := controllers.PurchaseController{DB: config.DB}
	saleController := controllers.SaleController{DB: config.DB}
	reportController := controllers.ReportController{DB: config.DB}
	
	// 1. Beritahu Gin untuk memuat semua file HTML di dalam folder templates
	app.LoadHTMLGlob("templates/*")

	// 2. Buat rute Root ("/") untuk menampilkan halaman Welcome
	app.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "welcome.tmpl", gin.H{
			"title": "Welcome to Go ERP Backend",
		})
	})

	api := app.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "Backend Go (Gin) dengan GORM MySQL berjalan di Vercel!",
			})
		})

		// Mengambil data langsung dari MySQL via Controller
		api.GET("/accounts", accountController.GetChartOfAccounts)
		api.POST("/accounts", accountController.CreateAccount)
		api.GET("/banks", accountController.GetBanks)
		api.GET("/journals/mutations", journalController.GetBankMutations)
		api.POST("/journals", journalController.CreateJournal)
		
		// Routes Produk
		api.GET("/products", productController.GetProducts)
		api.POST("/products", productController.CreateProduct)
		api.PUT("/products/:id", productController.UpdateProduct)
		api.DELETE("/products/:id", productController.DeleteProduct)
		
		api.GET("/categories", categoryController.GetCategories)
		api.POST("/categories", categoryController.CreateCategory)
		api.PUT("/categories/:id", categoryController.UpdateCategory)
		api.DELETE("/categories/:id", categoryController.DeleteCategory)
		
		// Routes Kontak (Customer & Supplier)
		api.GET("/contacts", contactController.GetContacts)
		api.POST("/contacts", contactController.CreateContact)
		api.PUT("/contacts/:id", contactController.UpdateContact)
		api.DELETE("/contacts/:id", contactController.DeleteContact)
		
		// Routes Pembelian
		api.GET("/purchases", purchaseController.GetPurchases)
		api.POST("/purchases", purchaseController.CreatePurchase)
		api.PUT("/purchases/:id/receive", purchaseController.ReceivePurchase)
		api.PUT("/purchases/:id/bill", purchaseController.CreateBill)
		
		api.GET("/sales", saleController.GetSales)
		api.POST("/sales", saleController.CreateSale)
		api.PUT("/sales/:id/deliver", saleController.DeliverSale)
		api.PUT("/sales/:id/invoice", saleController.CreateInvoice)

		api.GET("/reports", reportController.GetFinancialReport)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}