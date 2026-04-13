# 构建脚本
$GOOS = "linux"
$GOARCH = "amd64"

Write-Host "Building for Linux/amd64..."

$env:GOOS = $GOOS
$env:GOARCH = $GOARCH

go build -o server ./cmd/server

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful: server"
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  docker-compose build"
    Write-Host "  docker-compose up -d"
} else {
    Write-Host "Build failed!"
    exit 1
}