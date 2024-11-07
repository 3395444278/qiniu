@echo off
echo Starting Redis server...
start redis-server

:: 等待 Redis 启动
timeout /t 2

:: 启动爬虫服务（包含评估服务）
go run cmd/crawler/main.go %*

:: 等待用户按键退出
pause 