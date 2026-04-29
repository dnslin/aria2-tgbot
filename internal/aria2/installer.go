package aria2

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 备用脚本下载地址列表
var ScriptURLs = []string{
	"https://git.io/aria2.sh",
	"https://raw.githubusercontent.com/P3TERX/aria2.sh/master/aria2.sh",
}

// LogFunc 日志/进度回调函数类型
type LogFunc func(format string, args ...any)

// Installer 负责 aria2 的安装、卸载、升级和进程管理
type Installer struct {
	client      *Client
	scriptURLs  []string
	installPath string
	configDir   string
	downloadDir string
	rpcHost     string
	rpcPort     int
	sessionDir  string
	autoStart   bool
	log         LogFunc
}

// NewInstaller 创建 Installer 实例
func NewInstaller(rpcHost string, rpcPort int, installPath, configDir, downloadDir, sessionDir string, autoStart bool) *Installer {
	return &Installer{
		scriptURLs:  ScriptURLs,
		installPath: installPath,
		configDir:   configDir,
		downloadDir: downloadDir,
		rpcHost:     rpcHost,
		rpcPort:     rpcPort,
		sessionDir:  sessionDir,
		autoStart:   autoStart,
	}
}

// SetLogCallback 设置日志回调函数，用于向调用者报告进度
func (i *Installer) SetLogCallback(fn LogFunc) {
	i.log = fn
}

// logf 输出日志（如果有回调）
func (i *Installer) logf(format string, args ...any) {
	if i.log != nil {
		i.log(format, args...)
	}
}

// ===== 安装/卸载/升级 =====

// Install 安装 aria2：下载脚本 → 执行安装 → 生成配置 → 自动启动，返回生成的密钥
func (i *Installer) Install() (string, error) {
	if i.IsInstalled() {
		return "", fmt.Errorf("aria2 已安装，如需重装请先使用 /uninstall 卸载")
	}

	// 1. 下载安装脚本
	i.logf("正在下载 aria2 安装脚本...")
	scriptPath, err := i.downloadScript()
	if err != nil {
		return "", fmt.Errorf("下载安装脚本失败: %w", err)
	}
	defer os.Remove(scriptPath)

	// 2. 执行安装
	i.logf("正在执行安装脚本，请耐心等待...")
	if err := i.runBash(scriptPath); err != nil {
		return "", fmt.Errorf("安装脚本执行失败: %w", err)
	}

	// 3. 生成密钥并写入配置
	secret, err := GenerateSecret()
	if err != nil {
		return "", fmt.Errorf("生成密钥失败: %w", err)
	}

	i.logf("正在生成 aria2 配置文件...")
	confContent := GenerateConfig(ConfigParams{
		DownloadDir: i.downloadDir,
		RpcPort:     i.rpcPort,
		RpcSecret:   secret,
		SessionDir:  i.sessionDir,
		ConfigDir:   i.configDir,
	})

	confPath := filepath.Join(i.configDir, "aria2.conf")
	if err := os.MkdirAll(i.configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %w", err)
	}
	if err := os.WriteFile(confPath, []byte(confContent), 0600); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %w", err)
	}

	// 创建 session 目录和文件
	if err := os.MkdirAll(i.sessionDir, 0755); err != nil {
		return "", fmt.Errorf("创建 session 目录失败: %w", err)
	}
	sessionFile := filepath.Join(i.sessionDir, "aria2.session")
	if err := os.WriteFile(sessionFile, []byte{}, 0644); err != nil {
		return "", fmt.Errorf("创建 session 文件失败: %w", err)
	}

	// 4. 初始化客户端连接
	i.client = NewClient(i.rpcHost, i.rpcPort, secret)

	// 5. 自动启动
	if i.autoStart {
		i.logf("正在启动 aria2...")
		if err := i.Start(); err != nil {
			return secret, fmt.Errorf("aria2 安装完成但启动失败: %w", err)
		}
		i.logf("aria2 安装并启动成功！")
	} else {
		i.logf("aria2 安装完成（未自动启动）")
	}

	return secret, nil
}

