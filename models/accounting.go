// models/accounting.go
package models

import "time"

// Model untuk Chart of Accounts
type Account struct {
	Code          string  `gorm:"primaryKey;type:varchar(20)" json:"code"`
	Name          string  `gorm:"type:varchar(100);not null" json:"name"`
	IsHeader      bool    `gorm:"default:false" json:"isHeader"`
	NormalBalance string  `gorm:"type:enum('D','K','-');default:'-'" json:"pos"`
	Level         int     `gorm:"type:int;default:0" json:"level"`
}

// Model untuk Header Jurnal
type Journal struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Date        time.Time      `gorm:"type:date;not null" json:"date"`
	ReferenceNo string         `gorm:"type:varchar(50)" json:"reference_no"`
	Description string         `gorm:"type:text" json:"description"`
	Entries     []JournalEntry `gorm:"foreignKey:JournalID" json:"entries"` // Relasi One-to-Many
}

// Model untuk Detail Jurnal (Debit/Kredit)
type JournalEntry struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	JournalID   uint    `gorm:"not null;index" json:"journal_id"`
	AccountCode string  `gorm:"type:varchar(20);not null;index" json:"account_code"`
	Debit       float64 `gorm:"type:decimal(15,2);default:0" json:"debit"`
	Credit      float64 `gorm:"type:decimal(15,2);default:0" json:"credit"`
	
	// Preload relasi untuk mendapatkan Nama Akun di tabel mutasi
	Account     Account `gorm:"foreignKey:AccountCode;references:Code" json:"account"` 
}