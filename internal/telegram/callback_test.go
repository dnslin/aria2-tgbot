package telegram

import (
	"testing"
)

func TestParseCallbackData_Pause(t *testing.T) {
	prefix, rest := parseCallbackData("pause:abc123")
	if prefix != "pause" {
		t.Errorf("prefix 应为 'pause', 实际 %q", prefix)
	}
	if rest != "abc123" {
		t.Errorf("rest 应为 'abc123', 实际 %q", rest)
	}
}

func TestParseCallbackData_Resume(t *testing.T) {
	prefix, rest := parseCallbackData("resume:xyz789")
	if prefix != "resume" {
		t.Errorf("prefix 应为 'resume', 实际 %q", prefix)
	}
	if rest != "xyz789" {
		t.Errorf("rest 应为 'xyz789', 实际 %q", rest)
	}
}

func TestParseCallbackData_Remove(t *testing.T) {
	prefix, rest := parseCallbackData("remove:gid123")
	if prefix != "remove" {
		t.Errorf("prefix 应为 'remove', 实际 %q", prefix)
	}
	if rest != "gid123" {
		t.Errorf("rest 应为 'gid123', 实际 %q", rest)
	}
}

func TestParseCallbackData_RemoveDone(t *testing.T) {
	prefix, rest := parseCallbackData("remove_done:gid456")
	if prefix != "remove_done" {
		t.Errorf("prefix 应为 'remove_done', 实际 %q", prefix)
	}
	if rest != "gid456" {
		t.Errorf("rest 应为 'gid456', 实际 %q", rest)
	}
}

func TestParseCallbackData_Page(t *testing.T) {
	prefix, rest := parseCallbackData("page:active:2")
	if prefix != "page" {
		t.Errorf("prefix 应为 'page', 实际 %q", prefix)
	}
	if rest != "active:2" {
		t.Errorf("rest 应为 'active:2', 实际 %q", rest)
	}
}

func TestParseCallbackData_Refresh(t *testing.T) {
	prefix, rest := parseCallbackData("refresh:active")
	if prefix != "refresh" {
		t.Errorf("prefix 应为 'refresh', 实际 %q", prefix)
	}
	if rest != "active" {
		t.Errorf("rest 应为 'active', 实际 %q", rest)
	}
}

func TestParseCallbackData_Confirm(t *testing.T) {
	prefix, rest := parseCallbackData("confirm:install")
	if prefix != "confirm" {
		t.Errorf("prefix 应为 'confirm', 实际 %q", prefix)
	}
	if rest != "install" {
		t.Errorf("rest 应为 'install', 实际 %q", rest)
	}
}

func TestParseCallbackData_Cancel(t *testing.T) {
	prefix, rest := parseCallbackData("cancel:uninstall")
	if prefix != "cancel" {
		t.Errorf("prefix 应为 'cancel', 实际 %q", prefix)
	}
	if rest != "uninstall" {
		t.Errorf("rest 应为 'uninstall', 实际 %q", rest)
	}
}

func TestParseCallbackData_EmptyData(t *testing.T) {
	prefix, rest := parseCallbackData("")
	if prefix != "" {
		t.Errorf("空 data 的 prefix 应为 '', 实际 %q", prefix)
	}
	if rest != "" {
		t.Errorf("空 data 的 rest 应为 '', 实际 %q", rest)
	}
}

func TestParseCallbackData_NoSeparator(t *testing.T) {
	prefix, rest := parseCallbackData("invalidformat")
	if prefix != "invalidformat" {
		t.Errorf("prefix 应为 'invalidformat', 实际 %q", prefix)
	}
	if rest != "" {
		t.Errorf("rest 应为 '', 实际 %q", rest)
	}
}

func TestParseCallbackData_UnknownPrefix(t *testing.T) {
	prefix, rest := parseCallbackData("unknown:somearg")
	if prefix != "unknown" {
		t.Errorf("prefix 应为 'unknown', 实际 %q", prefix)
	}
	if rest != "somearg" {
		t.Errorf("rest 应为 'somearg', 实际 %q", rest)
	}
}

func TestParseCallbackData_GidWithSpecialChars(t *testing.T) {
	// aria2 GID 是 16 进制字符串，应完整传递
	prefix, rest := parseCallbackData("pause:b469d2d63f9a8b3e")
	if prefix != "pause" {
		t.Errorf("prefix 应为 'pause', 实际 %q", prefix)
	}
	if rest != "b469d2d63f9a8b3e" {
		t.Errorf("rest 应为完整 GID, 实际 %q", rest)
	}
}

func TestValidActions(t *testing.T) {
	if !validActions["install"] {
		t.Error("install 应在白名单中")
	}
	if !validActions["uninstall"] {
		t.Error("uninstall 应在白名单中")
	}
	if !validActions["reset_secret"] {
		t.Error("reset_secret 应在白名单中")
	}
	if validActions["unknown"] {
		t.Error("unknown 不应在白名单中")
	}
	if validActions[""] {
		t.Error("空字符串不应在白名单中")
	}
}
