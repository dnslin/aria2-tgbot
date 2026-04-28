.PHONY: build run test install clean

# 编译二进制
build:
	go build -o ./bin/aria2-tgbot ./cmd/bot/

# 本地运行
run:
	go run ./cmd/bot/

# 运行测试
test:
	go test -v -count=1 ./...

# 安装到系统（需要 root 权限）
install: build
	install -d /etc/aria2-tgbot
	cp ./bin/aria2-tgbot /usr/local/bin/aria2-tgbot
	cp ./config.yaml /etc/aria2-tgbot/config.yaml
	cp ./aria2-tgbot.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable aria2-tgbot
	systemctl start aria2-tgbot

# 清理编译产物
clean:
	rm -rf ./bin/
