// // controllers/account_controller.go
// package controllers

// import (
// 	"net/http"
// 	"strings" // Ini yang sebelumnya hilang
// 	"accounting-be/models" // Sesuaikan dengan nama modul Anda
// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// type AccountController struct {
// 	DB *gorm.DB
// }

// // // GET /api/accounts - Untuk mensuplai data Accounts.tsx
// // func (ac *AccountController) GetChartOfAccounts(c *gin.Context) {
// // 	var accounts []models.Account
	
// // 	// Mengambil seluruh data akun diurutkan berdasarkan kode (agar hierarki rapi)
// // 	if err := ac.DB.Order("code asc").Find(&accounts).Error; err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data akun"})
// // 		return
// // 	}

// // 	// TODO: Logika kalkulasi Debit/Kredit dan Sisa Saldo dinamis 
// // 	// akan disisipkan di sini pada tahap selanjutnya melalui query SUM(journal_entries)

// // 	c.JSON(http.StatusOK, gin.H{
// // 		"status": "success",
// // 		"data":   accounts,
// // 	})
// // }

// func (ac *AccountController) GetChartOfAccounts(c *gin.Context) {
// 	var accounts []models.Account
	
// 	// Ambil data akun diurutkan berdasarkan kode
// 	if err := ac.DB.Order("code asc").Find(&accounts).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data akun"})
// 		return
// 	}

// 	// 1. Ambil total Debit dan Kredit dari seluruh Jurnal, dikelompokkan per Akun
// 	type AccountSum struct {
// 		AccountCode string
// 		TotalDebit  float64
// 		TotalCredit float64
// 	}
// 	var sums []AccountSum
// 	ac.DB.Model(&models.JournalEntry{}).
// 		Select("account_code, SUM(debit) as total_debit, SUM(credit) as total_credit").
// 		Group("account_code").
// 		Find(&sums)

// 	// 2. Buat Hash Map untuk pencarian data instan (O(1))
// 	sumMap := make(map[string]AccountSum)
// 	for _, s := range sums {
// 		sumMap[s.AccountCode] = s
// 	}

// 	// 3. Gabungkan data Master Akun dengan hasil kalkulasi Saldo
// 	var response []map[string]interface{}

// 	for _, acc := range accounts {
// 		var debit, kredit, sisa float64

// 		if acc.IsHeader {
// 			// ROLLUP: Jika ini Header, otomatis jumlahkan semua transaksi dari sub-akunnya
// 			// (Misal: Header "11" akan menarik total dari "111", "112", dst)
// 			for _, s := range sums {
// 				if strings.HasPrefix(s.AccountCode, acc.Code) {
// 					debit += s.TotalDebit
// 					kredit += s.TotalCredit
// 				}
// 			}
// 		} else {
// 			// Jika ini detail transaksi biasa, langsung ambil nominal dari Map
// 			if s, exists := sumMap[acc.Code]; exists {
// 				debit = s.TotalDebit
// 				kredit = s.TotalCredit
// 			}
// 		}

// 		// 4. Hitung SISA SALDO berdasarkan aturan Normal Balance Akuntansi
// 		if acc.NormalBalance == "D" {
// 			sisa = debit - kredit
// 		} else if acc.NormalBalance == "K" {
// 			sisa = kredit - debit
// 		} else {
// 			// Untuk Header yang normal balancenya "-", gunakan nilai absolut
// 			sisa = debit - kredit
// 			if sisa < 0 {
// 				sisa = -sisa
// 			}
// 		}

// 		// 5. Susun format JSON (Hanya kirim field angka jika bernilai > 0 agar UI rapi)
// 		item := map[string]interface{}{
// 			"code":     acc.Code,
// 			"name":     acc.Name,
// 			"isHeader": acc.IsHeader,
// 			"pos":      acc.NormalBalance,
// 			"level":    acc.Level,
// 		}
// 		if debit > 0 { item["debit"] = debit }
// 		if kredit > 0 { item["kredit"] = kredit }
// 		if sisa > 0 { item["sisa"] = sisa }

// 		response = append(response, item)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"data":   response,
// 	})
// }

// // GET /api/banks - Untuk mensuplai dropdown di TabelMutasiBank.tsx
// func (ac *AccountController) GetBanks(c *gin.Context) {
// 	var banks []models.Account
	
// 	// Filter pintar: Asumsi semua akun bank memiliki kode awalan "11" dan bukan header
// 	if err := ac.DB.Where("code LIKE ? AND is_header = ?", "11%", false).Find(&banks).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data bank"})
// 		return
// 	}

// 	// Format data agar persis sesuai ekspektasi array state daftarBank di React
// 	var responseData []map[string]interface{}
// 	for _, bank := range banks {
// 		responseData = append(responseData, map[string]interface{}{
// 			"id":   bank.Code,
// 			"nama": bank.Name,
// 			// Anda bisa menambahkan field 'account_number' khusus di tabel accounts jika diperlukan
// 			"rek":  "", 
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"data":   responseData,
// 	})
// }

// // controllers/account_controller.go
// package controllers

// import (
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"

// 	"accounting-be/models"
// )

