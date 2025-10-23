package initialization

import (
	"fmt"

	"github.com/mycodeLife01/qa/config"
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

	return db, nil
}
