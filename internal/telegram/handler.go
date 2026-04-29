package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/dnslin/aria2-tgbot/internal/config"
	"github.com/dnslin/aria2-tgbot/internal/logger"
)

// Handler Telegram Update 分发处理器。
// 负责命令路由、参数解析、权限校验、gidMap 维护。
type Handler struct {
	svc    any             // service.Container 占位，后续 #5 注入
	cfg    *config.Config  // 全局配置
	bot    *tgbotapi.BotAPI
	logger *logger.Logger

	msgMgr   *MessageManager              // 消息管理器
	commands map[string]*Command           // 命令注册表
	gidMap   map[int64]map[int]string     // chatID -> 编号 -> GID（待 #5 添加访问方法）
}

// NewHandler 创建 Handler 实例。
// svc 参数当前为占位接口，后续 Issue #5 完成后注入真实 service.Container。
func NewHandler(svc any, cfg *config.Config, bot *tgbotapi.BotAPI, log *logger.Logger, msgMgr *MessageManager) *Handler {
	h := &Handler{
		svc:    svc,
		cfg:    cfg,
		bot:    bot,
		logger: log,
		msgMgr: msgMgr,
		gidMap: make(map[int64]map[int]string),
	}
	h.commands = h.RegisterCommands()
	return h
}

// HandleUpdate 处理 Telegram Update 的总入口。
// 根据 Update 类型分发到命令处理或回调处理。
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	// 回调查询
	if update.CallbackQuery != nil {
		h.HandleCallback(update.CallbackQuery)
		return
	}

	// 文本消息
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	msg := update.Message
	userID := msg.From.ID

	// 权限校验
	if !h.authorize(userID) {
		h.logger.Debug("未授权用户 %d 尝试使用 Bot，已忽略", userID)
		return
	}

	// 命令匹配和路由
	text := msg.Text
	cmdName, args := parseCommand(text)

	cmd, ok := h.commands[cmdName]
	if !ok {
		h.logger.Debug("未知命令: %s (来自用户 %d)", cmdName, userID)
		return // 未知命令静默忽略
	}

	// 前置检查：aria2 状态
	if cmd.NeedAria2 || cmd.NeedRunning {
		// TODO: 当 #5 Service 层就绪后，通过 h.svc 检查 aria2 安装和运行状态
		// 当前占位阶段跳过检查，直接执行命令
	}

	h.logger.Debug("执行命令: %s (用户: %d, 参数: %v)", cmdName, userID, args)
	cmd.Handler(msg, args)
}

// authorize 校验用户是否有权限使用 Bot。
func (h *Handler) authorize(userID int64) bool {
	return Authorize(h.cfg, userID)
}

// sendReply 发送纯文本回复消息（占位实现，后续 #4 替换为 MessageManager.Send）。
func (h *Handler) sendReply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error("发送消息失败 (chatID: %d): %v", chatID, err)
	}
}

// parseCommand 从消息文本中解析命令名和参数。
// "/add http://example.com" → "/add", ["http://example.com"]
// "/help" → "/help", []
// "hello" → "", []
func parseCommand(text string) (string, []string) {
	text = strings.TrimSpace(text)
	if text == "" || text[0] != '/' {
		return "", nil
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil
	}

	cmdName := parts[0]
	// 处理如 "/add@BotName" 格式的命令
	if idx := strings.Index(cmdName, "@"); idx != -1 {
		cmdName = cmdName[:idx]
	}

	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	return cmdName, args
}
