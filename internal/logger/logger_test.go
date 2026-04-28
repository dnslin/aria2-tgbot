package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 测试：各级别日志正常输出
func TestLogger_LogLevels(t *testing.T) {
	dir := t.TempDir()
	log, err := New(Config{
		Dir:        dir,
		Level:      "debug",
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	})
	if err != nil {
		t.Fatalf("创建 Logger 失败: %v", err)
	}
	defer log.Close()

	log.Debug("调试信息")
	log.Info("普通信息")
	log.Warn("警告信息")
	log.Error("错误信息")

	// 读取日志文件验证各条消息存在
	files, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	if len(files) == 0 {
		t.Fatal("未生成日志文件")
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	s := string(content)
	for _, expected := range []string{"[DEBUG]", "调试信息", "[INFO]", "普通信息", "[WARN]", "警告信息", "[ERROR]", "错误信息"} {
		if !strings.Contains(s, expected) {
			t.Errorf("日志应包含 %q, content=%s", expected, s)
		}
	}
}

// 测试：日志级别过滤，info 级别时不输出 debug 日志
func TestLogger_LevelFilter(t *testing.T) {
	dir := t.TempDir()
	log, err := New(Config{
		Dir:        dir,
		Level:      "info",
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	})
	if err != nil {
		t.Fatalf("创建 Logger 失败: %v", err)
	}
	defer log.Close()

	log.Debug("这条不应出现")
	log.Info("这条应出现")

	files, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	if len(files) == 0 {
		t.Fatal("未生成日志文件")
	}
	content, _ := os.ReadFile(files[0])
	s := string(content)
	if strings.Contains(s, "这条不应出现") {
		t.Error("debug 级别日志应被过滤")
	}
	if !strings.Contains(s, "这条应出现") {
		t.Error("info 级别日志应出现")
	}
}

// 测试：日志包含调用位置信息
func TestLogger_CallerInfo(t *testing.T) {
	dir := t.TempDir()
	log, err := New(Config{
		Dir:        dir,
		Level:      "debug",
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	})
	if err != nil {
		t.Fatalf("创建 Logger 失败: %v", err)
	}
	defer log.Close()

	log.Info("定位测试")

	files, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	content, _ := os.ReadFile(files[0])
	s := string(content)

	// 断言日志包含调用位置（logger_test.go:行号:TestLogger_CallerInfo）
	if !strings.Contains(s, "logger_test.go") {
		t.Errorf("日志应包含调用文件信息, content=%s", s)
	}
}

// 测试：日志目录不存在时自动创建
func TestLogger_AutoCreateDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new-subdir", "logs")
	log, err := New(Config{
		Dir:        dir,
		Level:      "info",
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	})
	if err != nil {
		t.Fatalf("自动创建目录时应成功: %v", err)
	}
	defer log.Close()

	log.Info("目录已创建")

	files, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	if len(files) == 0 {
		t.Fatal("日志目录应被自动创建并写入文件")
	}
}
