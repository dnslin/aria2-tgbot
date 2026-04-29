// Package telegram 提供 Telegram Bot 交互层：命令注册、Update 分发、权限校验、消息管理。
package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Command 描述一个 Bot 命令的元数据和处理函数。
type Command struct {
	Name        string                                           // 命令名称（以 / 开头）
	Description string                                           // 命令描述
	Args        string                                           // 参数说明（用于 /help 展示）
	Handler     func(msg *tgbotapi.Message, args []string)       // 命令处理函数
	NeedAria2   bool                                             // 是否需要 aria2 已安装
	NeedRunning bool                                             // 是否需要 aria2 正在运行
}

// RegisterCommands 注册全部 Bot 命令并返回命令映射表。
// 所有 30 个命令在此注册，Handler 目前使用占位实现，
// 后续 #4、#5 完成后注入真实的 Service 层逻辑。
func (h *Handler) RegisterCommands() map[string]*Command {
	return map[string]*Command{
		// ===== 帮助 & 状态 =====
		"/help": {
			Name:        "help",
			Description: "显示命令列表和使用帮助",
			Args:        "",
			Handler:     h.handleHelp,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/ping": {
			Name:        "ping",
			Description: "Bot 存活检查",
			Args:        "",
			Handler:     h.handlePing,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/health": {
			Name:        "health",
			Description: "综合健康检查（Bot + aria2 连接 + 磁盘）",
			Args:        "",
			Handler:     h.handleHealth,
			NeedAria2:   true,
			NeedRunning: true,
		},

		// ===== 安装管理 =====
		"/install": {
			Name:        "install",
			Description: "安装 aria2（确认弹窗）",
			Args:        "",
			Handler:     h.handleInstall,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/uninstall": {
			Name:        "uninstall",
			Description: "卸载 aria2（确认弹窗）",
			Args:        "",
			Handler:     h.handleUninstall,
			NeedAria2:   true,
			NeedRunning: false,
		},
		"/upgrade": {
			Name:        "upgrade",
			Description: "升级 aria2 到最新版本",
			Args:        "",
			Handler:     h.handleUpgrade,
			NeedAria2:   true,
			NeedRunning: false,
		},

		// ===== 进程管理 =====
		"/aria_start": {
			Name:        "aria_start",
			Description: "启动 aria2 进程",
			Args:        "",
			Handler:     h.handleAriaStart,
			NeedAria2:   true,
			NeedRunning: false,
		},
		"/aria_stop": {
			Name:        "aria_stop",
			Description: "停止 aria2 进程",
			Args:        "",
			Handler:     h.handleAriaStop,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/aria_restart": {
			Name:        "aria_restart",
			Description: "重启 aria2 进程",
			Args:        "",
			Handler:     h.handleAriaRestart,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/aria_status": {
			Name:        "aria_status",
			Description: "查看 aria2 运行状态",
			Args:        "",
			Handler:     h.handleAriaStatus,
			NeedAria2:   true,
			NeedRunning: true,
		},

		// ===== 下载操作 =====
		"/add": {
			Name:        "add",
			Description: "添加 HTTP/FTP 下载",
			Args:        "<url> [url...]",
			Handler:     h.handleAdd,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/add_magnet": {
			Name:        "add_magnet",
			Description: "添加磁力链接下载",
			Args:        "<磁力链接>",
			Handler:     h.handleAddMagnet,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/pause": {
			Name:        "pause",
			Description: "暂停指定任务",
			Args:        "<编号>",
			Handler:     h.handlePause,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/resume": {
			Name:        "resume",
			Description: "恢复指定任务",
			Args:        "<编号>",
			Handler:     h.handleResume,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/remove": {
			Name:        "remove",
			Description: "删除任务（含文件）",
			Args:        "<编号>",
			Handler:     h.handleRemove,
			NeedAria2:   true,
			NeedRunning: true,
		},

		// ===== 列表查询 =====
		"/list": {
			Name:        "list",
			Description: "查看活动任务列表",
			Args:        "",
			Handler:     h.handleList,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/done": {
			Name:        "done",
			Description: "查看已完成任务列表",
			Args:        "",
			Handler:     h.handleDone,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/clear": {
			Name:        "clear",
			Description: "清理所有已完成/失败记录",
			Args:        "",
			Handler:     h.handleClear,
			NeedAria2:   true,
			NeedRunning: true,
		},

		// ===== 速度控制 =====
		"/limit_global": {
			Name:        "limit_global",
			Description: "全局限速",
			Args:        "<速度> （如 5M，0 取消）",
			Handler:     h.handleLimitGlobal,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/limit_task": {
			Name:        "limit_task",
			Description: "单任务限速",
			Args:        "<编号> <速度>",
			Handler:     h.handleLimitTask,
			NeedAria2:   true,
			NeedRunning: true,
		},

		// ===== Aria2 配置 =====
		"/conf": {
			Name:        "conf",
			Description: "查看当前 aria2 核心配置",
			Args:        "",
			Handler:     h.handleConf,
			NeedAria2:   true,
			NeedRunning: true,
		},
		"/conf_dir": {
			Name:        "conf_dir",
			Description: "修改下载目录",
			Args:        "<路径>",
			Handler:     h.handleConfDir,
			NeedAria2:   true,
			NeedRunning: false,
		},
		"/conf_rpc_port": {
			Name:        "conf_rpc_port",
			Description: "修改 RPC 端口（需重启生效）",
			Args:        "<端口>",
			Handler:     h.handleConfRpcPort,
			NeedAria2:   true,
			NeedRunning: false,
		},
		"/conf_secret": {
			Name:        "conf_secret",
			Description: "查看当前 RPC 密钥",
			Args:        "",
			Handler:     h.handleConfSecret,
			NeedAria2:   true,
			NeedRunning: false,
		},
		"/reset_secret": {
			Name:        "reset_secret",
			Description: "重置 RPC 密钥（确认弹窗）",
			Args:        "",
			Handler:     h.handleResetSecret,
			NeedAria2:   true,
			NeedRunning: false,
		},

		// ===== 统计 =====
		"/stats": {
			Name:        "stats",
			Description: "全局下载统计",
			Args:        "",
			Handler:     h.handleStats,
			NeedAria2:   true,
			NeedRunning: true,
		},

		// ===== 消息行为 =====
		"/msg_config": {
			Name:        "msg_config",
			Description: "查看当前消息行为配置",
			Args:        "",
			Handler:     h.handleMsgConfig,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/msg_autodel": {
			Name:        "msg_autodel",
			Description: "开启/关闭全局自动删除",
			Args:        "<on/off>",
			Handler:     h.handleMsgAutodel,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/msg_time": {
			Name:        "msg_time",
			Description: "设置默认删除时间（秒，0=不删）",
			Args:        "<秒>",
			Handler:     h.handleMsgTime,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/msg_result": {
			Name:        "msg_result",
			Description: "命令结果保留时间",
			Args:        "<秒>",
			Handler:     h.handleMsgResult,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/msg_error": {
			Name:        "msg_error",
			Description: "错误消息保留时间",
			Args:        "<秒>",
			Handler:     h.handleMsgError,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/msg_notify": {
			Name:        "msg_notify",
			Description: "通知消息保留时间",
			Args:        "<秒>",
			Handler:     h.handleMsgNotify,
			NeedAria2:   false,
			NeedRunning: false,
		},
		"/msg_progress": {
			Name:        "msg_progress",
			Description: "进度刷新间隔",
			Args:        "<秒>",
			Handler:     h.handleMsgProgress,
			NeedAria2:   false,
			NeedRunning: false,
		},
	}
}

