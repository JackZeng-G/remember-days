package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"remember/internal/config"
	"remember/internal/handler"
	"remember/internal/service"
	"remember/internal/store"
)

//go:embed web/templates/*.html web/static/*
var webAssets embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	logger := log.New(os.Stdout, "[REMEMBER] ", log.LstdFlags)

	// 存储层
	sqliteStore, err := store.NewSQLiteStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer sqliteStore.Close()

	if err := store.MigrateFromJSON(sqliteStore, cfg.DataDir); err != nil {
		logger.Printf("JSON数据迁移失败: %v", err)
	}

	// 服务层
	sessionMgr := service.NewSessionManager(cfg.SessionSecret)
	userSvc := service.NewUserService(sqliteStore, sqliteStore)
	annSvc := service.New(sqliteStore)

	// 模板
	tmpl, err := handler.NewTemplateRendererFromFS(webAssets)
	if err != nil {
		log.Fatalf("模板加载失败: %v", err)
	}

	// Handler
	h := handler.New(annSvc, userSvc, sessionMgr, tmpl)

	// 路由
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(handler.CSRFMiddleware())

	h.RegisterRoutes(r)

	staticFS, _ := fs.Sub(webAssets, "web/static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Printf("服务启动，地址: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalf("服务器启动失败: %v", err)
	}
}
