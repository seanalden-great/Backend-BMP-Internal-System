// controllers/inventory_controller.go
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type InventoryController struct {
	DB *gorm.DB
}

type AdjustmentInput struct {
	Date       string `json:"date" binding:"required"`
	ProductSKU string `json:"product_sku" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	Reason     string `json:"reason" binding:"required"`
}

// GET /api/inventory/adjustments
func (ic *InventoryController) GetAdjustments(c *gin.Context) {
	var adjustments []models.StockAdjustment
	if err := ic.DB.Preload("Product").Order("date desc, id desc").Find(&adjustments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data penyesuaian stok"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": adjustments})
}

// POST /api/inventory/adjust
func (ic *InventoryController) AdjustStock(c *gin.Context) {
	var input AdjustmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	parsedDate, _ := time.Parse("2006-01-02", input.Date)
	tx := ic.DB.Begin()

	var product models.Product
	if err := tx.Where("sku = ?", input.ProductSKU).First(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	// Validasi pengurangan stok
	if input.Type == "Pengurangan" && product.Stock < input.Quantity {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stok tidak mencukupi untuk dikurangi"})
		return
	}

	// Nilai penyesuaian dihitung berdasarkan HPP saat ini
	totalValue := float64(input.Quantity) * product.HPP

	adj := models.StockAdjustment{
		Date:       parsedDate,
		ProductSKU: input.ProductSKU,
		Type:       input.Type,
		Quantity:   input.Quantity,
		Reason:     input.Reason,
		TotalValue: totalValue,
	}

	if err := tx.Create(&adj).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencatat penyesuaian"})
		return
	}

	// Update Stok di Master Produk
	if input.Type == "Penambahan" {
		product.Stock += input.Quantity
	} else {
		product.Stock -= input.Quantity
	}

	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate stok master"})
		return
	}

	// Catat Jurnal Akuntansi (Persediaan vs HPP/Beban Penyusutan)
	var invAcc, cogsAcc models.Account
	tx.Where("code LIKE ?", "113%").First(&invAcc)  // Persediaan
	tx.Where("code LIKE ?", "511%").First(&cogsAcc) // HPP / Selisih Persediaan

	if invAcc.Code != "" && cogsAcc.Code != "" && totalValue > 0 {
		jCode := "ADJ-" + input.ProductSKU + "-" + time.Now().Format("150405")
		j1 := models.Journal{Date: parsedDate, Description: "Penyesuaian Stok: " + input.Reason + " (" + jCode + ")"}
		tx.Create(&j1)

		if input.Type == "Penambahan" {
			tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: invAcc.Code, Debit: totalValue, Credit: 0})
			tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: cogsAcc.Code, Debit: 0, Credit: totalValue})
		} else {
			tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: cogsAcc.Code, Debit: totalValue, Credit: 0})
			tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: invAcc.Code, Debit: 0, Credit: totalValue})
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Stok berhasil disesuaikan!"})
}