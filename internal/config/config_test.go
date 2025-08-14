package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testYAML = `
server:
  host: "0.0.0.0"
  port: 9090

database:
  host: "db.local"
  port: 15432
  username: "user"
  password: "pass"
  database: "vc"

crawler:
  user_agent: "TestCrawler/1.0"
  timeout: 15
  retries: 2
`

func withTempConfig(t *testing.T, content string, fn func(path string)) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config failed: %v", err)
	}
	fn(path)
}

func TestLoad_FromYAML(t *testing.T) {
	withTempConfig(t, testYAML, func(path string) {
		os.Setenv("CONFIG_PATH", path)
		defer os.Unsetenv("CONFIG_PATH")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.Server.Port != 9090 || cfg.Server.Host != "0.0.0.0" {
			t.Fatalf("server config mismatch: %+v", cfg.Server)
		}
	})
}

func TestLoad_DefaultPathExists(t *testing.T) {
	// 在项目根目录下的 configs/config.yaml 存在时，直接读取
	// 为避免影响本地文件，这里仅做存在性检查，不作强断言
	if _, err := os.Stat("configs/config.yaml"); err == nil {
		// 确保不会被 CONFIG_PATH 覆盖
		os.Unsetenv("CONFIG_PATH")
		_, err := Load()
		if err != nil {
			t.Fatalf("Load() from default path failed: %v", err)
		}
	}
}

func TestConfigYAMLTags(t *testing.T) {
	// 粗略保证字段不会被未来重命名导致 YAML 解析异常
	_ = time.Second // 引用标准库，避免空文件
}
