// models/purchase.go
package models

import (
	"time"
)

// Purchase: Tabel Header Transaksi Pembelian
type Purchase struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	DueDate     time.Time `gorm:"type:date" json:"due_date"` // Tanggal jatuh tempo untuk Tagihan
	ReferenceNo string    `gorm:"type:varchar(50);uniqueIndex" json:"reference_no"`
	ContactID   uint      `gorm:"not null;index" json:"contact_id"` // Relasi ke Supplier
	
	// Status menentukan fase dokumen: 'Pesanan', 'Pengiriman', 'Tagihan', 'Lunas'
	Status      string    `gorm:"type:enum('Pesanan','Pengiriman','Tagihan','Lunas');default:'Pesanan'" json:"status"`
	
	Total       float64   `gorm:"type:decimal(15,2);default:0" json:"total"`
	Notes       string    `gorm:"type:text" json:"notes"`

	// Preload Relasi
	Contact     Contact        `gorm:"foreignKey:ContactID" json:"contact"`
	Items       []PurchaseItem `gorm:"foreignKey:PurchaseID;constraint:OnDelete:CASCADE;" json:"items"`
}

// PurchaseItem: Tabel Detail Barang yang Dibeli
type PurchaseItem struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	PurchaseID uint    `gorm:"not null;index" json:"purchase_id"`
	ProductSKU string  `gorm:"type:varchar(50);not null;index" json:"product_sku"`
	Quantity   int     `gorm:"type:int;not null;default:1" json:"quantity"`
	Price      float64 `gorm:"type:decimal(15,2);not null" json:"price"` // Harga Beli dari Supplier
	SubTotal   float64 `gorm:"type:decimal(15,2);not null" json:"sub_total"` // Qty * Price

	// Preload Relasi
	Product    Product `gorm:"foreignKey:ProductSKU;references:SKU" json:"product"`
}