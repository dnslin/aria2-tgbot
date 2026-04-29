package telegram

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dnslin/aria2-tgbot/internal/config"
)

// 创建带临时文件的测试用 Config。
func newTestConfigWithFile(t *testing.T, enabled bool, allowUsers []int64) *config.Config {
	t.Helper()
	enabledPtr := enabled

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	// 写入最小配置文件
	data := []byte("bot:\n  token: \"test\"\n")
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		t.Fatalf("写入临时配置文件失败: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("加载临时配置失败: %v", err)
	}
	cfg.Auth.Enabled = &enabledPtr
	cfg.Auth.AllowUsers = allowUsers
	return cfg
}

// 创建内存 Config（不关联文件，用于测试纯逻辑的 Authorize）。
func newTestConfig(enabled bool, allowUsers []int64) *config.Config {
	enabledPtr := enabled
	return &config.Config{
		Auth: config.AuthConfig{
			Enabled:    &enabledPtr,
			AllowUsers: allowUsers,
		},
	}
}

// 权限关闭时，任意用户均可操作。
func TestAuthorize_AuthDisabled_AllowsAnyUser(t *testing.T) {
	cfg := newTestConfig(false, []int64{})
	if !Authorize(cfg, 12345) {
		t.Error("权限关闭时应允许任意用户")
	}
	if !Authorize(cfg, 99999) {
		t.Error("权限关闭时应允许任意用户")
	}
}

// 权限开启时，白名单内的用户允许操作。
func TestAuthorize_AuthEnabled_Whitelisted(t *testing.T) {
	cfg := newTestConfig(true, []int64{100, 200, 300})
	if !Authorize(cfg, 200) {
		t.Error("白名单用户应被允许")
	}
}

// 权限开启时，非白名单用户被拒绝。
func TestAuthorize_AuthEnabled_NotWhitelisted(t *testing.T) {
	cfg := newTestConfig(true, []int64{100, 200})
	if Authorize(cfg, 99999) {
		t.Error("非白名单用户应被拒绝")
	}
}

// 权限开启但白名单为空时，所有用户被拒绝。
func TestAuthorize_AuthEnabled_EmptyWhitelist(t *testing.T) {
	cfg := newTestConfig(true, []int64{})
	if Authorize(cfg, 12345) {
		t.Error("空白名单时所有用户应被拒绝")
	}
}

// 权限未设置（nil），默认启用。
func TestAuthorize_NilEnabled_DefaultsToEnabled(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Enabled:    nil,
			AllowUsers: []int64{100},
		},
	}
	if !Authorize(cfg, 100) {
		t.Error("nil Enabled 默认启用，白名单用户应被允许")
	}
	if Authorize(cfg, 99999) {
		t.Error("nil Enabled 默认启用，非白名单用户应被拒绝")
	}
}

// WhitelistAdd 添加用户到白名单并持久化。
func TestWhitelistAdd_AddsUserAndPersists(t *testing.T) {
	cfg := newTestConfigWithFile(t, true, []int64{100})
	if err := WhitelistAdd(cfg, 200); err != nil {
		t.Fatalf("WhitelistAdd 失败: %v", err)
	}
	if len(cfg.Auth.AllowUsers) != 2 {
		t.Errorf("白名单长度应为 2, 实际 %d", len(cfg.Auth.AllowUsers))
	}
	if !Authorize(cfg, 200) {
		t.Error("添加后用户 200 应有权限")
	}

	// 重新加载验证持久化
	reloaded, err := config.Load(cfg.FilePath())
	if err != nil {
		t.Fatalf("重新加载配置失败: %v", err)
	}
	if len(reloaded.Auth.AllowUsers) != 2 {
		t.Errorf("持久化后白名单长度应为 2, 实际 %d", len(reloaded.Auth.AllowUsers))
	}
}

// WhitelistAdd 幂等：重复添加不创建重复条目。
func TestWhitelistAdd_Idempotent(t *testing.T) {
	cfg := newTestConfigWithFile(t, true, []int64{100})
	if err := WhitelistAdd(cfg, 100); err != nil {
		t.Fatalf("重复添加 WhitelistAdd 失败: %v", err)
	}
	if len(cfg.Auth.AllowUsers) != 1 {
		t.Errorf("幂等添加后白名单长度仍为 1, 实际 %d", len(cfg.Auth.AllowUsers))
	}
}