// Uninstall 卸载 aria2：停止进程 → 执行卸载 → 清理残留文件
func (i *Installer) Uninstall() error {
	if !i.IsInstalled() {
		return fmt.Errorf("aria2 未安装")
	}

	// 1. 停止进程
	i.logf("正在停止 aria2 进程...")
	if err := i.Stop(); err != nil {
		i.logf("停止 aria2 时出现警告: %v", err)
	}

	// 2. 下载脚本并执行卸载
	i.logf("正在下载卸载脚本...")
	scriptPath, err := i.downloadScript()
	if err != nil {
		return fmt.Errorf("下载卸载脚本失败: %w", err)
	}
	defer os.Remove(scriptPath)

	i.logf("正在执行卸载...")
	if err := i.runBash(scriptPath, "uninstall"); err != nil {
		return fmt.Errorf("卸载脚本执行失败: %w", err)
	}

	// 3. 清理残留文件
	i.logf("正在清理残留文件...")
	os.RemoveAll(i.configDir)
	os.RemoveAll(i.sessionDir)

	i.client = nil
	i.logf("aria2 已成功卸载")
	return nil
}

// Upgrade 升级 aria2 到最新版本（保留配置）
func (i *Installer) Upgrade() (string, error) {
	if !i.IsInstalled() {
		return "", fmt.Errorf("aria2 未安装，请先使用 /install 安装")
	}

	i.logf("正在下载最新安装脚本...")
	scriptPath, err := i.downloadScript()
	if err != nil {
		return "", fmt.Errorf("下载升级脚本失败: %w", err)
	}
	defer os.Remove(scriptPath)

	i.logf("正在执行升级（配置将被保留）...")
	// 传入 "1" 选择安装/升级选项，避免交互菜单阻塞
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdin = strings.NewReader("1\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("升级脚本执行失败: %s", string(out))
	}
	i.logf("  升级脚本执行完成")

	// 重启以应用新版本
	if i.IsRunning() {
		i.logf("正在重启 aria2 以应用新版本...")
		i.Restart()
	}

	// 获取新版本号
	version := "最新版"
	if i.client != nil {
		if info, err := i.client.GetVersion(); err == nil {
			version = info.Version
		}
	}

	i.logf("aria2 升级完成: %s", version)
	return version, nil
}

// ===== 进程管理 =====

// Start 启动 aria2 进程
func (i *Installer) Start() error {
	confPath := filepath.Join(i.configDir, "aria2.conf")

	// 检查配置文件是否存在
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s，请先运行 /install", confPath)
	}

	// 先尝试 systemctl
	if i.hasSystemd() {
		cmd := exec.Command("systemctl", "start", "aria2")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("systemctl start 失败: %s", string(out))
		}
		return nil
	}

	// 否则直接后台启动
	cmd := exec.Command("aria2c", "--conf-path="+confPath, "-D")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启动 aria2 失败: %s", string(out))
	}

	// 等待进程就绪
	time.Sleep(1 * time.Second)
	return nil
}

// Stop 停止 aria2 进程（幂等：进程已停止时返回 nil）
func (i *Installer) Stop() error {
	if i.hasSystemd() {
		cmd := exec.Command("systemctl", "stop", "aria2")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("systemctl stop 失败: %s", string(out))
		}
		return nil
	}

	// 先检查进程是否存在
	if !i.IsRunning() {
		return nil
	}

	cmd := exec.Command("pkill", "-f", "aria2c")
	if out, err := cmd.CombinedOutput(); err != nil {
		// pkill 退出码 1 表示无匹配进程，视为正常
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("停止 aria2 失败: %s", string(out))
	}
	return nil
}

// Restart 重启 aria2 进程
func (i *Installer) Restart() error {
	if i.hasSystemd() {
		cmd := exec.Command("systemctl", "restart", "aria2")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("systemctl restart 失败: %s", string(out))
		}
		return nil
	}

	if err := i.Stop(); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return i.Start()
}

