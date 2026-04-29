package telegram

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 回调数据前缀常量。
const (
	CbConfirm    = "confirm"     // confirm:install / confirm:uninstall / confirm:reset_secret
	CbCancel     = "cancel"      // cancel:install / cancel:uninstall / cancel:reset_secret
	CbPause      = "pause"       // pause:<gid>
	CbResume     = "resume"      // resume:<gid>
	CbRemove     = "remove"      // remove:<gid>
	CbRemoveDone = "remove_done" // remove_done:<gid>
	CbPage       = "page"        // page:active:2 / page:done:2
	CbRefresh    = "refresh"     // refresh:active / refresh:done
)

// 确认/取消操作的合法 action 白名单。
var validActions = map[string]bool{
	"install":      true,
	"uninstall":    true,
	"reset_secret": true,
}

// HandleCallback 处理 Inline Keyboard 回调查询。
func (h *Handler) HandleCallback(query *tgbotapi.CallbackQuery) {
	// 立即回复空回调确认，消除客户端 loading
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.bot.Request(callback); err != nil {
		h.logger.Error("回调确认失败: %v", err)
	}

	prefix, rest := parseCallbackData(query.Data)
	if prefix == "" && rest == "" {
		return // 空 data 或解析失败
	}

	switch prefix {
	case CbConfirm:
		h.handleConfirmCallback(query, rest)
	case CbCancel:
		h.handleCancelCallback(query, rest)
	case CbPause:
		h.handlePauseCallback(query, rest)
	case CbResume:
		h.handleResumeCallback(query, rest)
	case CbRemove:
		h.handleRemoveCallback(query, rest)
	case CbRemoveDone:
		h.handleRemoveDoneCallback(query, rest)
	case CbPage:
		h.handlePageCallback(query, rest)
	case CbRefresh:
		h.handleRefreshCallback(query, rest)
	default:
		h.logger.Debug("未知回调前缀: %s", prefix)
	}
}

// handleConfirmCallback 处理确认回调。rest = "<action>"
func (h *Handler) handleConfirmCallback(query *tgbotapi.CallbackQuery, action string) {
	if !validActions[action] {
		h.logger.Debug("无效确认 action: %s", action)
		return
	}
	chatID := query.Message.Chat.ID
	h.msgMgr.Send(chatID, "确认操作: "+action+" — 该功能将在后续版本实现", LabelCommand)
}

// handleCancelCallback 处理取消回调。rest = "<action>"
func (h *Handler) handleCancelCallback(query *tgbotapi.CallbackQuery, action string) {
	if !validActions[action] {
		h.logger.Debug("无效取消 action: %s", action)
		return
	}
	chatID := query.Message.Chat.ID
	h.msgMgr.Send(chatID, "已取消操作: "+action, LabelCommand)
}

// handlePauseCallback 处理暂停回调。rest = "<gid>"
func (h *Handler) handlePauseCallback(query *tgbotapi.CallbackQuery, gid string) {
	if gid == "" {
		return
	}
	chatID := query.Message.Chat.ID
	h.logger.Debug("暂停任务: gid=%s, chatID=%d", gid, chatID)
	h.msgMgr.Send(chatID, "该功能将在后续版本实现", LabelCommand)
}

// handleResumeCallback 处理恢复回调。rest = "<gid>"
func (h *Handler) handleResumeCallback(query *tgbotapi.CallbackQuery, gid string) {
	if gid == "" {
		return
	}
	chatID := query.Message.Chat.ID
	h.logger.Debug("恢复任务: gid=%s, chatID=%d", gid, chatID)
	h.msgMgr.Send(chatID, "该功能将在后续版本实现", LabelCommand)
}

// handleRemoveCallback 处理删除回调。rest = "<gid>"
func (h *Handler) handleRemoveCallback(query *tgbotapi.CallbackQuery, gid string) {
	if gid == "" {
		return
	}
	chatID := query.Message.Chat.ID
	h.logger.Debug("删除任务: gid=%s, chatID=%d", gid, chatID)
	h.msgMgr.Send(chatID, "该功能将在后续版本实现", LabelCommand)
}

// handleRemoveDoneCallback 处理已完成记录删除回调。rest = "<gid>"
func (h *Handler) handleRemoveDoneCallback(query *tgbotapi.CallbackQuery, gid string) {
	if gid == "" {
		return
	}
	chatID := query.Message.Chat.ID
	h.logger.Debug("删除记录: gid=%s, chatID=%d", gid, chatID)
	h.msgMgr.Send(chatID, "该功能将在后续版本实现", LabelCommand)
}

// handlePageCallback 处理分页回调。rest = "<type>:<page>"
func (h *Handler) handlePageCallback(query *tgbotapi.CallbackQuery, rest string) {
	parts := strings.SplitN(rest, ":", 2)
	listType := parts[0]
	page := 1
	if len(parts) > 1 {
		if n, err := strconv.Atoi(parts[1]); err == nil {
			page = n
		}
	}
	chatID := query.Message.Chat.ID
	h.logger.Debug("翻页: type=%s, page=%d, chatID=%d", listType, page, chatID)
	h.msgMgr.Send(chatID, "该功能将在后续版本实现", LabelCommand)
}

// handleRefreshCallback 处理刷新回调。rest = "<type>"
func (h *Handler) handleRefreshCallback(query *tgbotapi.CallbackQuery, listType string) {
	chatID := query.Message.Chat.ID
	h.logger.Debug("刷新列表: type=%s, chatID=%d", listType, chatID)
	h.msgMgr.Send(chatID, "该功能将在后续版本实现", LabelCommand)
}

// parseCallbackData 解析回调数据字符串，返回前缀和剩余参数。
// "pause:abc123" → "pause", "abc123"
// "page:active:2" → "page", "active:2"
// "" → "", ""
// "invalid" → "invalid", ""
func parseCallbackData(data string) (prefix, rest string) {
	if data == "" {
		return "", ""
	}
	parts := strings.SplitN(data, ":", 2)
	prefix = parts[0]
	if len(parts) > 1 {
		rest = parts[1]
	}
	return
}
