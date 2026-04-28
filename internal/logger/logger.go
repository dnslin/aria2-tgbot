// Package logger 提供文本格式的结构化日志输出。
// 支持 debug/info/warn/error 四个级别，自动记录调用位置（文件:行号:函数名）。
// 基于 lumberjack 实现日志文件按大小轮转。
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Level 表示日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Config 日志配置
type Config struct {
	Dir        string // 日志文件目录
	Level      string // debug / info / warn / error
	MaxSize    int    // 单文件最大 MB
	MaxBackups int    // 保留旧文件数
	MaxAge     int    // 保留天数
}

// Logger 日志实例
type Logger struct {
	internal *log.Logger
	level    Level
	closer   io.Closer
}

// New 创建 Logger 实例。日志目录不存在时自动创建。
func New(cfg Config) (*Logger, error) {
	level := parseLevel(cfg.Level)

	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录 %s 失败: %w", cfg.Dir, err)
	}

	logPath := filepath.Join(cfg.Dir, "bot.log")
	writer := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		LocalTime:  true,
	}

	l := &Logger{
		internal: log.New(writer, "", 0),
		level:    level,
		closer:   writer,
	}

	return l, nil
}

// Debug 输出调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 输出信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 输出警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 输出错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Close 关闭日志文件句柄
func (l *Logger) Close() {
	if l.closer != nil {
		l.closer.Close()
	}
}

// log 内部日志输出
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)

	// 获取调用位置（跳过 2 帧：log → Debug/Info/Warn/Error → 调用者）
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	// 仅保留文件名而非完整路径
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	// 获取函数名
	pc, _, _, _ := runtime.Caller(2)
	funcName := "???"
	if f := runtime.FuncForPC(pc); f != nil {
		funcName = f.Name()
		// 仅保留方法名，去掉包路径
		if idx := strings.LastIndex(funcName, "."); idx >= 0 {
			funcName = funcName[idx+1:]
		}
	}

	// 格式: [LEVEL] [文件:行号:函数] 消息
	l.internal.Printf("[%s] [%s:%d:%s] %s", levelNames[level], file, line, funcName, msg)
}

// parseLevel 将配置字符串转为 Level 常量
func parseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}
