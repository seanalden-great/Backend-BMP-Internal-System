// models/setting.go
package models

type CompanySetting struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	Name    string  `gorm:"type:varchar(100);not null;default:'PT Berlianmandiri Perkasa'" json:"name"`
	Address string  `gorm:"type:text" json:"address"`
	Phone   string  `gorm:"type:varchar(20)" json:"phone"`
	Email   string  `gorm:"type:varchar(100)" json:"email"`
	TaxRate float64 `gorm:"type:decimal(5,2);default:11.0" json:"tax_rate"` // Persentase PPN default
}