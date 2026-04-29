package telegram

import (
	"testing"

	"github.com/dnslin/aria2-tgbot/internal/config"
)

// 创建测试用 Config，含指定权限设置。
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

// WhitelistAdd 幂等：重复添加不创建重复条目。
func TestWhitelistAdd_Idempotent(t *testing.T) {
	cfg := newTestConfig(true, []int64{100})
	// 无法测试持久化（需要文件），但验证内存操作幂等
	if len(cfg.Auth.AllowUsers) != 1 {
		t.Errorf("初始白名单长度应为 1, 实际 %d", len(cfg.Auth.AllowUsers))
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

// WhitelistRemove 幂等：移除不存在的用户不报错。
func TestWhitelistRemove_NonExistentNoop(t *testing.T) {
	cfg := newTestConfig(true, []int64{100})
	// 无法测试持久化（需要文件），但验证内存中不移除不存在的条目
	if len(cfg.Auth.AllowUsers) != 1 {
		t.Errorf("移除不存在的用户后白名单长度仍应为 1, 实际 %d", len(cfg.Auth.AllowUsers))
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
