package telegram

import (
	"testing"
	"time"

	"github.com/dnslin/aria2-tgbot/internal/config"
)

// newTestMessageConfig 创建测试用 MessageConfig。
func newTestMessageConfig(autoDelete bool, resultTTL, progressTTL, errorTTL, notifyTTL int, installLogDelete bool) *config.MessageConfig {
	ad := autoDelete
	return &config.MessageConfig{
		AutoDelete:             &ad,
		ResultDeleteAfter:      resultTTL,
		ProgressUpdateInterval: progressTTL,
		ErrorDeleteAfter:       errorTTL,
		NotifyDeleteAfter:      notifyTTL,
		InstallLogDelete:       installLogDelete,
	}
}

// newTestMessageManager 创建测试用 MessageManager（不需要真实 bot）。
func newTestMessageManager(msgCfg *config.MessageConfig) *MessageManager {
	cfg := &config.Config{Message: *msgCfg}
	return &MessageManager{
		bot:    nil,
		cfg:    cfg,
		timers: make(map[int64]map[int]*time.Timer),
	}
}

func TestGetTTL_CommandWithAutoDelete(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelCommand)
	if ttl != 300 {
		t.Errorf("LabelCommand TTL 应为 300, 实际 %d", ttl)
	}
}

func TestGetTTL_CommandAutoDeleteDisabled(t *testing.T) {
	msgCfg := newTestMessageConfig(false, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelCommand)
	if ttl != 0 {
		t.Errorf("AutoDelete 关闭时 LabelCommand TTL 应为 0, 实际 %d", ttl)
	}
}

func TestGetTTL_Progress(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelProgress)
	if ttl != 30 {
		t.Errorf("LabelProgress TTL 应为 30, 实际 %d", ttl)
	}
}

func TestGetTTL_InstallWithLogDelete(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, true)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelInstall)
	if ttl != 300 {
		t.Errorf("InstallLogDelete=true 时 LabelInstall TTL 应为 300, 实际 %d", ttl)
	}
}

func TestGetTTL_InstallLogDeleteDisabled(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelInstall)
	if ttl != 0 {
		t.Errorf("InstallLogDelete=false 时 LabelInstall TTL 应为 0, 实际 %d", ttl)
	}
}

func TestGetTTL_ErrorZero(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelError)
	if ttl != 0 {
		t.Errorf("ErrorDeleteAfter=0 时 LabelError TTL 应为 0, 实际 %d", ttl)
	}
}

func TestGetTTL_ErrorNonZero(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 120, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelError)
	if ttl != 120 {
		t.Errorf("ErrorDeleteAfter=120 时 LabelError TTL 应为 120, 实际 %d", ttl)
	}
}

func TestGetTTL_NotifyZero(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelNotify)
	if ttl != 0 {
		t.Errorf("NotifyDeleteAfter=0 时 LabelNotify TTL 应为 0, 实际 %d", ttl)
	}
}

func TestGetTTL_NotifyNonZero(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 600, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(LabelNotify)
	if ttl != 600 {
		t.Errorf("NotifyDeleteAfter=600 时 LabelNotify TTL 应为 600, 实际 %d", ttl)
	}
}

func TestGetTTL_UnknownLabel(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 300, 30, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	ttl := mgr.getTTL(MessageLabel("unknown"))
	if ttl != 0 {
		t.Errorf("未知 label TTL 应为 0, 实际 %d", ttl)
	}
}

func TestReloadConfig(t *testing.T) {
	msgCfg := newTestMessageConfig(true, 100, 10, 0, 0, false)
	mgr := newTestMessageManager(msgCfg)

	if mgr.getTTL(LabelCommand) != 100 {
		t.Error("重载前 TTL 应为 100")
	}

	newCfg := &config.Config{Message: config.MessageConfig{
		AutoDelete:        &[]bool{true}[0],
		ResultDeleteAfter: 500,
	}}
	mgr.ReloadConfig(newCfg)

	if mgr.getTTL(LabelCommand) != 500 {
		t.Errorf("重载后 TTL 应为 500, 实际 %d", mgr.getTTL(LabelCommand))
	}
}
