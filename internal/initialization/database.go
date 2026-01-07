package initialization

import (
	"fmt"

	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDatabase 初始化数据库连接
func InitDatabase() (*gorm.DB, error) {
	dsn := config.C.Database.DatabaseURL
	fmt.Println("Database URL: ", dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移所有模型
	migrateErr := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(
		&model.User{},
		&model.File{},
		&model.UserFile{},
		&model.Thread{},
		&model.Message{},
	)
	if migrateErr != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", migrateErr)
	}

	return db, nil
}
