// models/product.go
package models

type Product struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	SKU       string  `gorm:"type:varchar(50);uniqueIndex;not null" json:"sku"`
	Name      string  `gorm:"type:varchar(150);not null" json:"name"`
	Category  string  `gorm:"type:varchar(50)" json:"category"`
	Unit      string  `gorm:"type:varchar(20)" json:"unit"` // Kolom Satuan
	BuyPrice  float64 `gorm:"type:decimal(15,2);default:0" json:"buyPrice"`
	SellPrice float64 `gorm:"type:decimal(15,2);default:0" json:"sellPrice"`
	HPP       float64 `gorm:"type:decimal(15,2);default:0" json:"hpp"` // Kolom HPP
	Stock     int     `gorm:"type:int;default:0" json:"stock"`
}