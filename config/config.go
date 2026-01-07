package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	COS      COSConfig      `mapstructure:"cos"`
	Services ServicesConfig `mapstructure:"services"`
	System   SystemConfig   `mapstructure:"index"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type ServerConfig struct {
	Port        int `mapstructure:"port"`
	ReadTimeout int `mapstructure:"read_timeout"`
}

type DatabaseConfig struct {
	DatabaseURL string `mapstructure:"url"`
}

type JWTConfig struct {
	JWTSecretKey string `mapstructure:"secretKey"`
}

type COSConfig struct {
	SecretID     string `mapstructure:"secretId"`
	SecretKey    string `mapstructure:"secretKey"`
	UploadFolder string `mapstructure:"upload_folder"`
}

type ServicesConfig struct {
	QAAgentURL       string `mapstructure:"qa_agent_url"`
	QAIndexWorkerURL string `mapstructure:"qa_index_worker_url"`
}

type RedisConfig struct {
	BrokerURL   string `mapstructure:"brokerUrl"`
	ResultURL   string `mapstructure:"resultUrl"`
	BusinessURL string `mapstructure:"businessUrl"`
}
type SystemConfig struct {
	IndexMode int `mapstructure:"mode"`
}

// C 是一个全局变量，用于在其他包中访问配置
var C Config

// LoadConfig 从文件和环境变量中加载配置
func LoadConfig() (err error) {

	// 1. 确定配置文件名
	configName := getConfigName()

	// 2. 设置配置文件路径
	viper.SetConfigName(configName) // 配置文件名 (不带后缀)
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config")

	// 3. 读取环境变量
	viper.BindEnv("cos.secretId", "COS_SECRET_ID")
	viper.BindEnv("cos.secretKey", "COS_SECRET_KEY")
	viper.BindEnv("jwt.secretKey", "JWT_SECRET_KEY")
	viper.BindEnv("database.url", "DATABASE_URL")
	viper.BindEnv("redis.brokerUrl", "BROKER_REDIS_URL")
	viper.BindEnv("redis.resultUrl", "RESULT_REDIS_URL")
	viper.BindEnv("redis.businessUrl", "BUSINESS_REDIS_URL")
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
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// 5. 将所有配置反序列化到 Config 结构体中
	err = viper.Unmarshal(&C)
	if err != nil {
		return fmt.Errorf("unable to decode into struct, %v", err)
	}
	fmt.Printf("Server port: %v\n", C.Server.Port)

	return nil
}

// getConfigName 根据环境变量或本地文件存在性确定配置文件名
func getConfigName() string {
	// 优先级1: 通过 CONFIG_NAME 环境变量显式指定
	if configName := os.Getenv("CONFIG_NAME"); configName != "" {
		return configName
	}

	// 优先级2: 通过 GO_ENV/APP_ENV 环境变量指定环境
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env != "" {
		return "config." + env // 例如: config.local, config.prod
	}

	// 优先级3: 自动检测 config.local.yaml 是否存在
	if _, err := os.Stat("./config/config.local.yaml"); err == nil {
		return "config.local"
	}

	// 默认使用 config.yaml
	return "config"
}
