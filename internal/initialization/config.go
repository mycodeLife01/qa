package initialization

import (
	"github.com/mycodeLife01/qa/config"
)

// InitConfig 初始化配置
func InitConfig() error {
	if err := config.LoadConfig(); err != nil {
		return err
	}
	return nil
}
