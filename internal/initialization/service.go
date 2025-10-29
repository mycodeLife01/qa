package initialization

import (
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/service/impl"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

// InitServices 初始化所有服务
func InitServices(db *gorm.DB, cosClient *cos.Client) *Services {
	// 初始化Python服务客户端
	pythonClient := client.NewPythonServiceClient()

	return &Services{
		AuthService: impl.NewAuthService(db),
		UserService: impl.NewUserService(db),
		FileService: impl.NewFileService(db, cosClient),
		AiService:   impl.NewAiService(db, pythonClient, cosClient),
	}
}