// ===== 帮助 & 状态命令处理（占位） =====

func (h *Handler) handleHelp(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handlePing(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "pong!")
}

func (h *Handler) handleHealth(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 安装管理命令处理（占位） =====

func (h *Handler) handleInstall(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleUninstall(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleUpgrade(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 进程管理命令处理（占位） =====

func (h *Handler) handleAriaStart(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleAriaStop(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleAriaRestart(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleAriaStatus(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 下载操作命令处理（占位） =====

func (h *Handler) handleAdd(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleAddMagnet(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handlePause(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleResume(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleRemove(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 列表查询命令处理（占位） =====

func (h *Handler) handleList(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleDone(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleClear(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 速度控制命令处理（占位） =====

func (h *Handler) handleLimitGlobal(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleLimitTask(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== Aria2 配置命令处理（占位） =====

func (h *Handler) handleConf(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleConfDir(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleConfRpcPort(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleConfSecret(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleResetSecret(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 统计命令处理（占位） =====

func (h *Handler) handleStats(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

// ===== 消息行为命令处理（占位） =====

func (h *Handler) handleMsgConfig(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleMsgAutodel(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleMsgTime(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleMsgResult(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleMsgError(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleMsgNotify(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}

func (h *Handler) handleMsgProgress(msg *tgbotapi.Message, args []string) {
	h.sendReply(msg.Chat.ID, "该功能将在后续版本实现")
}