// type AccountController struct {
// 	DB *gorm.DB
// }

// func (ac *AccountController) GetChartOfAccounts(c *gin.Context) {
// 	var accounts []models.Account
	
// 	if err := ac.DB.Order("code asc").Find(&accounts).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data akun"})
// 		return
// 	}

// 	type AccountSum struct {
// 		AccountCode string
// 		TotalDebit  float64
// 		TotalCredit float64
// 	}
// 	var sums []AccountSum
// 	ac.DB.Model(&models.JournalEntry{}).
// 		Select("account_code, SUM(debit) as total_debit, SUM(credit) as total_credit").
// 		Group("account_code").
// 		Find(&sums)

// 	sumMap := make(map[string]AccountSum)
// 	for _, s := range sums {
// 		sumMap[s.AccountCode] = s
// 	}

// 	var response []map[string]interface{}

// 	for _, acc := range accounts {
// 		var debit, kredit, sisa float64

// 		if acc.IsHeader {
// 			for _, s := range sums {
// 				if strings.HasPrefix(s.AccountCode, acc.Code) {
// 					debit += s.TotalDebit
// 					kredit += s.TotalCredit
// 				}
// 			}
// 		} else {
// 			if s, exists := sumMap[acc.Code]; exists {
// 				debit = s.TotalDebit
// 				kredit = s.TotalCredit
// 			}
// 		}

// 		// Perbaikan operator pengurangan yang sebelumnya hilang
// 		if acc.NormalBalance == "D" {
// 			sisa = debit - kredit
// 		} else if acc.NormalBalance == "K" {
// 			sisa = kredit - debit
// 		} else {
// 			sisa = debit - kredit
// 			if sisa < 0 {
// 				sisa = -sisa
// 			}
// 		}

// 		item := map[string]interface{}{
// 			"code":     acc.Code,
// 			"name":     acc.Name,
// 			"isHeader": acc.IsHeader,
// 			"pos":      acc.NormalBalance,
// 			"level":    acc.Level,
// 		}
// 		if debit > 0 { item["debit"] = debit }
// 		if kredit > 0 { item["kredit"] = kredit }
// 		if sisa > 0 { item["sisa"] = sisa }

// 		response = append(response, item)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"data":   response,
// 	})
// }

// func (ac *AccountController) GetBanks(c *gin.Context) {
// 	var banks []models.Account
	
// 	if err := ac.DB.Where("code LIKE ? AND is_header = ?", "11%", false).Find(&banks).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data bank"})
// 		return
// 	}

// 	var responseData []map[string]interface{}
// 	for _, bank := range banks {
// 		responseData = append(responseData, map[string]interface{}{
// 			"id":   bank.Code,
// 			"nama": bank.Name,
// 			"rek":  "", 
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"data":   responseData,
// 	})
// }

// // Struktur input JSON untuk Tambah Akun
// type AccountInput struct {
// 	Code          string `json:"code" binding:"required"`
// 	Name          string `json:"name" binding:"required"`
// 	IsHeader      bool   `json:"isHeader"`
// 	NormalBalance string `json:"pos" binding:"required"`
// 	Level         int    `json:"level"`
// }

// // POST /api/accounts - Logika Simpan Akun Baru
// func (ac *AccountController) CreateAccount(c *gin.Context) {
// 	var input AccountInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
// 		return
// 	}

// 	// Validasi apakah kode akun sudah ada
// 	var existing models.Account
// 	if err := ac.DB.Where("code = ?", input.Code).First(&existing).Error; err == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode akun sudah terdaftar!"})
// 		return
// 	}

// 	account := models.Account{
// 		Code:          input.Code,
// 		Name:          input.Name,
// 		IsHeader:      input.IsHeader,
// 		NormalBalance: input.NormalBalance,
// 		Level:         input.Level,
// 	}

// 	if err := ac.DB.Create(&account).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan akun baru"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":  "success",
// 		"message": "Akun berhasil ditambahkan",
// 	})
// }

// controllers/account_controller.go
package controllers

import (
	"net/http"

	// "strings"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type AccountController struct {
	DB *gorm.DB
}

