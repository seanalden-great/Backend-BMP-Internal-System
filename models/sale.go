// models/sale.go
package models

import (
	"time"
)

type Sale struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	DueDate     time.Time `gorm:"type:date" json:"due_date"`
	ReferenceNo string    `gorm:"type:varchar(50);uniqueIndex" json:"reference_no"`
	ContactID   uint      `gorm:"not null;index" json:"contact_id"` // Relasi ke Customer
	Status      string    `gorm:"type:enum('Pesanan','Pengiriman','Tagihan','Lunas');default:'Pesanan'" json:"status"`
	Total       float64   `gorm:"type:decimal(15,2);default:0" json:"total"`
	Notes       string    `gorm:"type:text" json:"notes"`

	Contact     Contact    `gorm:"foreignKey:ContactID" json:"contact"`
	Items       []SaleItem `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE;" json:"items"`
}

type SaleItem struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	SaleID     uint    `gorm:"not null;index" json:"sale_id"`
	ProductSKU string  `gorm:"type:varchar(50);not null;index" json:"product_sku"`
	Quantity   int     `gorm:"type:int;not null;default:1" json:"quantity"`
	Price      float64 `gorm:"type:decimal(15,2);not null" json:"price"` // Harga Jual
	SubTotal   float64 `gorm:"type:decimal(15,2);not null" json:"sub_total"`

	Product    Product `gorm:"foreignKey:ProductSKU;references:SKU" json:"product"`
}