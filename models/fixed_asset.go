// models/fixed_asset.go
package models

import (
	"time"
)

type FixedAsset struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"type:varchar(100);not null" json:"name"`
	PurchaseDate      time.Time `gorm:"type:date;not null" json:"purchase_date"`
	PurchasePrice     float64   `gorm:"type:decimal(15,2);not null" json:"purchase_price"`
	UsefulLife        int       `gorm:"not null" json:"useful_life"` // Umur ekonomis dalam satuan bulan
	AccumDepreciation float64   `gorm:"type:decimal(15,2);default:0" json:"accum_depreciation"` // Total yang sudah disusutkan
	Status            string    `gorm:"type:enum('Aktif','Lunas Disusutkan','Dijual');default:'Aktif'" json:"status"`
}