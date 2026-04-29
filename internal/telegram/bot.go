package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/dnslin/aria2-tgbot/internal/config"
	"github.com/dnslin/aria2-tgbot/internal/logger"
)

// Bot Telegram Bot 实例，管理 API 客户端、事件循环和生命周期。
type Bot struct {
	api    *tgbotapi.BotAPI
	cfg    *config.Config
	handler *Handler
	logger  *logger.Logger
	stopCh chan struct{} // 关闭信号通道
}

// New 创建 Bot 实例，初始化 Telegram API 客户端和命令处理器。
// svc 参数当前为占位接口，后续 Issue #5 完成后注入真实 service.Container。
func New(svc any, cfg *config.Config, log *logger.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		return nil, err
	}

	api.Debug = cfg.Bot.Debug

	bot := &Bot{
		api:    api,
		cfg:    cfg,
		logger: log,
		stopCh: make(chan struct{}),
	}

	bot.handler = NewHandler(svc, cfg, api, log)
	return bot, nil
}

// Run 启动 Bot 事件循环，获取 Update 并分发给 Handler 处理。
// 此方法会阻塞直到 Stop() 被调用。
func (b *Bot) Run() {
	b.logger.Info("Bot 已授权为 @%s", b.api.Self.UserName)

	// 注册 Telegram 命令菜单
	b.SetupCommands()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				b.logger.Warn("Update 通道已关闭，Bot 退出")
				return
			}
			b.handler.HandleUpdate(update)

		case <-b.stopCh:
			b.logger.Info("Bot 收到停止信号，正在退出事件循环")
			return
		}
	}
}

// Stop 优雅关闭 Bot，停止 Update 轮询。
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
	close(b.stopCh)
}

// SetupCommands 向 Telegram 注册 Bot 命令菜单（显示在输入框上方）。
func (b *Bot) SetupCommands() {
	cfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "help", Description: "显示命令列表和使用帮助"},
		tgbotapi.BotCommand{Command: "ping", Description: "Bot 存活检查"},
		tgbotapi.BotCommand{Command: "health", Description: "综合健康检查"},
		tgbotapi.BotCommand{Command: "install", Description: "安装 aria2"},
		tgbotapi.BotCommand{Command: "uninstall", Description: "卸载 aria2"},
		tgbotapi.BotCommand{Command: "upgrade", Description: "升级 aria2"},
		tgbotapi.BotCommand{Command: "aria_start", Description: "启动 aria2 进程"},
		tgbotapi.BotCommand{Command: "aria_stop", Description: "停止 aria2 进程"},
		tgbotapi.BotCommand{Command: "aria_restart", Description: "重启 aria2 进程"},
		tgbotapi.BotCommand{Command: "aria_status", Description: "查看 aria2 运行状态"},
		tgbotapi.BotCommand{Command: "add", Description: "添加 HTTP/FTP 下载"},
		tgbotapi.BotCommand{Command: "add_magnet", Description: "添加磁力链接下载"},
		tgbotapi.BotCommand{Command: "pause", Description: "暂停指定任务"},
		tgbotapi.BotCommand{Command: "resume", Description: "恢复指定任务"},
		tgbotapi.BotCommand{Command: "remove", Description: "删除任务（含文件）"},
		tgbotapi.BotCommand{Command: "list", Description: "查看活动任务列表"},
		tgbotapi.BotCommand{Command: "done", Description: "查看已完成任务列表"},
		tgbotapi.BotCommand{Command: "clear", Description: "清理已完成/失败记录"},
		tgbotapi.BotCommand{Command: "limit_global", Description: "全局限速"},
		tgbotapi.BotCommand{Command: "limit_task", Description: "单任务限速"},
		tgbotapi.BotCommand{Command: "conf", Description: "查看 aria2 配置"},
		tgbotapi.BotCommand{Command: "conf_dir", Description: "修改下载目录"},
		tgbotapi.BotCommand{Command: "conf_rpc_port", Description: "修改 RPC 端口"},
		tgbotapi.BotCommand{Command: "conf_secret", Description: "查看 RPC 密钥"},
		tgbotapi.BotCommand{Command: "reset_secret", Description: "重置 RPC 密钥"},
		tgbotapi.BotCommand{Command: "stats", Description: "全局下载统计"},
		tgbotapi.BotCommand{Command: "msg_config", Description: "查看消息行为配置"},
		tgbotapi.BotCommand{Command: "msg_autodel", Description: "开关自动删除"},
		tgbotapi.BotCommand{Command: "msg_time", Description: "设置删除时间"},
		tgbotapi.BotCommand{Command: "msg_result", Description: "命令结果保留时间"},
		tgbotapi.BotCommand{Command: "msg_error", Description: "错误消息保留时间"},
		tgbotapi.BotCommand{Command: "msg_notify", Description: "通知保留时间"},
		tgbotapi.BotCommand{Command: "msg_progress", Description: "进度刷新间隔"},
	)

	if _, err := b.api.Request(cfg); err != nil {
		b.logger.Error("注册命令菜单失败: %v", err)
	} else {
		b.logger.Info("命令菜单已注册 (%d 个命令)", 33)
	}
}
