package model

import "gorm.io/gorm"

// File 文件表，存储文件的物理信息（去重存储）
type File struct {
	gorm.Model
	ContentHash string `gorm:"unique;not null;size:255"` // 文件内容哈希，唯一约束保证去重
	ObjectKey   string `gorm:"not null;size:255"`
	BucketName  string `gorm:"not null;size:255"`
	FileType    string `gorm:"not null;size:32"`
	Status      string `gorm:"default:'PENDING';size:32"` // PENDING, PROCESSING, SUCCESS, FAILURE
}
