// models/expense.go
package models

import (
	"time"
)

type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	ReferenceNo string    `gorm:"type:varchar(50)" json:"reference_no"`
	ContactID   *uint     `json:"contact_id"` // Opsional, bisa ada supplier atau tidak
	AccountCode string    `gorm:"type:varchar(20);not null" json:"account_code"` // Akun Beban (cth: 611-Gaji)
	PaymentCode string    `gorm:"type:varchar(20);not null" json:"payment_code"` // Akun Kas/Bank (cth: 111-Kas)
	Amount      float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`

	Contact     *Contact  `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	Account     Account   `gorm:"foreignKey:AccountCode;references:Code" json:"account"`
	PaymentAcc  Account   `gorm:"foreignKey:PaymentCode;references:Code" json:"payment_account"`
}