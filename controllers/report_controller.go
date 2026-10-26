// controllers/report_controller.go
package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"accounting-be/models"
)

type ReportController struct {
	DB *gorm.DB
}

type AccountBalance struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
}

// GET /api/reports
func (rc *ReportController) GetFinancialReport(c *gin.Context) {
	var accounts []models.Account
	if err := rc.DB.Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data akun"})
		return
	}

	// Agregasi total debit dan kredit per akun dari tabel journal_entries
	type JournalSummary struct {
		AccountCode string
		TotalDebit  float64
		TotalCredit float64
	}
	var summaries []JournalSummary
	rc.DB.Table("journal_entries").
		Select("account_code, COALESCE(SUM(debit), 0) as total_debit, COALESCE(SUM(credit), 0) as total_credit").
		Group("account_code").
		Scan(&summaries)

	summaryMap := make(map[string]JournalSummary)
	for _, s := range summaries {
		summaryMap[s.AccountCode] = s
	}

	var reportData []AccountBalance
	var netProfit float64

	for _, acc := range accounts {
		sum := summaryMap[acc.Code]
		var balance float64
		accType := ""

		// Logika Saldo Normal berdasarkan Digit Pertama Kode Akun
		if strings.HasPrefix(acc.Code, "1") {
			accType = "Harta"
			balance = sum.TotalDebit - sum.TotalCredit
		} else if strings.HasPrefix(acc.Code, "2") {
			accType = "Kewajiban"
			balance = sum.TotalCredit - sum.TotalDebit
		} else if strings.HasPrefix(acc.Code, "3") {
			accType = "Modal"
			balance = sum.TotalCredit - sum.TotalDebit
		} else if strings.HasPrefix(acc.Code, "4") {
			accType = "Pendapatan"
			balance = sum.TotalCredit - sum.TotalDebit
			netProfit += balance
		} else if strings.HasPrefix(acc.Code, "5") {
			accType = "HPP"
			balance = sum.TotalDebit - sum.TotalCredit
			netProfit -= balance
		} else if strings.HasPrefix(acc.Code, "6") {
			accType = "Beban"
			balance = sum.TotalDebit - sum.TotalCredit
			netProfit -= balance
		}

		if balance != 0 || accType != "" {
			reportData = append(reportData, AccountBalance{
				Code:    acc.Code,
				Name:    acc.Name,
				Type:    accType,
				Balance: balance,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"data":       reportData,
		"net_profit": netProfit,
	})
}