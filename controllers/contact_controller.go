// controllers/contact_controller.go
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type ContactController struct {
	DB *gorm.DB
}

// GET /api/contacts (Bisa pakai filter ?type=Customer atau ?type=Supplier)
func (cc *ContactController) GetContacts(c *gin.Context) {
	var contacts []models.Contact
	
	query := cc.DB.Order("name asc")
	
	// Tangkap query param jika ada filter tipe
	contactType := c.Query("type")
	if contactType != "" {
		query = query.Where("type = ?", contactType)
	}

	if err := query.Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kontak"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": contacts})
}

// POST /api/contacts
func (cc *ContactController) CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	if err := cc.DB.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kontak"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Kontak berhasil ditambahkan", "data": contact})
}

// PUT /api/contacts/:id
func (cc *ContactController) UpdateContact(c *gin.Context) {
	id := c.Param("id")
	var contact models.Contact
	
	if err := cc.DB.First(&contact, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kontak tidak ditemukan"})
		return
	}

	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	if err := cc.DB.Save(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate kontak"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Kontak berhasil diupdate"})
}

// DELETE /api/contacts/:id
func (cc *ContactController) DeleteContact(c *gin.Context) {
	id := c.Param("id")
	if err := cc.DB.Delete(&models.Contact{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kontak"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Kontak berhasil dihapus"})
}