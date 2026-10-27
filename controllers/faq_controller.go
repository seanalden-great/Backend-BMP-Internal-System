// controllers/faq_controller.go
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type FAQController struct {
	DB *gorm.DB
}

type FAQInput struct {
	Category string `json:"category" binding:"required"`
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
}

// GET /api/faqs
func (fc *FAQController) GetFAQs(c *gin.Context) {
	var faqs []models.FAQ
	if err := fc.DB.Order("category asc, id desc").Find(&faqs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data FAQ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": faqs})
}

// POST /api/faqs
func (fc *FAQController) CreateFAQ(c *gin.Context) {
	var input FAQInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	faq := models.FAQ{
		Category: input.Category,
		Question: input.Question,
		Answer:   input.Answer,
	}

	if err := fc.DB.Create(&faq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan FAQ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "FAQ berhasil ditambahkan!"})
}

// DELETE /api/faqs/:id
func (fc *FAQController) DeleteFAQ(c *gin.Context) {
	id := c.Param("id")
	if err := fc.DB.Delete(&models.FAQ{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus FAQ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "FAQ berhasil dihapus"})
}