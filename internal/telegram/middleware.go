package telegram

import (
	"slices"

	"github.com/dnslin/aria2-tgbot/internal/config"
)

// Authorize 校验用户是否有权限使用 Bot。
// 权限关闭时所有用户可操作，开启时仅白名单用户可操作。
func Authorize(cfg *config.Config, userID int64) bool {
	if !cfg.Auth.IsEnabled() {
		return true
	}
	return slices.Contains(cfg.Auth.AllowUsers, userID)
}

// WhitelistAdd 添加用户到白名单并持久化到配置文件。
func WhitelistAdd(cfg *config.Config, userID int64) error {
	if slices.Contains(cfg.Auth.AllowUsers, userID) {
		return nil // 已存在，幂等
	}
	cfg.Auth.AllowUsers = append(cfg.Auth.AllowUsers, userID)
	return cfg.Save(cfg.FilePath())
}

// WhitelistRemove 从白名单移除用户并持久化到配置文件。
func WhitelistRemove(cfg *config.Config, userID int64) error {
	idx := slices.Index(cfg.Auth.AllowUsers, userID)
	if idx == -1 {
		return nil // 不存在，幂等
	}
	cfg.Auth.AllowUsers = slices.Delete(cfg.Auth.AllowUsers, idx, idx+1)
	return cfg.Save(cfg.FilePath())
}

// WhitelistList 返回当前白名单用户 ID 列表。
func WhitelistList(cfg *config.Config) []int64 {
	return cfg.Auth.AllowUsers
}
