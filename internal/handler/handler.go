package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"remember/internal/service"
)

// Handler HTTP处理器
type Handler struct {
	service service.AnniversaryService
	tmpl    *TemplateRenderer
	logger  *log.Logger
}

// New 创建新的 Handler
func New(svc service.AnniversaryService, tmpl *TemplateRenderer, logger *log.Logger) *Handler {
	return &Handler{
		service: svc,
		tmpl:    tmpl,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.Index)
	r.Get("/add", h.AddForm)
	r.Post("/add", h.Add)
	r.Get("/edit/{id}", h.EditForm)
	r.Post("/edit/{id}", h.Edit)
	r.Post("/delete/{id}", h.Delete)
	r.Get("/api/reminders", h.APIReminders)
	r.Get("/api/status", h.APIStatus)
}

// Index 首页
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	views, err := h.service.List()
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	csrfToken, _ := GetCSRFCookie(r)

	data := struct {
		Views     interface{}
		Now       string
		CSRFToken string
	}{
		Views:     views,
		Now:       time.Now().Format("2006年1月2日"),
		CSRFToken: csrfToken,
	}

	if err := h.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

// AddForm 添加表单页面
func (h *Handler) AddForm(w http.ResponseWriter, r *http.Request) {
	csrfToken, _ := GetCSRFCookie(r)
	data := struct {
		CSRFToken string
	}{
		CSRFToken: csrfToken,
	}
	if err := h.tmpl.ExecuteTemplate(w, "add.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

// Add 处理添加请求
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	year := strings.TrimSpace(r.FormValue("year"))
	month := strings.TrimSpace(r.FormValue("month"))
	day := strings.TrimSpace(r.FormValue("day"))
	desc := strings.TrimSpace(r.FormValue("description"))

	date := year + "-" + month + "-" + day

	_, err := h.service.Create(name, date, desc)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	h.logger.Printf("添加纪念日: %s (%s)", service.SanitizeForLog(name), date)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditForm 编辑表单页面
func (h *Handler) EditForm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := h.service.Get(id)
	if err != nil {
		h.handleError(w, err, http.StatusNotFound)
		return
	}

	csrfToken, _ := GetCSRFCookie(r)

	// 解析日期
	parts := strings.Split(view.Date, "-")
	year, month, day := "", "", ""
	if len(parts) == 3 {
		year = parts[0]
		month = parts[1]
		day = parts[2]
	}

	data := struct {
		ID          string
		Name        string
		Year        string
		Month       string
		Day         string
		Description string
		CSRFToken   string
	}{
		ID:          view.ID,
		Name:        view.Name,
		Year:        year,
		Month:       month,
		Day:         day,
		Description: view.Description,
		CSRFToken:   csrfToken,
	}

	if err := h.tmpl.ExecuteTemplate(w, "edit.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

// Edit 处理编辑请求
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	name := strings.TrimSpace(r.FormValue("name"))
	year := strings.TrimSpace(r.FormValue("year"))
	month := strings.TrimSpace(r.FormValue("month"))
	day := strings.TrimSpace(r.FormValue("day"))
	desc := strings.TrimSpace(r.FormValue("description"))

	date := year + "-" + month + "-" + day

	err := h.service.Update(id, name, date, desc)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	h.logger.Printf("编辑纪念日: %s (%s)", service.SanitizeForLog(name), date)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Delete 处理删除请求
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.Delete(id)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	h.logger.Printf("删除纪念日: %s", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// APIReminders JSON API
func (h *Handler) APIReminders(w http.ResponseWriter, r *http.Request) {
	views, err := h.service.List()
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

// APIStatus 服务状态
func (h *Handler) APIStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "2.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleError 统一错误处理
func (h *Handler) handleError(w http.ResponseWriter, err error, statusCode int) {
	h.logger.Printf("错误: %v", err)
	http.Error(w, err.Error(), statusCode)
}
