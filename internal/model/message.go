package model

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	ThreadID uint   `gorm:"not null;index"`
	Role     string `gorm:"not null;size:20"` // user, assistant
	Content  string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
}
