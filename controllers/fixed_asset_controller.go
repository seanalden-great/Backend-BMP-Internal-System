// controllers/fixed_asset_controller.go
package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type FixedAssetController struct {
	DB *gorm.DB
}

type AssetInput struct {
	Name           string  `json:"name" binding:"required"`
	PurchaseDate   string  `json:"purchase_date" binding:"required"`
	PurchasePrice  float64 `json:"purchase_price" binding:"required,min=1"`
	UsefulLife     int     `json:"useful_life" binding:"required,min=1"`
	PaymentAccount string  `json:"payment_account" binding:"required"` // Kas/Bank (111)
	AssetAccount   string  `json:"asset_account" binding:"required"`   // Aset Tetap (121)
}

// GET /api/fixed-assets
func (fac *FixedAssetController) GetAssets(c *gin.Context) {
	var assets []models.FixedAsset
	if err := fac.DB.Order("purchase_date desc, id desc").Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data aset"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": assets})
}

// POST /api/fixed-assets (Beli & Daftar Aset)
func (fac *FixedAssetController) CreateAsset(c *gin.Context) {
	var input AssetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	parsedDate, _ := time.Parse("2006-01-02", input.PurchaseDate)
	tx := fac.DB.Begin()

	asset := models.FixedAsset{
		Name:          input.Name,
		PurchaseDate:  parsedDate,
		PurchasePrice: input.PurchasePrice,
		UsefulLife:    input.UsefulLife,
	}

	if err := tx.Create(&asset).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan aset"})
		return
	}

	// Jurnal Pembelian Aset Tetap (Debit: Aset Tetap, Kredit: Kas/Bank)
	j1 := models.Journal{Date: parsedDate, Description: "Pembelian Aset Tetap: " + input.Name}
	tx.Create(&j1)
	tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: input.AssetAccount, Debit: input.PurchasePrice, Credit: 0})
	tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: input.PaymentAccount, Debit: 0, Credit: input.PurchasePrice})

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Aset berhasil didaftarkan dan dibayar!"})
}

// PUT /api/fixed-assets/:id/depreciate (Susutkan Aset 1 Bulan)
func (fac *FixedAssetController) DepreciateAsset(c *gin.Context) {
	id := c.Param("id")
	tx := fac.DB.Begin()

	var asset models.FixedAsset
	if err := tx.First(&asset, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Aset tidak ditemukan"})
		return
	}

	if asset.Status != "Aktif" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aset sudah lunas disusutkan atau tidak aktif"})
		return
	}

	// Hitung penyusutan per bulan (Garis Lurus)
	deprAmount := asset.PurchasePrice / float64(asset.UsefulLife)
	
	// Cek apakah bulan ini akan melunasi aset
	if asset.AccumDepreciation+deprAmount >= asset.PurchasePrice {
		deprAmount = asset.PurchasePrice - asset.AccumDepreciation
		asset.Status = "Lunas Disusutkan"
	}

	asset.AccumDepreciation += deprAmount
	if err := tx.Save(&asset).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate nilai penyusutan"})
		return
	}

	// Jurnal Penyusutan (Asumsi: Akun 512 untuk Beban Penyusutan, Akun 122 untuk Akumulasi Penyusutan)
	// Catatan: Anda bisa menyesuaikan kode akun ini sesuai Master Data Anda
	var deprExpAcc, accumAcc models.Account
	tx.Where("code LIKE ?", "51%").First(&deprExpAcc) // Cari Beban
	tx.Where("code LIKE ?", "122%").First(&accumAcc)  // Cari Akumulasi
	
	if deprExpAcc.Code != "" && accumAcc.Code != "" {
		j2 := models.Journal{Date: time.Now(), Description: fmt.Sprintf("Penyusutan 1 Bln: %s", asset.Name)}
		tx.Create(&j2)
		tx.Create(&models.JournalEntry{JournalID: j2.ID, AccountCode: deprExpAcc.Code, Debit: deprAmount, Credit: 0})
		tx.Create(&models.JournalEntry{JournalID: j2.ID, AccountCode: accumAcc.Code, Debit: 0, Credit: deprAmount})
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Penyusutan 1 bulan berhasil diproses!"})
}