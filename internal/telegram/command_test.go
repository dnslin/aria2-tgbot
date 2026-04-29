package telegram

import (
	"testing"
)

// 创建一个最小 Handler 用于测试命令注册表。
func newTestHandler() *Handler {
	h := &Handler{}
	h.commands = h.RegisterCommands()
	return h
}

// 验证命令注册表完整性：30 个命令无重复。
func TestRegisterCommands_AllCommandsRegistered(t *testing.T) {
	h := newTestHandler()
	expectedCommands := []string{
		"/help", "/ping", "/health",
		"/install", "/uninstall", "/upgrade",
		"/aria_start", "/aria_stop", "/aria_restart", "/aria_status",
		"/add", "/add_magnet", "/pause", "/resume", "/remove",
		"/list", "/done", "/clear",
		"/limit_global", "/limit_task",
		"/conf", "/conf_dir", "/conf_rpc_port", "/conf_secret", "/reset_secret",
		"/stats",
		"/msg_config", "/msg_autodel", "/msg_time", "/msg_result", "/msg_error", "/msg_notify", "/msg_progress",
	}

	if len(h.commands) != len(expectedCommands) {
		t.Errorf("命令数量不符: 期望 %d, 实际 %d", len(expectedCommands), len(h.commands))
	}

	for _, name := range expectedCommands {
		if _, ok := h.commands[name]; !ok {
			t.Errorf("缺命令: %s", name)
		}
	}
}

// 验证命令 Name 字段与注册键一致。
func TestRegisterCommands_NameMatchesKey(t *testing.T) {
	h := newTestHandler()
	for key, cmd := range h.commands {
		expectedName := key[1:] // 去掉 "/"
		if cmd.Name != expectedName {
			t.Errorf("命令 %s 的 Name 字段为 %q, 期望 %q", key, cmd.Name, expectedName)
		}
	}
}

// 验证 NeedAria2 和 NeedRunning 标志正确。
func TestRegisterCommands_Aria2Flags(t *testing.T) {
	h := newTestHandler()

	// 不需要 aria2 的命令
	noAria2 := map[string]bool{
		"/help": true, "/ping": true,
		"/install": true, "/uninstall": false, "/upgrade": false,
		"/msg_config": true, "/msg_autodel": true, "/msg_time": true,
		"/msg_result": true, "/msg_error": true, "/msg_notify": true, "/msg_progress": true,
	}

	for name, expectNoAria2 := range noAria2 {
		cmd, ok := h.commands[name]
		if !ok {
			t.Errorf("命令未找到: %s", name)
			continue
		}
		if expectNoAria2 && cmd.NeedAria2 {
			t.Errorf("命令 %s 不应要求 aria2 已安装", name)
		}
		if !expectNoAria2 && !cmd.NeedAria2 {
			t.Errorf("命令 %s 应要求 aria2 已安装", name)
		}
	}

	// 不需要 aria2 运行的命令（安装管理类 / conf 管理类）
	noRunning := map[string]bool{
		"/help": true, "/ping": true,
		"/install": true, "/uninstall": true, "/upgrade": true,
		"/aria_start": true,
		"/conf_dir": true, "/conf_rpc_port": true, "/conf_secret": true, "/reset_secret": true,
		"/msg_config": true, "/msg_autodel": true, "/msg_time": true,
		"/msg_result": true, "/msg_error": true, "/msg_notify": true, "/msg_progress": true,
	}

	for name, expectNoRunning := range noRunning {
		cmd, ok := h.commands[name]
		if !ok {
			t.Errorf("命令未找到: %s", name)
			continue
		}
		if expectNoRunning && cmd.NeedRunning {
			t.Errorf("命令 %s 不应要求 aria2 正在运行", name)
		}
		if !expectNoRunning && !cmd.NeedRunning {
			t.Errorf("命令 %s 应要求 aria2 正在运行", name)
		}
	}
}

// 验证 30 个命令无重复名称。
func TestRegisterCommands_NoDuplicateNames(t *testing.T) {
	h := newTestHandler()
	names := make(map[string]bool)
	for _, cmd := range h.commands {
		if names[cmd.Name] {
			t.Errorf("命令名称重复: %s", cmd.Name)
		}
		names[cmd.Name] = true
	}
}

// 验证每个命令的 Handler 不为 nil。
func TestRegisterCommands_AllHandlersNonNil(t *testing.T) {
	h := newTestHandler()
	for key, cmd := range h.commands {
		if cmd.Handler == nil {
			t.Errorf("命令 %s 的 Handler 为 nil", key)
		}
	}
}

// ===== parseCommand 测试 =====

func TestParseCommand_Simple(t *testing.T) {
	cmd, args := parseCommand("/help")
	if cmd != "/help" {
		t.Errorf("期望命令 /help, 实际 %s", cmd)
	}
	if len(args) != 0 {
		t.Errorf("期望无参数, 实际 %v", args)
	}
}

func TestParseCommand_WithArgs(t *testing.T) {
	cmd, args := parseCommand("/add http://example.com/file.iso")
	if cmd != "/add" {
		t.Errorf("期望命令 /add, 实际 %s", cmd)
	}
	if len(args) != 1 || args[0] != "http://example.com/file.iso" {
		t.Errorf("期望参数 [http://example.com/file.iso], 实际 %v", args)
	}
}

func TestParseCommand_WithMultipleArgs(t *testing.T) {
	cmd, args := parseCommand("/add url1 url2 url3")
	if cmd != "/add" {
		t.Errorf("期望命令 /add, 实际 %s", cmd)
	}
	if len(args) != 3 {
		t.Errorf("期望 3 个参数, 实际 %d: %v", len(args), args)
	}
}

func TestParseCommand_WithBotSuffix(t *testing.T) {
	cmd, args := parseCommand("/help@MyBot")
	if cmd != "/help" {
		t.Errorf("期望命令 /help, 实际 %s", cmd)
	}
	if len(args) != 0 {
		t.Errorf("期望无参数, 实际 %v", args)
	}
}

func TestParseCommand_NonCommand(t *testing.T) {
	cmd, args := parseCommand("hello world")
	if cmd != "" {
		t.Errorf("期望空命令, 实际 %s", cmd)
	}
	if args != nil {
		t.Errorf("期望 nil 参数, 实际 %v", args)
	}
}

func TestParseCommand_EmptyText(t *testing.T) {
	cmd, args := parseCommand("")
	if cmd != "" {
		t.Errorf("期望空命令, 实际 %s", cmd)
	}
	if args != nil {
		t.Errorf("期望 nil 参数, 实际 %v", args)
	}
}

func TestParseCommand_LimitTaskArgs(t *testing.T) {
	cmd, args := parseCommand("/limit_task 3 5M")
	if cmd != "/limit_task" {
		t.Errorf("期望命令 /limit_task, 实际 %s", cmd)
	}
	if len(args) != 2 || args[0] != "3" || args[1] != "5M" {
		t.Errorf("期望参数 [3 5M], 实际 %v", args)
	}
}

func TestParseCommand_LeadingSpaces(t *testing.T) {
	cmd, args := parseCommand("  /ping  ")
	if cmd != "/ping" {
		t.Errorf("期望命令 /ping, 实际 %s", cmd)
	}
	if len(args) != 0 {
		t.Errorf("期望无参数, 实际 %v", args)
	}
}
