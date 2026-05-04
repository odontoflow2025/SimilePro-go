package models

import (
	"time"

	"gorm.io/gorm"
)

type Ticket struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Protocol  string         `gorm:"uniqueIndex;not null" json:"protocol"`
	Email     string         `gorm:"not null" json:"email"`
	Subject   string         `gorm:"not null" json:"subject"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
