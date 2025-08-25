package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	Env    string       `yaml:"env"` // 运行环境: dev, prod
	Auth   AuthConfig   `yaml:"auth"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port      int    `yaml:"port"`
	Host      string `yaml:"host"`
	JwtSecret string `yaml:"jwt_secret"`
	JwtExpire int    `yaml:"jwt_expire"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	RequireLogin bool `yaml:"require_login"` // 是否需要登录注册，默认为false
}

var conf *Config

// Load 从 YAML 文件加载配置。默认从 configs/config.yaml 读取，也可通过环境变量 CONFIG_PATH 指定路径
func Load(force bool) (*Config, error) {
	if conf != nil && !force {
		return conf, nil
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 检查是否为Wails应用模式
		wailsConfigDir := os.Getenv("VIDEO_CRAWLER_CONFIG_DIR")
		if wailsConfigDir != "" {
			configPath = os.Getenv("VIDEO_CRAWLER_CONFIG_DIR") + "/config.yaml"
		} else {
			configPath = "configs/config.yaml"
		}
	}

	logrus.WithFields(logrus.Fields{
		"configPath": configPath,
	}).Info("加载配置文件")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("解析 YAML 配置失败: %w", err)
	}

	// 设置默认值
	if conf.Auth.RequireLogin == false {
		conf.Auth.RequireLogin = false // 默认为false
	}

	return &conf, nil
}

// GetDataDir 获取数据目录路径
func GetDataDir() string {
	// 检查是否为Wails应用模式
	wailsConfigDir := os.Getenv("VIDEO_CRAWLER_CONFIG_DIR")
	if wailsConfigDir != "" {
		return wailsConfigDir
	}

	// 非Wails模式，使用相对路径
	return "configs"
}
