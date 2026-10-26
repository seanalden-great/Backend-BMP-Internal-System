// models/contact.go
package models

type Contact struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"type:varchar(150);not null" json:"name"`
	Type    string `gorm:"type:enum('Customer','Supplier');not null" json:"type"`
	Phone   string `gorm:"type:varchar(20)" json:"phone"`
	Email   string `gorm:"type:varchar(100)" json:"email"`
	Address string `gorm:"type:text" json:"address"`
}