// models/faq.go
package models

type FAQ struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Category string `gorm:"type:varchar(50);not null" json:"category"`
	Question string `gorm:"type:text;not null" json:"question"`
	Answer   string `gorm:"type:text;not null" json:"answer"`
}