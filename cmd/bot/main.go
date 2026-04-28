// Aria2 Telegram Bot 入口。
// 加载配置、初始化日志、检测 aria2 状态，然后启动 Bot。
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dnslin/aria2-tgbot/internal/config"
	"github.com/dnslin/aria2-tgbot/internal/logger"
)

func main() {
	configPath := flag.String("c", "./config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.New(logger.Config{
		Dir:        cfg.Log.Dir,
		Level:      cfg.Log.Level,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Aria2 Telegram Bot 启动中...")
	log.Info("配置文件: %s", *configPath)
	log.Info("日志级别: %s", cfg.Log.Level)
	log.Info("权限控制: %v", cfg.Auth.IsEnabled())
	log.Info("Bot 已就绪，等待 Telegram Updates")

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Info("收到信号 %v，Bot 正在关闭...", sig)
}
