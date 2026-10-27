// controllers/purchase_controller.go
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type PurchaseController struct {
	DB *gorm.DB
}

// Format JSON yang diharapkan dari Frontend
type PurchaseItemInput struct {
	ProductSKU string  `json:"product_sku" binding:"required"`
	Quantity   int     `json:"quantity" binding:"required,min=1"`
	Price      float64 `json:"price" binding:"required,min=0"`
}

type PurchaseInput struct {
	Date        string              `json:"date" binding:"required"`
	DueDate     string              `json:"due_date"`
	ReferenceNo string              `json:"reference_no"`
	ContactID   uint                `json:"contact_id" binding:"required"`
	Notes       string              `json:"notes"`
	Items       []PurchaseItemInput `json:"items" binding:"required,min=1"`
}

// POST /api/purchases
func (pc *PurchaseController) CreatePurchase(c *gin.Context) {
	var input PurchaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	var parsedDueDate time.Time
	if input.DueDate != "" {
		parsedDueDate, _ = time.Parse("2006-01-02", input.DueDate)
	} else {
		parsedDueDate = parsedDate // Default disamakan dengan tanggal nota
	}

	// Mulai Transaksi Database
	tx := pc.DB.Begin()

	// 1. Simpan Header (Nilai Total di-set 0 dulu)
	purchase := models.Purchase{
		Date:        parsedDate,
		DueDate:     parsedDueDate,
		ReferenceNo: input.ReferenceNo,
		ContactID:   input.ContactID,
		Status:      "Pesanan", // Default Fase Awal
		Notes:       input.Notes,
		Total:       0, 
	}

	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan header pesanan"})
		return
	}

	// 2. Simpan Detail Item sekaligus menghitung Grand Total
	var grandTotal float64
	for _, itemInput := range input.Items {
		subTotal := float64(itemInput.Quantity) * itemInput.Price
		grandTotal += subTotal

		purchaseItem := models.PurchaseItem{
			PurchaseID: purchase.ID,
			ProductSKU: itemInput.ProductSKU,
			Quantity:   itemInput.Quantity,
			Price:      itemInput.Price,
			SubTotal:   subTotal,
		}

		if err := tx.Create(&purchaseItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan detail pesanan: " + itemInput.ProductSKU})
			return
		}
	}

	// 3. Perbarui Nilai Total di Header
	if err := tx.Model(&purchase).Update("Total", grandTotal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate total pesanan"})
		return
	}

	// Selesaikan Transaksi
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pesanan Pembelian (PO) berhasil dibuat",
	})
}

// GET /api/purchases (Menyediakan filter ?status=Pesanan)
func (pc *PurchaseController) GetPurchases(c *gin.Context) {
	var purchases []models.Purchase
	
	// Tarik semua relasi secara otomatis agar JSON komplit
	query := pc.DB.Preload("Contact").
		Preload("Items").
		Preload("Items.Product").
		Order("date desc, id desc")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pembelian"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": purchases})
}

func (pc *PurchaseController) ReceivePurchase(c *gin.Context) {
	id := c.Param("id")
	var purchase models.Purchase

	// Tarik data Header beserta Detail Item-nya
	if err := pc.DB.Preload("Items").First(&purchase, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan"})
		return
	}

	// Validasi status
	if purchase.Status != "Pesanan" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pesanan ini sudah diproses atau dibatalkan"})
		return
	}

	// Mulai Transaksi
	tx := pc.DB.Begin()

	// 1. Ubah Status Header menjadi Pengiriman
	purchase.Status = "Pengiriman"
	if err := tx.Save(&purchase).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate status pesanan"})
		return
	}

	// 2. Loop setiap barang, lalu Update Stok & Kalkulasi HPP (Moving Average)
	for _, item := range purchase.Items {
		var product models.Product
		if err := tx.Where("sku = ?", item.ProductSKU).First(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Produk tidak ditemukan: " + item.ProductSKU})
			return
		}

		// Rumus Moving Average HPP:
		// HPP Baru = ((Stok Lama * HPP Lama) + (Qty Datang * Harga Beli Datang)) / (Stok Lama + Qty Datang)
		oldTotalValue := float64(product.Stock) * product.HPP
		newIncomingValue := float64(item.Quantity) * item.Price
		newTotalStock := product.Stock + item.Quantity

		var newHPP float64
		if newTotalStock > 0 {
			newHPP = (oldTotalValue + newIncomingValue) / float64(newTotalStock)
		}

		// Update Master Produk
		product.Stock = newTotalStock
		product.HPP = newHPP
		product.BuyPrice = item.Price // Update referensi harga beli terakhir

		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate stok produk: " + product.SKU})
			return
		}
	}

	// Commit Transaksi
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success", 
		"message": "Barang berhasil diterima, Stok & HPP telah diperbarui otomatis!",
	})
}

