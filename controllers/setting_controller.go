// controllers/setting_controller.go
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type SettingController struct {
	DB *gorm.DB
}

// GET /api/settings
func (sc *SettingController) GetSettings(c *gin.Context) {
	var setting models.CompanySetting
	// Selalu ambil record ID 1. Jika tidak ada, buatkan otomatis.
	if err := sc.DB.First(&setting, 1).Error; err != nil {
		setting = models.CompanySetting{
			ID: 1, Name: "PT Berlianmandiri Perkasa", TaxRate: 11.0,
		}
		sc.DB.Create(&setting)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": setting})
}

// PUT /api/settings
func (sc *SettingController) UpdateSettings(c *gin.Context) {
	var input models.CompanySetting
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	var setting models.CompanySetting
	if err := sc.DB.First(&setting, 1).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengaturan tidak ditemukan"})
		return
	}

	// Update spesifik field
	setting.Name = input.Name
	setting.Address = input.Address
	setting.Phone = input.Phone
	setting.Email = input.Email
	setting.TaxRate = input.TaxRate

	if err := sc.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengaturan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profil Perusahaan berhasil diperbarui!"})
}