// Status 获取 aria2 进程运行状态
func (i *Installer) Status() (*ProcessStatus, error) {
	status := &ProcessStatus{Running: false}

	if !i.IsInstalled() {
		return status, nil
	}

	status.Running = i.IsRunning()

	if i.client != nil {
		if info, err := i.client.GetVersion(); err == nil {
			status.Version = info.Version
		}
		if active, err := i.client.TellActive(); err == nil {
			status.Tasks = len(active)
		}
	}

	// 获取 PID 和运行时长
	if status.Running {
		out, err := exec.Command("pgrep", "-f", "aria2c").Output()
		if err == nil {
			pidStr := strings.TrimSpace(string(out))
			if pid, parseErr := strconv.Atoi(strings.Split(pidStr, "\n")[0]); parseErr == nil {
				status.PID = pid
			}
		}

		// 通过 ps 获取运行时长
		out, err = exec.Command("ps", "-p", strconv.Itoa(status.PID), "-o", "etime=").Output()
		if err == nil {
			status.Uptime = strings.TrimSpace(string(out))
		}
	}

	return status, nil
}

// IsInstalled 检测 aria2c 二进制是否存在
func (i *Installer) IsInstalled() bool {
	binPath := filepath.Join(i.installPath, "aria2c")
	_, err := os.Stat(binPath)
	return err == nil
}

// IsRunning 检测 aria2 进程是否运行中
func (i *Installer) IsRunning() bool {
	// 尝试通过 pgrep 检测
	cmd := exec.Command("pgrep", "-f", "aria2c")
	err := cmd.Run()
	if err == nil {
		return true
	}

	// pgrep 失败时，尝试 RPC 连接
	if i.client != nil {
		if _, err := i.client.GetVersion(); err == nil {
			return true
		}
	}

	return false
}

// HealthCheck 综合健康检查
func (i *Installer) HealthCheck() (*HealthInfo, error) {
	info := &HealthInfo{
		Aria2Running: i.IsRunning(),
	}

	if i.IsRunning() && i.client != nil {
		if ver, err := i.client.GetVersion(); err == nil {
			info.Aria2Version = ver.Version
			info.RPCOK = true
		}
	}

	// 检查下载目录磁盘空间
	if stat, err := diskUsage(i.downloadDir); err == nil {
		info.DiskFree = formatBytes(stat.Free)
		info.DiskPercent = stat.UsedPercent
	}

	return info, nil
}

// ===== 内部方法 =====

// scriptHTTPClient 用于下载脚本的 HTTP 客户端（30 秒超时）
var scriptHTTPClient = &http.Client{Timeout: 30 * time.Second}

// downloadScript 下载安装脚本，支持多 URL fallback
func (i *Installer) downloadScript() (string, error) {
	var lastErr error

	for _, url := range i.scriptURLs {
		i.logf("  尝试: %s", url)

		resp, err := scriptHTTPClient.Get(url)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		tmpFile, err := os.CreateTemp("", "aria2-sh-*.sh")
		if err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("创建临时文件失败: %w", err)
		}

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			resp.Body.Close()
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			lastErr = err
			continue
		}
		resp.Body.Close()
		tmpFile.Close()

		// 确保脚本可执行
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("设置脚本可执行权限失败: %w", err)
		}

		return tmpFile.Name(), nil
	}

	return "", fmt.Errorf("所有下载地址均失败，最后错误: %v", lastErr)
}

// runBash 以 bash 执行脚本
func (i *Installer) runBash(scriptPath string, args ...string) error {
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("bash", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("脚本执行失败: %s", string(output))
	}

	// 输出最后几行作为日志
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	start := len(lines) - 5
	if start < 0 {
		start = 0
	}
	for _, line := range lines[start:] {
		i.logf("  %s", line)
	}

	return nil
}

// hasSystemd 检测系统是否使用 systemd
func (i *Installer) hasSystemd() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

// diskUsage 磁盘使用信息
type diskUsageInfo struct {
	Free        uint64
	UsedPercent float64
}

func diskUsage(path string) (*diskUsageInfo, error) {
	// 使用 df 命令获取磁盘信息
	cmd := exec.Command("df", "-B1", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("df 输出格式异常")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return nil, fmt.Errorf("df 输出字段不足")
	}

	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析磁盘总量失败: %w", err)
	}
	used, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析磁盘已用失败: %w", err)
	}
	free, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析磁盘可用失败: %w", err)
	}

	var usedPercent float64
	if total > 0 {
		usedPercent = float64(used) / float64(total) * 100
	}

	return &diskUsageInfo{Free: free, UsedPercent: usedPercent}, nil
}

// formatBytes 将字节数格式化为人类可读的字符串
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
