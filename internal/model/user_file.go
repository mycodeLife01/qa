package model

import "gorm.io/gorm"

// UserFile 用户-文件关联中间表，支持多对多关系
type UserFile struct {
	gorm.Model
	UserID   uint   `gorm:"not null;uniqueIndex:idx_user_file"`
	FileID   uint   `gorm:"not null;uniqueIndex:idx_user_file;index"`
	FileName string `gorm:"not null;size:255"` // 用户自定义的文件名

	// 关联关系
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	File File `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE"`
}
