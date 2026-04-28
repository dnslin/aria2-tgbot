package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试：加载有效的 YAML 配置文件
func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
bot:
  token: "test123:abc"
  debug: true
auth:
  enabled: true
  allow_users:
    - 111
    - 222
message:
  auto_delete: true
  delete_after: 60
  result_delete_after: 120
  progress_update_interval: 15
  install_log_delete: false
  error_delete_after: 30
  notify_delete_after: 60
aria2:
  install_script_url: "https://example.com/aria2.sh"
  install_path: "/usr/bin"
  config_dir: "/tmp/aria2"
  download_dir: "/tmp/downloads"
  rpc_host: "127.0.0.1"
  rpc_port: 6800
  rpc_secret: ""
  session_dir: "/tmp/aria2"
  auto_start: true
  enable_systemd: true
log:
  dir: "/var/log/bot"
  level: "debug"
  max_size: 20
  max_backups: 10
  max_age: 60
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Bot.Token != "test123:abc" {
		t.Errorf("Token 不匹配, got=%s", cfg.Bot.Token)
	}
	if !cfg.Bot.Debug {
		t.Error("Debug 应为 true")
	}
	if !cfg.Auth.IsEnabled() {
		t.Error("Auth.Enabled 应为 true")
	}
	if len(cfg.Auth.AllowUsers) != 2 {
		t.Errorf("AllowUsers 长度应为 2, got=%d", len(cfg.Auth.AllowUsers))
	}
	if cfg.Aria2.RpcPort != 6800 {
		t.Errorf("RpcPort 应为 6800, got=%d", cfg.Aria2.RpcPort)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level 应为 debug, got=%s", cfg.Log.Level)
	}
}

// 测试：缺失字段自动填充默认值
func TestSetDefaults_FillsMissingValues(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefaults()

	if cfg.Bot.Debug != false {
		t.Error("Debug 默认值应为 false")
	}
	if !cfg.Auth.IsEnabled() {
		t.Error("Auth.Enabled 默认值应为 true")
	}
	if cfg.Message.AutoDelete == nil || *cfg.Message.AutoDelete != true {
		t.Error("Message.AutoDelete 默认值应为 true")
	}
	if cfg.Message.DeleteAfter != 180 {
		t.Errorf("Message.DeleteAfter 默认值应为 180, got=%d", cfg.Message.DeleteAfter)
	}
	if cfg.Message.ProgressUpdateInterval != 30 {
		t.Errorf("ProgressUpdateInterval 默认值应为 30, got=%d", cfg.Message.ProgressUpdateInterval)
	}
	if cfg.Aria2.RpcHost != "127.0.0.1" {
		t.Error("RpcHost 默认值应为 127.0.0.1")
	}
	if cfg.Aria2.RpcPort != 6800 {
		t.Error("RpcPort 默认值应为 6800")
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level 默认值应为 info, got=%s", cfg.Log.Level)
	}
	if cfg.Log.MaxSize != 10 {
		t.Errorf("Log.MaxSize 默认值应为 10, got=%d", cfg.Log.MaxSize)
	}
}

// 测试：Token 支持从环境变量覆盖
func TestLoad_TokenFromEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
bot:
  token: ""
auth:
  enabled: false
`
	os.WriteFile(path, []byte(content), 0644)
	os.Setenv("BOT_TOKEN", "env_token_123")
	defer os.Unsetenv("BOT_TOKEN")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Bot.Token != "env_token_123" {
		t.Errorf("Token 应从环境变量读取, got=%s", cfg.Bot.Token)
	}
}

// 测试：配置持久化回写 + 重新读取一致
func TestSaveAndReload_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
bot:
  token: "original"
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	cfg.Bot.Token = "modified"
	cfg.Bot.Debug = true

	if err := cfg.Save(path); err != nil {
		t.Fatalf("保存配置失败: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("重新加载配置失败: %v", err)
	}

	if reloaded.Bot.Token != "modified" {
		t.Errorf("Token 应为 modified, got=%s", reloaded.Bot.Token)
	}
	if !reloaded.Bot.Debug {
		t.Error("Debug 应为 true")
	}
}

// 测试：配置文件不存在时返回错误
func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("不存在的文件应返回错误")
	}
}

// 测试：无效 YAML 返回错误
func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("this: [is not valid yaml: {{{"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("无效 YAML 应返回错误")
	}
}
