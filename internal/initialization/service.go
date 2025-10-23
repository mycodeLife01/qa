package initialization

import (
	"github.com/mycodeLife01/qa/internal/service/impl"
	"gorm.io/gorm"
)

// InitServices 初始化所有服务
func InitServices(db *gorm.DB) *Services {
	return &Services{
		AuthService: impl.NewAuthService(db),
	}
}
