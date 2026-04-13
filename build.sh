#!/bin/bash
# 构建脚本

set -e

echo "Building for Linux/amd64..."

GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

echo "Build successful: server"
echo ""
echo "Next steps:"
echo "  docker compose build"
echo "  docker compose up -d"