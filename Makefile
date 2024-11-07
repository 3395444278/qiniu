.PHONY: run build clean test

# 设置 Go 编译器参数
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR = build

# 主要的二进制文件名
BINARY_NAME = qiniu-app

# 开发环境运行
run:
ifeq ($(OS),Windows_NT)
	scripts\start.bat
else
	./scripts/start.sh
endif

# 构建项目
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)

# 运行测试
test:
	go test -v ./...

# 安装依赖
deps:
	go mod download
	go get github.com/redis/go-redis/v9

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 启动 Redis
redis:
ifeq ($(OS),Windows_NT)
	start redis-server
else
	redis-server &
endif

# 帮助信息
help:
	@echo "可用的命令："
	@echo "  make run         - 运行开发环境"
	@echo "  make build      - 构建项目"
	@echo "  make clean      - 清理构建文件"
	@echo "  make test       - 运行测试"
	@echo "  make deps       - 安装依赖"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 运行代码检查"
	@echo "  make redis      - 启动 Redis"