package model

import "gorm.io/gorm"

type File struct {
	gorm.Model
	ContentHash string `gorm:"not null;size:255"`
	ObjectKey   string `gorm:"not null;size:255"`
	BucketName  string `gorm:"not null;size:255"`
	FileType    string `gorm:"not null;size:32"`
}
