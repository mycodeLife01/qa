package initialization

import (
	"log"

	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/service/impl"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

// InitServices 初始化所有服务
func InitServices(db *gorm.DB, cosClient *cos.Client) *Services {
	// 初始化Python服务客户端
	pythonClient := client.NewPythonServiceClient()

	// 初始化Celery客户端
	taskClient, err := client.NewCeleryClient(
		config.C.Redis.BusinessURL,
		config.C.Redis.BrokerURL,
		config.C.Redis.ResultURL,
		"celery", // 默认队列名
	)
	if err != nil {
		log.Fatalf("初始化Celery客户端失败: %v", err)
	}

	return &Services{
		AuthService:    impl.NewAuthService(db),
		UserService:    impl.NewUserService(db),
		FileService:    impl.NewFileService(db, cosClient, taskClient),
		AiService:      impl.NewAiService(db, pythonClient, cosClient, taskClient),
		ThreadService:  impl.NewThreadService(db),
		MessageService: impl.NewMessageService(db),
	}
}
