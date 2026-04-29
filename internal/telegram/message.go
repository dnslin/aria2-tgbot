package telegram

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/dnslin/aria2-tgbot/internal/config"
)

// MessageLabel 消息类型标签，用于 getTTL 查表获取对应的自动删除时间。
type MessageLabel string

const (
	LabelCommand  MessageLabel = "command"
	LabelProgress MessageLabel = "progress"
	LabelInstall  MessageLabel = "install"
	LabelError    MessageLabel = "error"
	LabelNotify   MessageLabel = "notify"
)

// sendConfig 内部配置结构体，由 SendOption 函数选项填充。
type sendConfig struct {
	inlineKeyboard *tgbotapi.InlineKeyboardMarkup
	replyKeyboard  interface{} // tgbotapi.ReplyKeyboardMarkup 或 ReplyKeyboardRemove
}

// SendOption 消息发送函数选项。
type SendOption func(*sendConfig)

// WithInlineKeyboard 为消息附加 Inline 键盘。
func WithInlineKeyboard(kb tgbotapi.InlineKeyboardMarkup) SendOption {
	return func(sc *sendConfig) {
		sc.inlineKeyboard = &kb
	}
}

// WithReplyKeyboard 为消息附加 Reply 键盘。
func WithReplyKeyboard(kb tgbotapi.ReplyKeyboardMarkup) SendOption {
	return func(sc *sendConfig) {
		sc.replyKeyboard = kb
	}
}

// MessageManager 管理 Bot 消息的发送、编辑、删除和自动删除定时器。
type MessageManager struct {
	bot    *tgbotapi.BotAPI
	cfg    *config.Config
	mu     sync.RWMutex
	timers map[int64]map[int]*time.Timer // chatID -> msgID -> timer
}

// NewMessageManager 创建 MessageManager 实例。
func NewMessageManager(bot *tgbotapi.BotAPI, cfg *config.Config) *MessageManager {
	return &MessageManager{
		bot:    bot,
		cfg:    cfg,
		timers: make(map[int64]map[int]*time.Timer),
	}
}

// getTTL 根据消息标签和当前配置返回自动删除的 TTL（秒）。0 表示不自动删除。
func (m *MessageManager) getTTL(label MessageLabel) int {
	switch label {
	case LabelCommand:
		if !m.cfg.Message.IsAutoDeleteEnabled() {
			return 0
		}
		return m.cfg.Message.ResultDeleteAfter
	case LabelProgress:
		return m.cfg.Message.ProgressUpdateInterval
	case LabelInstall:
		if !m.cfg.Message.InstallLogDelete {
			return 0
		}
		return m.cfg.Message.ResultDeleteAfter
	case LabelError:
		return m.cfg.Message.ErrorDeleteAfter
	case LabelNotify:
		return m.cfg.Message.NotifyDeleteAfter
	default:
		return 0
	}
}

// setTimer 为指定消息创建自动删除定时器。
func (m *MessageManager) setTimer(chatID int64, msgID int, ttl int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.timers[chatID] == nil {
		m.timers[chatID] = make(map[int]*time.Timer)
	}

	// 取消旧定时器
	if old, ok := m.timers[chatID][msgID]; ok {
		old.Stop()
	}

	m.timers[chatID][msgID] = time.AfterFunc(time.Duration(ttl)*time.Second, func() {
		m.deleteAndCleanup(chatID, msgID)
	})
}

// deleteAndCleanup 删除消息并清理定时器记录。
func (m *MessageManager) deleteAndCleanup(chatID int64, msgID int) {
	// 请求 Telegram 删除消息（忽略错误，消息可能已被手动删除）
	_, err := m.bot.Request(tgbotapi.NewDeleteMessage(chatID, msgID))
	if err != nil {
		// 静默忽略：消息可能已被手动删除或不存在
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.timers[chatID] != nil {
		delete(m.timers[chatID], msgID)
		if len(m.timers[chatID]) == 0 {
			delete(m.timers, chatID)
		}
	}
}

// stopTimer 停止并清理指定消息的定时器。
func (m *MessageManager) stopTimer(chatID int64, msgID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.timers[chatID] != nil {
		if timer, ok := m.timers[chatID][msgID]; ok {
			timer.Stop()
			delete(m.timers[chatID], msgID)
			if len(m.timers[chatID]) == 0 {
				delete(m.timers, chatID)
			}
		}
	}
}

// ReloadConfig 更新内部配置引用（不清理已有定时器）。
func (m *MessageManager) ReloadConfig(cfg *config.Config) {
	m.cfg = cfg
}

// Send 发送文本消息并根据 label 设置自动删除定时器。
// 支持 WithInlineKeyboard 和 WithReplyKeyboard 函数选项。
func (m *MessageManager) Send(chatID int64, text string, label MessageLabel, opts ...SendOption) (int, error) {
	sc := &sendConfig{}
	for _, opt := range opts {
		opt(sc)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if sc.inlineKeyboard != nil {
		msg.ReplyMarkup = *sc.inlineKeyboard
	}
	if sc.replyKeyboard != nil {
		msg.ReplyMarkup = sc.replyKeyboard
	}

	sent, err := m.bot.Send(msg)
	if err != nil {
		return 0, err
	}

	ttl := m.getTTL(label)
	if ttl > 0 {
		m.setTimer(chatID, sent.MessageID, ttl)
	}

	return sent.MessageID, nil
}

// Edit 编辑已有消息的文本内容，同时重置自动删除定时器。
func (m *MessageManager) Edit(chatID int64, msgID int, text string, label MessageLabel) error {
	edit := tgbotapi.NewEditMessageText(chatID, msgID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown

	if _, err := m.bot.Send(edit); err != nil {
		return err
	}

	// 重置自动删除定时器
	ttl := m.getTTL(label)
	if ttl > 0 {
		m.setTimer(chatID, msgID, ttl)
	} else {
		m.stopTimer(chatID, msgID)
	}

	return nil
}

// Delete 删除消息并清理关联的定时器。
func (m *MessageManager) Delete(chatID int64, msgID int) error {
	_, err := m.bot.Request(tgbotapi.NewDeleteMessage(chatID, msgID))
	if err != nil {
		return err
	}

	m.stopTimer(chatID, msgID)
	return nil
}
