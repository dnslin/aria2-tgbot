package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// BuildConfirmKeyboard 构建确认/取消 Inline 键盘。
// action 取值: "install", "uninstall", "reset_secret"
func BuildConfirmKeyboard(action string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("确认", CbConfirm+":"+action),
			tgbotapi.NewInlineKeyboardButtonData("取消", CbCancel+":"+action),
		),
	)
}

// BuildTaskKeyboard 构建活动任务操作 Inline 键盘。
// isPaused=true 时显示[恢复]，否则显示[暂停]；始终显示[删除]。
func BuildTaskKeyboard(gid string, isPaused bool) tgbotapi.InlineKeyboardMarkup {
	var actionBtn tgbotapi.InlineKeyboardButton
	if isPaused {
		actionBtn = tgbotapi.NewInlineKeyboardButtonData("恢复", CbResume+":"+gid)
	} else {
		actionBtn = tgbotapi.NewInlineKeyboardButtonData("暂停", CbPause+":"+gid)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			actionBtn,
			tgbotapi.NewInlineKeyboardButtonData("删除", CbRemove+":"+gid),
		),
	)
}

// BuildDoneTaskKeyboard 构建已完成任务操作 Inline 键盘（仅删除记录按钮）。
func BuildDoneTaskKeyboard(gid string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("删除记录", CbRemoveDone+":"+gid),
		),
	)
}

// BuildPaginationKeyboard 构建分页导航 Inline 键盘。
// page=1 时隐藏[上一页]，page=total 时隐藏[下一页]，total<=1 时仅显示[刷新]。
func BuildPaginationKeyboard(listType string, page, total int) tgbotapi.InlineKeyboardMarkup {
	var row []tgbotapi.InlineKeyboardButton

	if page > 1 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			"上一页", CbPage+":"+listType+":"+itoa(page-1),
		))
	}
	if page < total {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			"下一页", CbPage+":"+listType+":"+itoa(page+1),
		))
	}
	row = append(row, tgbotapi.NewInlineKeyboardButtonData(
		"刷新", CbRefresh+":"+listType,
	))

	return tgbotapi.NewInlineKeyboardMarkup(row)
}

// itoa 是 fmt.Sprintf("%d", n) 的轻量替代，避免导入 fmt。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// BuildMainKeyboard 构建默认 Reply 键盘（已安装 aria2 时使用）。
func BuildMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("列表"),
			tgbotapi.NewKeyboardButton("状态"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("统计"),
			tgbotapi.NewKeyboardButton("帮助"),
		),
	)
}

// BuildPreInstallKeyboard 构建预安装 Reply 键盘（未安装 aria2 时使用）。
func BuildPreInstallKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("安装 Aria2"),
			tgbotapi.NewKeyboardButton("帮助"),
		),
	)
}
