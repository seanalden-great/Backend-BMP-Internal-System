// controllers/sale_controller.go
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type SaleController struct {
	DB *gorm.DB
}

type SaleItemInput struct {
	ProductSKU string  `json:"product_sku" binding:"required"`
	Quantity   int     `json:"quantity" binding:"required,min=1"`
	Price      float64 `json:"price" binding:"required,min=0"`
}

type SaleInput struct {
	Date        string          `json:"date" binding:"required"`
	DueDate     string          `json:"due_date"`
	ReferenceNo string          `json:"reference_no"`
	ContactID   uint            `json:"contact_id" binding:"required"`
	Notes       string          `json:"notes"`
	Items       []SaleItemInput `json:"items" binding:"required,min=1"`
}

// GET /api/sales
func (sc *SaleController) GetSales(c *gin.Context) {
	var sales []models.Sale
	query := sc.DB.Preload("Contact").
		Preload("Items").
		Preload("Items.Product").
		Order("date desc, id desc")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data penjualan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": sales})
}

// POST /api/sales (Buat Pesanan Penjualan / SO)
func (sc *SaleController) CreateSale(c *gin.Context) {
	var input SaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	parsedDate, _ := time.Parse("2006-01-02", input.Date)
	parsedDueDate, _ := time.Parse("2006-01-02", input.DueDate)
	if input.DueDate == "" { parsedDueDate = parsedDate }

	tx := sc.DB.Begin()

	sale := models.Sale{
		Date:        parsedDate,
		DueDate:     parsedDueDate,
		ReferenceNo: input.ReferenceNo,
		ContactID:   input.ContactID,
		Status:      "Pesanan",
		Notes:       input.Notes,
		Total:       0,
	}

	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pesanan penjualan"})
		return
	}

	var grandTotal float64
	for _, itemInput := range input.Items {
		subTotal := float64(itemInput.Quantity) * itemInput.Price
		grandTotal += subTotal

		saleItem := models.SaleItem{
			SaleID:     sale.ID,
			ProductSKU: itemInput.ProductSKU,
			Quantity:   itemInput.Quantity,
			Price:      itemInput.Price,
			SubTotal:   subTotal,
		}
		if err := tx.Create(&saleItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan detail item penjualan"})
			return
		}
	}

	if err := tx.Model(&sale).Update("Total", grandTotal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui total penjualan"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pesanan Penjualan berhasil dibuat"})
}

// PUT /api/sales/:id/deliver (Pengiriman: Otomatis Kurangi Stok Gudang)
func (sc *SaleController) DeliverSale(c *gin.Context) {
	id := c.Param("id")
	var sale models.Sale
	if err := sc.DB.Preload("Items").First(&sale, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan"})
		return
	}

	if sale.Status != "Pesanan" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status pesanan tidak valid untuk dikirim"})
		return
	}

	tx := sc.DB.Begin()
	sale.Status = "Pengiriman"
	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate status pengiriman"})
		return
	}

	// Kurangi Stok Produk
	for _, item := range sale.Items {
		var product models.Product
		if err := tx.Where("sku = ?", item.ProductSKU).First(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Produk tidak ditemukan: " + item.ProductSKU})
			return
		}

		if product.Stock < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Stok tidak mencukupi untuk produk: " + product.Name})
			return
		}

		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengurangi stok produk"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Barang berhasil dikirim dan stok gudang dikurangi!"})
}

// PUT /api/sales/:id/invoice (Tagihan: Otomatis Buat Jurnal Piutang, Pendapatan, dan HPP)
func (sc *SaleController) CreateInvoice(c *gin.Context) {
	id := c.Param("id")
	var sale models.Sale
	if err := sc.DB.Preload("Items").First(&sale, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengiriman tidak ditemukan"})
		return
	}

	if sale.Status != "Pengiriman" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya dokumen dengan status 'Pengiriman' yang dapat ditagih"})
		return
	}

	tx := sc.DB.Begin()
	sale.Status = "Tagihan"
	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate status tagihan"})
		return
	}

	// Hitung Total HPP dari item yang terjual
	var totalHPP float64
	for _, item := range sale.Items {
		var prod models.Product
		tx.Where("sku = ?", item.ProductSKU).First(&prod)
		totalHPP += float64(item.Quantity) * prod.HPP
	}

	// Cari Akun Akuntansi (Piutang Usaha '112', Pendapatan '411', HPP '511', Persediaan '113')
	var arAcc, revAcc, cogsAcc, invAcc models.Account
	tx.Where("code LIKE ?", "112%").First(&arAcc)   // Piutang Usaha
	tx.Where("code LIKE ?", "411%").First(&revAcc)   // Pendapatan Penjualan
	tx.Where("code LIKE ?", "511%").First(&cogsAcc)  // Beban Pokok Penjualan (HPP)
	tx.Where("code LIKE ?", "113%").First(&invAcc)   // Persediaan

	if arAcc.Code == "" || revAcc.Code == "" {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konfigurasi Akun Piutang atau Pendapatan di CoA belum lengkap"})
		return
	}

	// // 1. Buat Jurnal Piutang & Pendapatan
	// j1 := models.Journal{Date: time.Now(), Description: "Tagihan Penjualan SO #" + id, Reference: "INV-S-" + id}
	// tx.Create(&j1)
	// tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: arAcc.Code, Debit: sale.Total, Credit: 0})
	// tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: revAcc.Code, Debit: 0, Credit: sale.Total})

	// // 2. Buat Jurnal HPP & Persediaan (Jika akun HPP & Persediaan tersedia)
	// if cogsAcc.Code != "" && invAcc.Code != "" && totalHPP > 0 {
	// 	j2 := models.Journal{Date: time.Now(), Description: "HPP Penjualan SO #" + id, Reference: "COGS-S-" + id}
	// 	tx.Create(&j2)
	// 	tx.Create(&models.JournalEntry{JournalID: j2.ID, AccountCode: cogsAcc.Code, Debit: totalHPP, Credit: 0})
	// 	tx.Create(&models.JournalEntry{JournalID: j2.ID, AccountCode: invAcc.Code, Debit: 0, Credit: totalHPP})
	// }
	
	// 1. Buat Jurnal Piutang & Pendapatan
	j1 := models.Journal{Date: time.Now(), Description: "Tagihan Penjualan SO #" + id + " (INV-S-" + id + ")"}
	tx.Create(&j1)
	tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: arAcc.Code, Debit: sale.Total, Credit: 0})
	tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: revAcc.Code, Debit: 0, Credit: sale.Total})

	// 2. Buat Jurnal HPP & Persediaan (Jika akun HPP & Persediaan tersedia)
	if cogsAcc.Code != "" && invAcc.Code != "" && totalHPP > 0 {
		j2 := models.Journal{Date: time.Now(), Description: "HPP Penjualan SO #" + id + " (COGS-S-" + id + ")"}
		tx.Create(&j2)
		tx.Create(&models.JournalEntry{JournalID: j2.ID, AccountCode: cogsAcc.Code, Debit: totalHPP, Credit: 0})
		tx.Create(&models.JournalEntry{JournalID: j2.ID, AccountCode: invAcc.Code, Debit: 0, Credit: totalHPP})
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Tagihan Penjualan & Jurnal Akuntansi berhasil diterbitkan!"})
}