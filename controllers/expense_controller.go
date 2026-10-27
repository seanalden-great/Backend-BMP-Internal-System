// controllers/expense_controller.go
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type ExpenseController struct {
	DB *gorm.DB
}

type ExpenseInput struct {
	Date        string  `json:"date" binding:"required"`
	ReferenceNo string  `json:"reference_no"`
	ContactID   *uint   `json:"contact_id"`
	AccountCode string  `json:"account_code" binding:"required"`
	PaymentCode string  `json:"payment_code" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,min=1"`
	Description string  `json:"description" binding:"required"`
}

// GET /api/expenses
func (ec *ExpenseController) GetExpenses(c *gin.Context) {
	var expenses []models.Expense
	if err := ec.DB.Preload("Contact").Preload("Account").Preload("PaymentAcc").Order("date desc, id desc").Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data biaya"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": expenses})
}

// POST /api/expenses (Buat Biaya & Jurnal Otomatis)
func (ec *ExpenseController) CreateExpense(c *gin.Context) {
	var input ExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	parsedDate, _ := time.Parse("2006-01-02", input.Date)

	tx := ec.DB.Begin()

	// 1. Simpan Data Pengeluaran Biaya
	expense := models.Expense{
		Date:        parsedDate,
		ReferenceNo: input.ReferenceNo,
		ContactID:   input.ContactID,
		AccountCode: input.AccountCode,
		PaymentCode: input.PaymentCode,
		Amount:      input.Amount,
		Description: input.Description,
	}

	if err := tx.Create(&expense).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengeluaran"})
		return
	}

	// 2. Buat Jurnal Otomatis (Debit: Akun Beban, Kredit: Akun Kas/Bank)
	j1 := models.Journal{Date: parsedDate, Description: "Biaya: " + input.Description}
	if input.ReferenceNo != "" {
		j1.Description += " (Ref: " + input.ReferenceNo + ")"
	}
	tx.Create(&j1)

	// Debit: Beban
	tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: input.AccountCode, Debit: input.Amount, Credit: 0})
	// Kredit: Kas/Bank
	tx.Create(&models.JournalEntry{JournalID: j1.ID, AccountCode: input.PaymentCode, Debit: 0, Credit: input.Amount})

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Biaya berhasil dicatat beserta Jurnalnya!"})
}

// DELETE /api/expenses/:id
func (ec *ExpenseController) DeleteExpense(c *gin.Context) {
	id := c.Param("id")
	if err := ec.DB.Delete(&models.Expense{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus biaya"})
		return
	}
	// Catatan: Di sistem ERP sungguhan, Jurnal tidak boleh dihapus, melainkan di-reverse. 
	// Namun untuk penyederhanaan saat ini, kita biarkan jurnal lama tetap ada atau dihapus manual.
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Biaya berhasil dihapus"})
}