func (ac *AccountController) GetChartOfAccounts(c *gin.Context) {
	var accounts []models.Account
	
	if err := ac.DB.Order("code asc").Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data akun"})
		return
	}

	type AccountSum struct {
		AccountCode string
		TotalDebit  float64
		TotalCredit float64
	}
	var sums []AccountSum
	ac.DB.Model(&models.JournalEntry{}).
		Select("account_code, SUM(debit) as total_debit, SUM(credit) as total_credit").
		Group("account_code").
		Find(&sums)

	sumMap := make(map[string]AccountSum)
	for _, s := range sums {
		sumMap[s.AccountCode] = s
	}

	var response []map[string]interface{}

	// for _, acc := range accounts {
	// 	var debit, kredit, sisa float64

	// 	if acc.IsHeader {
	// 		for _, s := range sums {
	// 			if strings.HasPrefix(s.AccountCode, acc.Code) {
	// 				debit += s.TotalDebit
	// 				kredit += s.TotalCredit
	// 			}
	// 		}
	// 	} else {
	// 		if s, exists := sumMap[acc.Code]; exists {
	// 			debit = s.TotalDebit
	// 			kredit = s.TotalCredit
	// 		}
	// 	}

	// 	if acc.NormalBalance == "D" {
	// 		sisa = debit - kredit
	// 	} else if acc.NormalBalance == "K" {
	// 		sisa = kredit - debit
	// 	} else {
	// 		sisa = debit - kredit
	// 		if sisa < 0 {
	// 			sisa = -sisa
	// 		}
	// 	}

	// 	item := map[string]interface{}{
	// 		"code":     acc.Code,
	// 		"name":     acc.Name,
	// 		"isHeader": acc.IsHeader,
	// 		"pos":      acc.NormalBalance,
	// 		"level":    acc.Level,
	// 	}
	// 	if debit > 0 { item["debit"] = debit }
	// 	if kredit > 0 { item["kredit"] = kredit }
	// 	if sisa > 0 { item["sisa"] = sisa }

	// 	response = append(response, item)
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"status": "success",
	// 	"data":   response,
	// })
	
	for i, acc := range accounts {
		var debit, kredit, sisa float64

		if acc.IsHeader {
			// LOGIKA BARU: Jumlahkan nilai dari akun anak berdasarkan LEVEL, bukan Prefix String
			for j := i + 1; j < len(accounts); j++ {
				child := accounts[j]
				
				// Jika bertemu dengan akun sejajar (level sama) atau induk lain (level lebih kecil), berhenti.
				if child.Level <= acc.Level {
					break 
				}
				
				// Hanya jumlahkan nominal jika anak tersebut adalah Akun Detail
				if !child.IsHeader {
					if s, exists := sumMap[child.Code]; exists {
						debit += s.TotalDebit
						kredit += s.TotalCredit
					}
				}
			}
		} else {
			// Jika ini akun detail, langsung ambil saldonya
			if s, exists := sumMap[acc.Code]; exists {
				debit = s.TotalDebit
				kredit = s.TotalCredit
			}
		}

		// Hitung Sisa Saldo sesuai Pos Normal Balance
		if acc.NormalBalance == "D" {
			sisa = debit - kredit
		} else if acc.NormalBalance == "K" {
			sisa = kredit - debit
		} else {
			sisa = debit - kredit
			if sisa < 0 {
				sisa = -sisa // Absolutkan jika netral
			}
		}

		item := map[string]interface{}{
			"code":     acc.Code,
			"name":     acc.Name,
			"isHeader": acc.IsHeader,
			"pos":      acc.NormalBalance,
			"level":    acc.Level,
		}
		if debit > 0 { item["debit"] = debit }
		if kredit > 0 { item["kredit"] = kredit }
		if sisa > 0 { item["sisa"] = sisa }

		response = append(response, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

func (ac *AccountController) GetBanks(c *gin.Context) {
	var banks []models.Account
	
	if err := ac.DB.Where("code LIKE ? AND is_header = ?", "11%", false).Find(&banks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data bank"})
		return
	}

	var responseData []map[string]interface{}
	for _, bank := range banks {
		responseData = append(responseData, map[string]interface{}{
			"id":   bank.Code,
			"nama": bank.Name,
			"rek":  "", 
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   responseData,
	})
}

type AccountInput struct {
	Code          string `json:"code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	IsHeader      bool   `json:"isHeader"`
	NormalBalance string `json:"pos" binding:"required"`
	Level         int    `json:"level"`
}

func (ac *AccountController) CreateAccount(c *gin.Context) {
	var input AccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	var existing models.Account
	if err := ac.DB.Where("code = ?", input.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode akun sudah terdaftar!"})
		return
	}

	account := models.Account{
		Code:          input.Code,
		Name:          input.Name,
		IsHeader:      input.IsHeader,
		NormalBalance: input.NormalBalance,
		Level:         input.Level,
	}

	if err := ac.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan akun baru"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Akun berhasil ditambahkan",
	})
}

// === FUNGSI BARU: UPDATE AKUN ===
func (ac *AccountController) UpdateAccount(c *gin.Context) {
	code := c.Param("code")
	var input AccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	var account models.Account
	if err := ac.DB.Where("code = ?", code).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun tidak ditemukan"})
		return
	}

	account.Name = input.Name
	account.IsHeader = input.IsHeader
	account.NormalBalance = input.NormalBalance
	account.Level = input.Level

	if err := ac.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui akun"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Akun berhasil diperbarui"})
}

// === FUNGSI BARU: DELETE AKUN ===
func (ac *AccountController) DeleteAccount(c *gin.Context) {
	code := c.Param("code")

	// Proteksi: Cek apakah akun sudah dipakai di Jurnal
	var count int64
	ac.DB.Model(&models.JournalEntry{}).Where("account_code = ?", code).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Akun tidak bisa dihapus karena sudah memiliki riwayat transaksi jurnal!"})
		return
	}

	if err := ac.DB.Where("code = ?", code).Delete(&models.Account{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus akun"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Akun berhasil dihapus"})
}