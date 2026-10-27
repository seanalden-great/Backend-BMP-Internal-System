// models/setup.go
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"accounting-be/models"
)

// Variabel DB global untuk digunakan oleh Controller
var DB *gorm.DB

func ConnectDatabase() {
	// Membaca DSN (Data Source Name) dari Environment Variable
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Fallback default untuk lokal (Sesuaikan username, password, dan nama database Anda)
		dsn = "root:@tcp(127.0.0.1:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database MySQL:", err)
	}

	// Optimasi koneksi (Sangat penting untuk arsitektur Serverless Vercel)
	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// Auto Migrate: Membuat/memperbarui tabel berdasarkan struct GORM otomatis
	err = database.AutoMigrate(&models.Account{}, &models.Journal{}, &models.JournalEntry{}, &models.Product{}, &models.Category{}, &models.Contact{}, &models.Purchase{},
		&models.PurchaseItem{}, &models.Sale{}, &models.SaleItem{}, &models.Expense{}, &models.StockAdjustment{}, &models.CompanySetting{},&models.FixedAsset{},&models.FAQ{},)
	if err != nil {
		fmt.Println("Peringatan: Gagal menjalankan AutoMigrate:", err)
	}

	DB = database
	fmt.Println("Koneksi Database MySQL Berhasil!")
}
