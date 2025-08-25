package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	Env    string       `yaml:"env"` // 运行环境: dev, prod
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port      int    `yaml:"port"`
	Host      string `yaml:"host"`
	JwtSecret string `yaml:"jwt_secret"`
	JwtExpire int    `yaml:"jwt_expire"`
}

var conf *Config

// Load 从 YAML 文件加载配置。默认从 configs/config.yaml 读取，也可通过环境变量 CONFIG_PATH 指定路径
func Load(force bool) (*Config, error) {
	if conf != nil && !force {
		return conf, nil
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("解析 YAML 配置失败: %w", err)
	}

	return &conf, nil
}
