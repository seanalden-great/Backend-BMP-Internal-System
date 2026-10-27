// models/inventory.go
package models

import (
	"time"
)

type StockAdjustment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Date       time.Time `gorm:"type:date;not null" json:"date"`
	ProductSKU string    `gorm:"type:varchar(50);not null;index" json:"product_sku"`
	Type       string    `gorm:"type:enum('Penambahan','Pengurangan');not null" json:"type"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	Reason     string    `gorm:"type:text" json:"reason"`
	TotalValue float64   `gorm:"type:decimal(15,2)" json:"total_value"`

	Product    Product   `gorm:"foreignKey:ProductSKU;references:SKU" json:"product"`
}