type BillInput struct {
	DueDate string `json:"due_date" binding:"required"`
}

// PUT /api/purchases/:id/bill
func (pc *PurchaseController) CreateBill(c *gin.Context) {
	id := c.Param("id")
	var input BillInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal jatuh tempo tidak valid"})
		return
	}

	parsedDueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal salah, gunakan YYYY-MM-DD"})
		return
	}

	var purchase models.Purchase
	if err := pc.DB.Preload("Items").First(&purchase, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengiriman pembelian tidak ditemukan"})
		return
	}

	if purchase.Status != "Pengiriman" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya dokumen dengan status 'Pengiriman' yang dapat dibuatkan tagihan"})
		return
	}

	tx := pc.DB.Begin()

	// 1. Update Status dan Tanggal Jatuh Tempo Header
	purchase.Status = "Tagihan"
	purchase.DueDate = parsedDueDate
	if err := tx.Save(&purchase).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate status tagihan"})
		return
	}

	// 2. Otomatis Buat Jurnal Akuntansi (Double-Entry)
	// Cari Akun Persediaan (Biasanya kode berawalan '113' atau sesuaikan CoA Anda) dan Utang Usaha ('211')
	// var inventoryAccount models.Account
	// var apAccount models.Account

	// tx.Where("code LIKE ?", "160%").First(&inventoryAccount) // Persediaan
	// if inventoryAccount.Code == "" {
	// 	tx.Where("code LIKE ?", "11%").First(&inventoryAccount) // Fallback ke aset lancar jika tidak ketemu
	// }
	// tx.Where("code LIKE ?", "310%").First(&apAccount)         // Utang Usaha

	// if inventoryAccount.Code == "" || apAccount.Code == "" {
	// 	tx.Rollback()
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Akun Persediaan atau Utang Usaha belum dikonfigurasi di CoA"})
	// 	return
	// }
	
	// 2. Otomatis Buat Jurnal Akuntansi (Double-Entry)
	var inventoryAccount models.Account
	var apAccount models.Account

	// PERBAIKAN: Cari berdasarkan awalan 16% DAN wajib is_header = false (Akun Detail)
	tx.Where("code LIKE ? AND is_header = ?", "16%", false).First(&inventoryAccount) // Persediaan
	if inventoryAccount.Code == "" {
		tx.Where("code LIKE ? AND is_header = ?", "11%", false).First(&inventoryAccount) // Fallback
	}
	
	// PERBAIKAN: Cari berdasarkan awalan 31% DAN wajib is_header = false (Akun Detail)
	tx.Where("code LIKE ? AND is_header = ?", "31%", false).First(&apAccount)         // Utang Usaha

	if inventoryAccount.Code == "" || apAccount.Code == "" {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Akun Persediaan atau Utang Usaha (Detail) belum dikonfigurasi di CoA"})
		return
	}

	// journalCode := "BILL-" + time.Now().Format("20060102") + "-" + id
	// journal := models.Journal{
	// 	Date:        time.Now(),
	// 	Description: "Tagihan Pembelian dari PO #" + id + " (Ref: " + purchase.ReferenceNo + ")",
	// 	Reference:   journalCode,
	// }
	
	journalCode := "BILL-" + time.Now().Format("20060102") + "-" + id
	journal := models.Journal{
		Date:        time.Now(),
		// Masukkan journalCode ke dalam Description
		Description: "Tagihan Pembelian dari PO #" + id + " (Ref: " + purchase.ReferenceNo + ") - " + journalCode, 
	}

	if err := tx.Create(&journal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat jurnal akuntansi"})
		return
	}

	entries := []models.JournalEntry{
		{JournalID: journal.ID, AccountCode: inventoryAccount.Code, Debit: purchase.Total, Credit: 0},
		{JournalID: journal.ID, AccountCode: apAccount.Code, Debit: 0, Credit: purchase.Total},
	}

	for _, entry := range entries {
		if err := tx.Create(&entry).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencatat rincian jurnal"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tagihan Pembelian berhasil diterbitkan & Jurnal Akuntansi otomatis tercatat!",
	})
}