// // models/setup.go
// package config

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"

// 	"accounting-be/models"
// )

// // Variabel DB global untuk digunakan oleh Controller
// var DB *gorm.DB

// func ConnectDatabase() {
// 	// Membaca DSN (Data Source Name) dari Environment Variable
// 	dsn := os.Getenv("DB_DSN")
// 	if dsn == "" {
// 		// Fallback default untuk lokal (Sesuaikan username, password, dan nama database Anda)
// 		dsn = "root:@tcp(127.0.0.1:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
// 	}

// 	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("Gagal koneksi ke database MySQL:", err)
// 	}

// 	// Optimasi koneksi (Sangat penting untuk arsitektur Serverless Vercel)
// 	sqlDB, err := database.DB()
// 	if err == nil {
// 		sqlDB.SetMaxIdleConns(2)
// 		sqlDB.SetMaxOpenConns(10)
// 		sqlDB.SetConnMaxLifetime(time.Hour)
// 	}

// 	// Auto Migrate: Membuat/memperbarui tabel berdasarkan struct GORM otomatis
// 	err = database.AutoMigrate(&models.Account{}, &models.Journal{}, &models.JournalEntry{}, &models.Product{}, &models.Category{}, &models.Contact{}, &models.Purchase{},
// 		&models.PurchaseItem{}, &models.Sale{}, &models.SaleItem{}, &models.Expense{}, &models.StockAdjustment{}, &models.CompanySetting{},&models.FixedAsset{},&models.FAQ{},)
// 	if err != nil {
// 		fmt.Println("Peringatan: Gagal menjalankan AutoMigrate:", err)
// 	}

// 	DB = database
// 	fmt.Println("Koneksi Database MySQL Berhasil!")
// }

// package config

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"

// 	"accounting-be/models"
// )

// var DB *gorm.DB

// func ConnectDatabase() {
// 	// 1. Baca dari Environment Variables Vercel
// 	host := os.Getenv("DB_HOST")
// 	user := os.Getenv("DB_USER")
// 	password := os.Getenv("DB_PASSWORD")
// 	dbname := os.Getenv("DB_NAME")
// 	port := os.Getenv("DB_PORT")

// 	// 2. Fallback ke Localhost JIKA variabel Vercel kosong (untuk testing di komputer Anda)
// 	if host == "" {
// 		host = "127.0.0.1"
// 		user = "root"          // Sesuaikan user lokal Anda
// 		password = ""          // Sesuaikan password lokal Anda
// 		dbname = "accounting_db" // Sesuaikan nama DB lokal Anda
// 		port = "3306"
// 	}

// 	// 3. Susun DSN (Data Source Name)
// 	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
// 	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=false",
// 		user, password, host, port, dbname)

// 	// 4. Buka Koneksi
// 	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("Gagal koneksi ke database MySQL: ", err)
// 	}

// 	// 5. Migrasi Tabel (Pastikan semua model Anda sudah terdaftar di sini)
// 	err = database.AutoMigrate(
// 		&models.Account{},
// 		&models.Journal{},
// 		&models.JournalEntry{},
// 		&models.Product{},
// 		&models.Category{},
// 		&models.Contact{},
// 		&models.Purchase{},
// 		&models.PurchaseItem{},
// 		&models.Sale{},
// 		&models.SaleItem{},
// 		&models.Expense{},
// 		&models.StockAdjustment{},
// 		&models.CompanySetting{},
// 		&models.FixedAsset{},
// 		&models.FAQ{},
// 	)
// 	if err != nil {
// 		log.Println("Error saat AutoMigrate:", err)
// 	}

// 	DB = database
// 	fmt.Println("Koneksi Database MySQL Berhasil!")
// }

package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"accounting-be/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	// 1. Baca Environment Variables (Dukung nama DB_HOST maupun bawaan Clever Cloud)
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = os.Getenv("MYSQL_ADDON_HOST")
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = os.Getenv("MYSQL_ADDON_USER")
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = os.Getenv("MYSQL_ADDON_PASSWORD")
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = os.Getenv("MYSQL_ADDON_DB")
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = os.Getenv("MYSQL_ADDON_PORT")
	}

	// 2. Fallback ke Localhost JIKA variabel Vercel masih kosong total
	if host == "" {
		host = "127.0.0.1"
		user = "root"
		password = ""
		dbname = "accounting_db"
		port = "3306"
		fmt.Println("WARNING: Menggunakan koneksi Localhost 127.0.0.1 karena Environment Variable tidak terbaca!")
	} else {
		fmt.Println("INFO: Menghubungkan ke database host:", host)
	}

	// 3. Susun DSN (Data Source Name)
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=false",
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=false",
		user, password, host, port, dbname)

	// 4. Buka Koneksi
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database MySQL: ", err)
	}

	// 5. Migrasi Tabel
	err = database.AutoMigrate(
		&models.Account{},
		&models.Journal{},
		&models.JournalEntry{},
		&models.Product{},
		&models.Category{},
		&models.Contact{},
		&models.Purchase{},
		&models.PurchaseItem{},
		&models.Sale{},
		&models.SaleItem{},
		&models.Expense{},
		&models.StockAdjustment{},
		&models.CompanySetting{},
		&models.FixedAsset{},
		&models.FAQ{},
	)
	if err != nil {
		log.Println("Error saat AutoMigrate:", err)
	}

	DB = database
	fmt.Println("Koneksi Database MySQL Berhasil!")
}