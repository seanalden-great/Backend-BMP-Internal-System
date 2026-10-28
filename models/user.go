// models/user.go
package models

import (
	"time"
)

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password    string    `gorm:"type:varchar(255);not null" json:"-"` // JSON di-ignore demi keamanan
	FullName    string    `gorm:"type:varchar(100)" json:"full_name"`
	Role        string    `gorm:"type:enum('Admin','Manager','Staff');default:'Staff'" json:"role"`
	Permissions string    `gorm:"type:text" json:"permissions"` // Menyimpan modul yg diizinkan (misal: "sales,purchase,inventory")
	CreatedAt   time.Time `json:"created_at"`
}