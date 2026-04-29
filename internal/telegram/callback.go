package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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

// HandleCallback 处理 Inline Keyboard 回调查询（占位实现）。
// 后续 Issue #4 完成完整的回调分发逻辑。
func (h *Handler) HandleCallback(query *tgbotapi.CallbackQuery) {
	// 占位：回复空回调确认，避免客户端一直显示 loading
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.bot.Request(callback); err != nil {
		h.logger.Error("回调确认失败: %v", err)
	}
}
