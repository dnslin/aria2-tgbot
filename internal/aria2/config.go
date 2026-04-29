package aria2

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// GenerateSecret 生成 16 位随机字母数字密钥
func GenerateSecret() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 16
	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("生成随机密钥失败: %w", err)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

// ConfigParams 用于生成 aria2.conf 的参数
type ConfigParams struct {
	DownloadDir string // 下载目录
	RpcPort     int    // RPC 端口
	RpcSecret   string // RPC 密钥
	SessionDir  string // 会话文件目录
	ConfigDir   string // 配置目录（用于 dht 文件等）
}

// GenerateConfig 根据参数生成 aria2.conf 配置内容
func GenerateConfig(p ConfigParams) string {
	var sb strings.Builder

	sb.WriteString("# ===== 基本设置 =====\n")
	sb.WriteString(fmt.Sprintf("dir=%s\n", p.DownloadDir))
	sb.WriteString("input-file=/root/.aria2c/aria2.session\n")
	sb.WriteString(fmt.Sprintf("save-session=%s/aria2.session\n", p.SessionDir))
	sb.WriteString("save-session-interval=30\n")
	sb.WriteString("force-save=true\n")
	sb.WriteString("log-level=info\n")
	sb.WriteString("\n")

	sb.WriteString("# ===== RPC 设置 =====\n")
	sb.WriteString(fmt.Sprintf("rpc-listen-port=%d\n", p.RpcPort))
	sb.WriteString("rpc-listen-all=true\n")
	sb.WriteString(fmt.Sprintf("rpc-secret=%s\n", p.RpcSecret))
	sb.WriteString("rpc-allow-origin-all=true\n")
	sb.WriteString("rpc-max-request-size=10M\n")
	sb.WriteString("\n")

	sb.WriteString("# ===== 网络连接设置 =====\n")
	sb.WriteString("max-concurrent-downloads=5\n")
	sb.WriteString("max-connection-per-server=16\n")
	sb.WriteString("min-split-size=20M\n")
	sb.WriteString("split=16\n")
	sb.WriteString("max-overall-download-limit=0\n")
	sb.WriteString("max-overall-upload-limit=0\n")
	sb.WriteString("disable-ipv6=true\n")
	sb.WriteString("check-certificate=true\n")
	sb.WriteString("user-agent=Transmission/4.0.4\n")
	sb.WriteString("\n")

	sb.WriteString("# ===== BT/PT 设置 =====\n")
	sb.WriteString("enable-dht=true\n")
	sb.WriteString("enable-dht6=false\n")
	sb.WriteString("enable-peer-exchange=true\n")
	sb.WriteString("bt-enable-lpd=true\n")
	sb.WriteString("bt-max-peers=128\n")
	sb.WriteString("bt-tracker-connect-timeout=10\n")
	sb.WriteString("bt-tracker-timeout=10\n")
	sb.WriteString("seed-ratio=1.0\n")
	sb.WriteString("seed-time=60\n")
	sb.WriteString("follow-torrent=mem\n")
	sb.WriteString("force-save=true\n")
	sb.WriteString(fmt.Sprintf("dht-file-path=%s/dht.dat\n", p.ConfigDir))
	sb.WriteString(fmt.Sprintf("dht-file-path6=%s/dht6.dat\n", p.ConfigDir))
	sb.WriteString("\n")

	sb.WriteString("# ===== 高级设置 =====\n")
	sb.WriteString("allow-overwrite=true\n")
	sb.WriteString("auto-file-renaming=true\n")
	sb.WriteString("file-allocation=falloc\n")
	sb.WriteString("console-log-level=info\n")
	sb.WriteString("continue=true\n")
	sb.WriteString("max-file-not-found=5\n")
	sb.WriteString("max-tries=0\n")
	sb.WriteString("max-resume-failure-tries=0\n")
	sb.WriteString("always-resume=true\n")
	sb.WriteString("keep-unfinished-download-result=true\n")
	sb.WriteString("remove-control-file=true\n")
	sb.WriteString("piece-length=1M\n")
	sb.WriteString("realtime-chunk-checksum=true\n")
	sb.WriteString("content-disposition-default-utf8=true\n")
	sb.WriteString("summary-interval=0\n")
	sb.WriteString("disk-cache=64M\n")

	return sb.String()
}