// WhitelistRemove 移除用户并持久化。
func TestWhitelistRemove_RemovesUserAndPersists(t *testing.T) {
	cfg := newTestConfigWithFile(t, true, []int64{100, 200, 300})
	if err := WhitelistRemove(cfg, 200); err != nil {
		t.Fatalf("WhitelistRemove 失败: %v", err)
	}
	if len(cfg.Auth.AllowUsers) != 2 {
		t.Errorf("移除后白名单长度应为 2, 实际 %d", len(cfg.Auth.AllowUsers))
	}
	if Authorize(cfg, 200) {
		t.Error("移除后用户 200 不应有权限")
	}

	// 重新加载验证持久化
	reloaded, err := config.Load(cfg.FilePath())
	if err != nil {
		t.Fatalf("重新加载配置失败: %v", err)
	}
	if len(reloaded.Auth.AllowUsers) != 2 {
		t.Errorf("持久化后白名单长度应为 2, 实际 %d", len(reloaded.Auth.AllowUsers))
	}
}

// WhitelistRemove 幂等：移除不存在的用户不报错。
func TestWhitelistRemove_NonExistentNoop(t *testing.T) {
	cfg := newTestConfigWithFile(t, true, []int64{100})
	if err := WhitelistRemove(cfg, 999); err != nil {
		t.Fatalf("移除不存在用户应无错误, 实际: %v", err)
	}
	if len(cfg.Auth.AllowUsers) != 1 {
		t.Errorf("移除不存在用户后白名单长度仍为 1, 实际 %d", len(cfg.Auth.AllowUsers))
	}
}

// WhitelistList 返回当前列表。
func TestWhitelistList_ReturnsCurrentList(t *testing.T) {
	cfg := newTestConfig(true, []int64{100, 200, 300})
	list := WhitelistList(cfg)
	if len(list) != 3 {
		t.Errorf("白名单长度应为 3, 实际 %d", len(list))
	}
}

// 边界：权限开启，多个白名单用户
func TestAuthorize_MultipleWhitelistedUsers(t *testing.T) {
	users := []int64{111, 222, 333, 444, 555}
	cfg := newTestConfig(true, users)
	for _, u := range users {
		if !Authorize(cfg, u) {
			t.Errorf("白名单用户 %d 应被允许", u)
		}
	}
	if Authorize(cfg, 666) {
		t.Error("非白名单用户 666 应被拒绝")
	}
}

// 边界：userID = 0
func TestAuthorize_UserIDZero(t *testing.T) {
	cfg := newTestConfig(true, []int64{0})
	if !Authorize(cfg, 0) {
		t.Error("userID 0 在白名单中应被允许")
	}
	cfg2 := newTestConfig(true, []int64{100})
	if Authorize(cfg2, 0) {
		t.Error("userID 0 不在白名单中应被拒绝")
	}
}

// 边界：白名单增删组合操作
func TestWhitelistAddRemove_Sequence(t *testing.T) {
	cfg := newTestConfigWithFile(t, true, []int64{100})

	// 添加多个用户
	for _, uid := range []int64{200, 300, 400} {
		if err := WhitelistAdd(cfg, uid); err != nil {
			t.Fatalf("添加用户 %d 失败: %v", uid, err)
		}
	}
	if len(cfg.Auth.AllowUsers) != 4 {
		t.Errorf("序列添加后应为 4 个用户, 实际 %d", len(cfg.Auth.AllowUsers))
	}

	// 移除中间用户
	if err := WhitelistRemove(cfg, 200); err != nil {
		t.Fatalf("移除用户 200 失败: %v", err)
	}
	if err := WhitelistRemove(cfg, 400); err != nil {
		t.Fatalf("移除用户 400 失败: %v", err)
	}
	if len(cfg.Auth.AllowUsers) != 2 {
		t.Errorf("序列移除后应为 2 个用户, 实际 %d", len(cfg.Auth.AllowUsers))
	}
	if !Authorize(cfg, 100) || !Authorize(cfg, 300) {
		t.Error("剩余用户应有权限")
	}
}
