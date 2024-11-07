#!/bin/bash

# 检查 Redis 是否运行
if ! pgrep redis-server > /dev/null; then
    echo "Starting Redis server..."
    redis-server &
    sleep 2
fi

# 启动爬虫服务（包含评估服务）
go run cmd/crawler/main.go $@

# 等待所有进程完成
wait 