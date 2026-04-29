package telegram

import (
	"testing"
)

// ===== BuildConfirmKeyboard 测试 =====

func TestBuildConfirmKeyboard_ButtonsAndData(t *testing.T) {
	kb := BuildConfirmKeyboard("install")
	if len(kb.InlineKeyboard) != 1 {
		t.Fatalf("期望 1 行, 实际 %d 行", len(kb.InlineKeyboard))
	}
	row := kb.InlineKeyboard[0]
	if len(row) != 2 {
		t.Fatalf("期望 2 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "确认" {
		t.Errorf("按钮0 文本应为 '确认', 实际 %q", row[0].Text)
	}
	if *row[0].CallbackData != "confirm:install" {
		t.Errorf("按钮0 data 应为 'confirm:install', 实际 %q", *row[0].CallbackData)
	}
	if row[1].Text != "取消" {
		t.Errorf("按钮1 文本应为 '取消', 实际 %q", row[1].Text)
	}
	if *row[1].CallbackData != "cancel:install" {
		t.Errorf("按钮1 data 应为 'cancel:install', 实际 %q", *row[1].CallbackData)
	}
}

// ===== BuildTaskKeyboard 测试 =====

func TestBuildTaskKeyboard_Running(t *testing.T) {
	kb := BuildTaskKeyboard("abc123", false)
	row := kb.InlineKeyboard[0]
	if len(row) != 2 {
		t.Fatalf("期望 2 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "暂停" {
		t.Errorf("运行中任务应显示 '暂停', 实际 %q", row[0].Text)
	}
	if *row[0].CallbackData != "pause:abc123" {
		t.Errorf("data 应为 'pause:abc123', 实际 %q", *row[0].CallbackData)
	}
	if row[1].Text != "删除" {
		t.Errorf("按钮1 文本应为 '删除', 实际 %q", row[1].Text)
	}
	if *row[1].CallbackData != "remove:abc123" {
		t.Errorf("data 应为 'remove:abc123', 实际 %q", *row[1].CallbackData)
	}
}

func TestBuildTaskKeyboard_Paused(t *testing.T) {
	kb := BuildTaskKeyboard("xyz789", true)
	row := kb.InlineKeyboard[0]
	if len(row) != 2 {
		t.Fatalf("期望 2 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "恢复" {
		t.Errorf("暂停任务应显示 '恢复', 实际 %q", row[0].Text)
	}
	if *row[0].CallbackData != "resume:xyz789" {
		t.Errorf("data 应为 'resume:xyz789', 实际 %q", *row[0].CallbackData)
	}
}

// ===== BuildDoneTaskKeyboard 测试 =====

func TestBuildDoneTaskKeyboard(t *testing.T) {
	kb := BuildDoneTaskKeyboard("done123")
	row := kb.InlineKeyboard[0]
	if len(row) != 1 {
		t.Fatalf("期望 1 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "删除记录" {
		t.Errorf("文本应为 '删除记录', 实际 %q", row[0].Text)
	}
	if *row[0].CallbackData != "remove_done:done123" {
		t.Errorf("data 应为 'remove_done:done123', 实际 %q", *row[0].CallbackData)
	}
}

// ===== BuildPaginationKeyboard 测试 =====

func TestBuildPaginationKeyboard_FirstPage(t *testing.T) {
	kb := BuildPaginationKeyboard("active", 1, 5)
	row := kb.InlineKeyboard[0]
	// 第一页：无[上一页]，有[下一页][刷新]
	if len(row) != 2 {
		t.Fatalf("第一页期望 2 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "下一页" {
		t.Errorf("按钮0 应为 '下一页', 实际 %q", row[0].Text)
	}
	if *row[0].CallbackData != "page:active:2" {
		t.Errorf("data 应为 'page:active:2', 实际 %q", *row[0].CallbackData)
	}
	if row[1].Text != "刷新" {
		t.Errorf("按钮1 应为 '刷新', 实际 %q", row[1].Text)
	}
}

func TestBuildPaginationKeyboard_LastPage(t *testing.T) {
	kb := BuildPaginationKeyboard("active", 5, 5)
	row := kb.InlineKeyboard[0]
	// 最后一页：有[上一页]，无[下一页]
	if len(row) != 2 {
		t.Fatalf("最后一页期望 2 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "上一页" {
		t.Errorf("按钮0 应为 '上一页', 实际 %q", row[0].Text)
	}
	if *row[0].CallbackData != "page:active:4" {
		t.Errorf("data 应为 'page:active:4', 实际 %q", *row[0].CallbackData)
	}
}

func TestBuildPaginationKeyboard_MiddlePage(t *testing.T) {
	kb := BuildPaginationKeyboard("done", 3, 5)
	row := kb.InlineKeyboard[0]
	// 中间页：三个按钮齐全
	if len(row) != 3 {
		t.Fatalf("中间页期望 3 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "上一页" {
		t.Errorf("按钮0 应为 '上一页', 实际 %q", row[0].Text)
	}
	if row[1].Text != "下一页" {
		t.Errorf("按钮1 应为 '下一页', 实际 %q", row[1].Text)
	}
	if row[2].Text != "刷新" {
		t.Errorf("按钮2 应为 '刷新', 实际 %q", row[2].Text)
	}
}

func TestBuildPaginationKeyboard_SinglePage(t *testing.T) {
	kb := BuildPaginationKeyboard("active", 1, 1)
	row := kb.InlineKeyboard[0]
	// total=1: 仅[刷新]
	if len(row) != 1 {
		t.Fatalf("单页期望 1 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "刷新" {
		t.Errorf("按钮应为 '刷新', 实际 %q", row[0].Text)
	}
}

func TestBuildPaginationKeyboard_ZeroTotal(t *testing.T) {
	kb := BuildPaginationKeyboard("active", 0, 0)
	row := kb.InlineKeyboard[0]
	// total=0: 仅[刷新]
	if len(row) != 1 {
		t.Fatalf("零条目期望 1 个按钮, 实际 %d", len(row))
	}
	if row[0].Text != "刷新" {
		t.Errorf("按钮应为 '刷新', 实际 %q", row[0].Text)
	}
}

func TestBuildPaginationKeyboard_DifferentListTypes(t *testing.T) {
	// 验证 listType 正确传入 callback data
	kb := BuildPaginationKeyboard("done", 2, 5)
	row := kb.InlineKeyboard[0]
	// 第2页共5页: [上一页][下一页][刷新]
	if *row[0].CallbackData != "page:done:1" {
		t.Errorf("上一页 data 应为 'page:done:1', 实际 %q", *row[0].CallbackData)
	}
	if *row[1].CallbackData != "page:done:3" {
		t.Errorf("下一页 data 应为 'page:done:3', 实际 %q", *row[1].CallbackData)
	}
	if *row[2].CallbackData != "refresh:done" {
		t.Errorf("刷新 data 应为 'refresh:done', 实际 %q", *row[2].CallbackData)
	}
}

// ===== BuildMainKeyboard 测试 =====

func TestBuildMainKeyboard(t *testing.T) {
	kb := BuildMainKeyboard()
	if len(kb.Keyboard) != 2 {
		t.Fatalf("期望 2 行, 实际 %d", len(kb.Keyboard))
	}
	if len(kb.Keyboard[0]) != 2 || len(kb.Keyboard[1]) != 2 {
		t.Fatal("每行应有 2 个按钮")
	}
	// 验证按钮文本
	texts := []string{
		kb.Keyboard[0][0].Text, kb.Keyboard[0][1].Text,
		kb.Keyboard[1][0].Text, kb.Keyboard[1][1].Text,
	}
	expected := []string{"列表", "状态", "统计", "帮助"}
	for i, txt := range texts {
		if txt != expected[i] {
			t.Errorf("按钮 %d: 期望 %q, 实际 %q", i, expected[i], txt)
		}
	}
	if !kb.ResizeKeyboard {
		t.Error("ResizeKeyboard 应为 true")
	}
}

// ===== BuildPreInstallKeyboard 测试 =====

func TestBuildPreInstallKeyboard(t *testing.T) {
	kb := BuildPreInstallKeyboard()
	if len(kb.Keyboard) != 1 {
		t.Fatalf("期望 1 行, 实际 %d", len(kb.Keyboard))
	}
	if len(kb.Keyboard[0]) != 2 {
		t.Fatalf("期望 2 个按钮, 实际 %d", len(kb.Keyboard[0]))
	}
	if kb.Keyboard[0][0].Text != "安装 Aria2" {
		t.Errorf("按钮0 应为 '安装 Aria2', 实际 %q", kb.Keyboard[0][0].Text)
	}
	if kb.Keyboard[0][1].Text != "帮助" {
		t.Errorf("按钮1 应为 '帮助', 实际 %q", kb.Keyboard[0][1].Text)
	}
	if !kb.ResizeKeyboard {
		t.Error("ResizeKeyboard 应为 true")
	}
}

// ===== itoa 辅助函数测试 =====

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{10, "10"},
		{100, "100"},
		{-50, "-50"},
		{999, "999"},
	}
	for _, tc := range tests {
		got := itoa(tc.n)
		if got != tc.want {
			t.Errorf("itoa(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// ===== Benchmark =====

func BenchmarkBuildConfirmKeyboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildConfirmKeyboard("install")
	}
}

func BenchmarkBuildPaginationKeyboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildPaginationKeyboard("active", 5, 20)
	}
}

func BenchmarkBuildTaskKeyboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildTaskKeyboard("abc123def456", false)
	}
}
