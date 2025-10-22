package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port        int
	ReadTimeout int
}

type DatabaseConfig struct {
	DatabaseURL string `mapstructure:"url"`
}

type JWTConfig struct {
	JWTSecretKey string `mapstructure:"secretKey"`
}

// C 是一个全局变量，用于在其他包中访问配置
var C Config

// LoadConfig 从文件和环境变量中加载配置
func LoadConfig() (err error) {
	// 1. 设置默认值
	viper.SetDefault("server.port", 8080)

	// 2. 设置配置文件路径
	viper.SetConfigName("config") // 配置文件名 (不带后缀)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath("./config")

	// 3. 读取环境变量
	viper.AutomaticEnv() // 自动读取环境变量
	// 例如，要覆盖数据库密码，可以设置环境变量：DATABASE_PASSWORD=prod_password
	// Viper 会自动将 . 替换为 _ 来匹配环境变量名
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 4. 读取配置文件
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，可以忽略，因为可能只使用环境变量
			fmt.Println("Config file not found; relying on environment variables and defaults.")
		} else {
			// 配置文件被找到但解析错误
			return err
		}
	}

	// 5. 将所有配置反序列化到 Config 结构体中
	err = viper.Unmarshal(&C)
	if err != nil {
		return fmt.Errorf("unable to decode into struct, %v", err)
	}

	return nil
}
