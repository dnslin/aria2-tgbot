// Package config 负责 YAML 配置文件的加载、校验、默认值填充和运行时持久化。
// 配置文件通过 -c 命令行参数指定，默认 ./config.yaml。
// 敏感字段（bot.token）支持从环境变量覆盖。
package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// BotConfig Bot 基础配置
type BotConfig struct {
	Token string `yaml:"token"` // Telegram Bot Token，支持环境变量 BOT_TOKEN
	Debug bool   `yaml:"debug"`
}

// AuthConfig 权限控制配置
type AuthConfig struct {
	Enabled    *bool   `yaml:"enabled"` // 指针以区分"未设置"和"显式设为 false"
	AllowUsers []int64 `yaml:"allow_users"`
}

// IsEnabled 返回权限是否开启，默认启用。
func (a *AuthConfig) IsEnabled() bool {
	if a.Enabled == nil {
		return true
	}
	return *a.Enabled
}

// IsAutoDeleteEnabled 返回全局自动删除是否开启，默认启用。
func (m *MessageConfig) IsAutoDeleteEnabled() bool {
	if m.AutoDelete == nil {
		return true
	}
	return *m.AutoDelete
}

// MessageConfig 消息行为配置，运行时可通过命令修改并持久化
type MessageConfig struct {
	AutoDelete              *bool `yaml:"auto_delete"`
	DeleteAfter             int  `yaml:"delete_after"`
	ResultDeleteAfter       int  `yaml:"result_delete_after"`
	ProgressUpdateInterval  int  `yaml:"progress_update_interval"`
	InstallLogDelete        bool `yaml:"install_log_delete"`
	ErrorDeleteAfter        int  `yaml:"error_delete_after"`
	NotifyDeleteAfter       int  `yaml:"notify_delete_after"`
}

// Aria2Config aria2 安装和 RPC 连接配置
type Aria2Config struct {
	InstallScriptURL string `yaml:"install_script_url"`
	InstallPath      string `yaml:"install_path"`
	ConfigDir        string `yaml:"config_dir"`
	DownloadDir      string `yaml:"download_dir"`
	RpcHost          string `yaml:"rpc_host"`
	RpcPort          int    `yaml:"rpc_port"`
	RpcSecret        string `yaml:"rpc_secret"`
	SessionDir       string `yaml:"session_dir"`
	AutoStart        bool   `yaml:"auto_start"`
	EnableSystemd    bool   `yaml:"enable_systemd"`
}

// LogConfig 日志配置
type LogConfig struct {
	Dir        string `yaml:"dir"`
	Level      string `yaml:"level"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// Config 聚合所有配置项
type Config struct {
	Bot     BotConfig     `yaml:"bot"`
	Auth    AuthConfig    `yaml:"auth"`
	Message MessageConfig `yaml:"message"`
	Aria2   Aria2Config   `yaml:"aria2"`
	Log     LogConfig     `yaml:"log"`

	mu       sync.RWMutex `yaml:"-"`
	filePath string       `yaml:"-"`
}

// Load 从指定路径加载 YAML 配置文件，填充默认值，并从环境变量覆盖敏感字段。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{filePath: path}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.SetDefaults()

	// 环境变量覆盖 Token
	if envToken := os.Getenv("BOT_TOKEN"); envToken != "" {
		cfg.Bot.Token = envToken
	}

	return cfg, nil
}

// SetDefaults 为未设置的字段填充默认值。
func (c *Config) SetDefaults() {
	if c.Message.DeleteAfter == 0 {
		c.Message.DeleteAfter = 180
	}
	if c.Message.ResultDeleteAfter == 0 {
		c.Message.ResultDeleteAfter = 300
	}
	if c.Message.ProgressUpdateInterval == 0 {
		c.Message.ProgressUpdateInterval = 30
	}
	if c.Message.AutoDelete == nil {
		autoDelete := true
		c.Message.AutoDelete = &autoDelete
	}
	if c.Auth.Enabled == nil {
		enabled := true
		c.Auth.Enabled = &enabled
	}
	if c.Aria2.RpcHost == "" {
		c.Aria2.RpcHost = "127.0.0.1"
	}
	if c.Aria2.RpcPort == 0 {
		c.Aria2.RpcPort = 6800
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.MaxSize == 0 {
		c.Log.MaxSize = 10
	}
	if c.Log.MaxBackups == 0 {
		c.Log.MaxBackups = 7
	}
	if c.Log.MaxAge == 0 {
		c.Log.MaxAge = 30
	}
	if c.Log.Dir == "" {
		c.Log.Dir = "./logs"
	}
	if c.Aria2.InstallScriptURL == "" {
		c.Aria2.InstallScriptURL = "https://git.io/aria2.sh"
	}
	if c.Aria2.InstallPath == "" {
		c.Aria2.InstallPath = "/usr/local/bin"
	}
	if c.Aria2.ConfigDir == "" {
		c.Aria2.ConfigDir = "/root/.aria2c"
	}
	if c.Aria2.DownloadDir == "" {
		c.Aria2.DownloadDir = "/root/downloads"
	}
	if c.Aria2.SessionDir == "" {
		c.Aria2.SessionDir = "/root/.aria2c"
	}
}

// FilePath 返回配置文件路径。
func (c *Config) FilePath() string {
	return c.filePath
}

// Save 将当前配置序列化为 YAML 写回文件。
// 使用互斥锁防止并发写入竞态。
func (c *Config) Save(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
