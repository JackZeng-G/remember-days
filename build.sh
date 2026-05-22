#!/bin/sh
# 构建 Linux 版本用于 Docker 部署
set -e

echo "构建 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -o build/remember .

echo "完成: build/remember"
ls -lh build/remember