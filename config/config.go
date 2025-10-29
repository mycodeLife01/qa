package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	COS      COSConfig
	Services ServicesConfig
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

type COSConfig struct {
	SecretID  string `mapstructure:"secretId"`
	SecretKey string `mapstructure:"secretKey"`
}

type ServicesConfig struct {
	QAAgentURL       string `mapstructure:"qa_agent_url"`
	QAIndexWorkerURL string `mapstructure:"qa_index_worker_url"`
}

// C 是一个全局变量，用于在其他包中访问配置
var C Config

// LoadConfig 从文件和环境变量中加载配置
func LoadConfig() (err error) {

	// 2. 设置配置文件路径
	viper.SetConfigName("config") // 配置文件名 (不带后缀)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath("./config")

	// 3. 读取环境变量
	viper.BindEnv("cos.secretId", "COS_SECRET_ID")
	viper.BindEnv("cos.secretKey", "COS_SECRET_KEY")
	viper.BindEnv("jwt.secretKey", "JWT_SECRET_KEY")
	viper.BindEnv("database.url", "DATABASE_URL")

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
