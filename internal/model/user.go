package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string `gorm:"unique;not null;size:20"`
	PasswordHashed string `gorm:"not null;size:255"`
	Email          string `gorm:"unique;not null;size:255"`
	Role           string `gorm:"not null;size:32"`
}
