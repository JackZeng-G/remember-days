package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"anniversary/internal/config"
	"anniversary/internal/handler"
	"anniversary/internal/service"
	"anniversary/internal/store"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 设置日志
	logger := log.New(os.Stdout, "[ANNIVERSARY] ", log.LstdFlags)

	// 获取程序所在目录
	exePath, err := os.Executable()
	if err != nil {
		logger.Printf("获取程序路径失败: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	os.Chdir(exeDir)

	// 初始化存储
	jsonStore := store.NewJSONStore(cfg.DataDir)

	// 初始化服务
	svc := service.New(jsonStore)

	// 初始化模板渲染器
	tmpl, err := handler.NewTemplateRenderer("web/templates/*.html")
	if err != nil {
		log.Fatalf("模板加载失败: %v", err)
	}

	// 创建 Handler
	h := handler.New(svc, tmpl, logger)

	// 创建路由
	r := chi.NewRouter()

	// 中间件链
	r.Use(middleware.Recoverer)
	r.Use(handler.LoggingMiddleware(logger))
	r.Use(handler.CSRFMiddleware(cfg))

	// 注册路由
	h.RegisterRoutes(r)

	// 静态文件服务
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Printf("服务启动，地址: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalf("服务器启动失败: %v", err)
	}
}