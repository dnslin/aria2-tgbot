# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 语言约定

- 与用户交流使用中文，称呼 BOSS
- 代码注释和日志消息使用中文

## 常用命令

```bash
# 编译
go build -o ./bin/aria2-tgbot ./cmd/bot/

# 运行全部测试
go test -v -count=1 ./...

# 运行单个包测试
go test -v -count=1 ./internal/config/

# 本地运行（需要先配置 config.yaml 中的 bot.token）
go run ./cmd/bot/ -c ./config.yaml

# 安装到系统（需要 root）
make install

# 清理编译产物
make clean
```

## 架构概览

`cmd/bot/main.go` 是唯一入口，遵循 `cmd/internal` Go 项目布局。分层依赖方向：

```
cmd → telegram → service → aria2
              ↘ config, logger
```

- **`internal/config/`** — YAML 配置加载、默认值填充、热更新持久化。Auth.Enabled 和 Message.AutoDelete 使用 `*bool` 指针区分"未设置"与"显式设为 false"。提供 `IsEnabled()` `IsAutoDeleteEnabled()` nil-safe 访问器。
- **`internal/logger/`** — 基于 lumberjack 的文本日志。格式 `[LEVEL] [文件:行号:函数] 消息`，支持 debug/info/warn/error 四级，`runtime.Caller(2)` 捕获调用位置。
- **`internal/aria2/`** — aria2 JSON-RPC HTTP 客户端 + 安装/卸载/升级编排。安装调用 P3TERX/aria2.sh 脚本。
- **`internal/telegram/`** — Bot 事件循环、命令注册、Inline Keyboard 回调、权限中间件、消息自动删除管理。
- **`internal/service/`** — 业务编排层，连接 telegram 和 aria2 层。

## 设计文档

完整设计说明见 `docs/superpowers/specs/2026-04-28-aria2-tgbot-design.md`。
GitHub Issues 用中文编写，Epic #8 追踪全部子任务。
