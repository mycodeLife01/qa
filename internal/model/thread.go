package model

import "gorm.io/gorm"

type Thread struct {
	gorm.Model
	UUID   string `gorm:"unique;not null;size:36"`
	UserID uint   `gorm:"not null;index"`
	FileID uint   `gorm:"not null;index"`
	Title  string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
}
