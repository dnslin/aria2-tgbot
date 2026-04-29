package aria2

import (
	"strings"
	"testing"
)

// 测试：GenerateSecret 生成 16 位长度密钥
func TestGenerateSecret_Length(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}
	if len(secret) != 16 {
		t.Errorf("密钥长度应为 16, got=%d, secret=%s", len(secret), secret)
	}
}

// 测试：GenerateSecret 每次生成不同密钥
func TestGenerateSecret_Unique(t *testing.T) {
	secrets := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s, err := GenerateSecret()
		if err != nil {
			t.Fatalf("生成密钥失败: %v", err)
		}
		if secrets[s] {
			t.Errorf("密钥冲突: %s", s)
		}
		secrets[s] = true
	}
}

// 测试：GenerateSecret 仅含字母数字
func TestGenerateSecret_Alphanumeric(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, c := range secret {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("密钥含非法字符: %c", c)
		}
	}
}

// 测试：GenerateConfig 包含所有关键配置段
func TestGenerateConfig_ContainsSections(t *testing.T) {
	params := ConfigParams{
		DownloadDir: "/tmp/downloads",
		RpcPort:     6800,
		RpcSecret:   "testSecret12345",
		SessionDir:  "/tmp/session",
		ConfigDir:   "/tmp/config",
	}
	content := GenerateConfig(params)

	checks := []string{
		"dir=/tmp/downloads",
		"rpc-listen-port=6800",
		"rpc-secret=testSecret12345",
		"save-session=/tmp/session/aria2.session",
		"dht-file-path=/tmp/config/dht.dat",
		"# ===== 基本设置",
		"# ===== RPC 设置",
		"# ===== 网络连接设置",
		"# ===== BT/PT 设置",
		"# ===== 高级设置",
		"disk-cache=64M",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("配置应包含 %q", check)
		}
	}
}

// 测试：GenerateConfig 不同参数生成不同配置
func TestGenerateConfig_DifferentParams(t *testing.T) {
	a := GenerateConfig(ConfigParams{
		DownloadDir: "/a",
		RpcPort:     6800,
		RpcSecret:   "secretA",
		SessionDir:  "/sessionA",
		ConfigDir:   "/configA",
	})
	b := GenerateConfig(ConfigParams{
		DownloadDir: "/b",
		RpcPort:     6801,
		RpcSecret:   "secretB",
		SessionDir:  "/sessionB",
		ConfigDir:   "/configB",
	})
	if a == b {
		t.Error("不同参数应生成不同配置")
	}
	if !strings.Contains(a, "dir=/a") || !strings.Contains(b, "dir=/b") {
		t.Error("配置应包含对应参数值")
	}
}
