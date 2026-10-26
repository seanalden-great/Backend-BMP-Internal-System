// controllers/journal_controller.go
package controllers

import (
	"time" // Tambahkan di bagian atas jika belum ada
	"net/http"

	"accounting-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Ini yang sebelumnya hilang
type JournalController struct {
	DB *gorm.DB
}

// Struktur format JSON yang diharapkan dari Frontend
type JournalInput struct {
	Date        string `json:"date" binding:"required"`
	Description string `json:"description" binding:"required"`
	ReferenceNo string `json:"reference_no"`
	Entries     []struct {
		AccountCode string  `json:"account_code" binding:"required"`
		Debit       float64 `json:"debit"`
		Credit      float64 `json:"credit"`
	} `json:"entries" binding:"required,min=2"` // Minimal harus ada 2 baris (Debit & Kredit)
}

// GET /api/journals/mutations?account_code=111
func (jc *JournalController) GetBankMutations(c *gin.Context) {
	accountCode := c.Query("account_code")
	if accountCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter account_code diperlukan"})
		return
	}

	// 1. Cari semua ID Jurnal yang melibatkan bank yang dipilih
	var journalIDs []uint
	jc.DB.Model(&models.JournalEntry{}).
		Where("account_code = ?", accountCode).
		Pluck("journal_id", &journalIDs)

	// Jika bank belum memiliki transaksi sama sekali
	if len(journalIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": []models.Journal{}})
		return
	}

	// 2. Tarik data Induk Jurnal, relasi Detail Baris (Entries), dan Nama Akunnya sekaligus
	var journals []models.Journal
	err := jc.DB.
		Preload("Entries").
		Preload("Entries.Account"). // Mengambil data relasi dari tabel accounts
		Where("id IN ?", journalIDs).
		Order("date DESC, id DESC"). // Urutkan dari transaksi terbaru
		Find(&journals).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data mutasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   journals,
	})
}

// POST /api/journals
func (jc *JournalController) CreateJournal(c *gin.Context) {
	var input JournalInput

	// 1. Validasi input JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	// 2. Validasi Akuntansi: Total Debit harus sama dengan Total Kredit
	var totalDebit, totalCredit float64
	for _, entry := range input.Entries {
		totalDebit += entry.Debit
		totalCredit += entry.Credit
	}
	if totalDebit != totalCredit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jurnal tidak seimbang (Total Debit != Total Kredit)"})
		return
	}

	// 3. Mulai Transaksi Database (GORM Transaction)
	// Jika salah satu gagal, seluruh proses akan di-rollback (dibatalkan)
	tx := jc.DB.Begin()

	parsedDate, _ := time.Parse("2006-01-02", input.Date)
	
	// Insert Header Jurnal
	journal := models.Journal{
		Date:        parsedDate,
		ReferenceNo: input.ReferenceNo,
		Description: input.Description,
	}

	if err := tx.Create(&journal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan header jurnal"})
		return
	}

	// Insert Detail Entries
	for _, e := range input.Entries {
		entry := models.JournalEntry{
			JournalID:   journal.ID,
			AccountCode: e.AccountCode,
			Debit:       e.Debit,
			Credit:      e.Credit,
		}
		if err := tx.Create(&entry).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan detail jurnal"})
			return
		}
	}

	// 4. Commit transaksi jika semuanya sukses
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Jurnal berhasil disimpan",
	})
}