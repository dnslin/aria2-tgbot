package aria2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 测试：logf 回调正常输出
func TestInstaller_Logf(t *testing.T) {
	var logs []string

	inst := NewInstaller("127.0.0.1", 6800, "/tmp", "/tmp", "/tmp", "/tmp", false)
	inst.SetLogCallback(func(format string, args ...any) {
		logs = append(logs, fmt.Sprintf(format, args...))
	})

	inst.logf("测试消息 %s", "hello")
	inst.logf("第 %d 条", 2)

	if len(logs) != 2 {
		t.Fatalf("应有 2 条日志, got=%d", len(logs))
	}
	if logs[0] != "测试消息 hello" {
		t.Errorf("第一条应为 测试消息 hello, got=%s", logs[0])
	}
	if logs[1] != "第 2 条" {
		t.Errorf("第二条应为 第 2 条, got=%s", logs[1])
	}
}

// 测试：未设置回调时 logf 不崩溃
func TestInstaller_LogfNilCallback(t *testing.T) {
	inst := NewInstaller("127.0.0.1", 6800, "/tmp", "/tmp", "/tmp", "/tmp", false)
	// 未设置回调，不应崩溃
	inst.logf("这条消息应该被丢弃")
}

// 测试：IsInstalled 检测逻辑
func TestInstaller_IsInstalled(t *testing.T) {
	// 创建临时目录结构，模拟已安装的 aria2c
	dir := t.TempDir()
	binPath := filepath.Join(dir, "aria2c")
	os.WriteFile(binPath, []byte("fake binary"), 0755)

	inst := NewInstaller("127.0.0.1", 6800, dir, "/tmp", "/tmp", "/tmp", false)
	if !inst.IsInstalled() {
		t.Error("应检测到已安装")
	}

	// 指向不存在的路径
	inst2 := NewInstaller("127.0.0.1", 6800, "/nonexistent/path", "/tmp", "/tmp", "/tmp", false)
	if inst2.IsInstalled() {
		t.Error("应检测到未安装")
	}
}

// 测试：NewInstaller 正确存储参数
func TestInstaller_NewInstaller(t *testing.T) {
	inst := NewInstaller("192.168.1.1", 6801, "/usr/bin", "/etc/aria2", "/data/downloads", "/data/session", true)

	if inst.rpcHost != "192.168.1.1" {
		t.Errorf("rpcHost 应为 192.168.1.1, got=%s", inst.rpcHost)
	}
	if inst.rpcPort != 6801 {
		t.Errorf("rpcPort 应为 6801, got=%d", inst.rpcPort)
	}
	if inst.installPath != "/usr/bin" {
		t.Errorf("installPath 应为 /usr/bin, got=%s", inst.installPath)
	}
	if inst.configDir != "/etc/aria2" {
		t.Errorf("configDir 应为 /etc/aria2, got=%s", inst.configDir)
	}
	if inst.downloadDir != "/data/downloads" {
		t.Errorf("downloadDir 应为 /data/downloads, got=%s", inst.downloadDir)
	}
	if inst.sessionDir != "/data/session" {
		t.Errorf("sessionDir 应为 /data/session, got=%s", inst.sessionDir)
	}
	if !inst.autoStart {
		t.Error("autoStart 应为 true")
	}
}

// 测试：formatBytes 格式化
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, want=%s", tt.bytes, result, tt.expected)
		}
	}
}

// 测试：formatBytes 始终是 xx.x XB 格式
func TestFormatBytes_Format(t *testing.T) {
	result := formatBytes(5368709120) // 5 GB
	if !strings.Contains(result, " ") {
		t.Errorf("结果应包含空格分隔, got=%s", result)
	}
}

// 测试：hasSystemd 检测（仅验证不崩溃）
func TestInstaller_HasSystemd(t *testing.T) {
	inst := NewInstaller("127.0.0.1", 6800, "/tmp", "/tmp", "/tmp", "/tmp", false)
	// hasSystemd 是内部方法，不导出，通过反射或直接测试行为
	// 这里仅验证函数可调用且不 panic
	result := inst.hasSystemd()
	t.Logf("系统是否使用 systemd: %v", result)
}
