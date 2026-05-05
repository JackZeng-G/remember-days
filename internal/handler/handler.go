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
	userSvc service.UserService
	session *service.SessionManager
	tmpl    *TemplateRenderer
	logger  *log.Logger
}

// New 创建新的 Handler
func New(svc service.AnniversaryService, userSvc service.UserService, session *service.SessionManager, tmpl *TemplateRenderer, logger *log.Logger) *Handler {
	return &Handler{
		service: svc,
		userSvc: userSvc,
		session: session,
		tmpl:    tmpl,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r chi.Router) {
	// 公开路由
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
	r.Get("/register", h.RegisterForm)
	r.Post("/register", h.Register)
	r.Get("/logout", h.Logout)
	r.Get("/api/status", h.APIStatus)

	// 认证保护路由
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(h.session))

		r.Get("/", h.Index)
		r.Get("/add", h.AddForm)
		r.Post("/add", h.Add)
		r.Get("/edit/{id}", h.EditForm)
		r.Post("/edit/{id}", h.Edit)
		r.Post("/delete/{id}", h.Delete)
		r.Get("/api/reminders", h.APIReminders)
	})

	// 管理员路由
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(h.session))
		r.Use(AdminMiddleware(h.userSvc))

		r.Get("/admin", h.AdminUsers)
		r.Post("/admin/users/{id}/delete", h.AdminDeleteUser)
	})
}

// ========== 认证 Handler ==========

func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	if userID, _ := h.session.GetSession(r); userID != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	csrfToken := GetCSRFToken(r)
	h.tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"CSRFToken": csrfToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "请求解析失败", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	user, err := h.userSvc.Login(username, password)
	if err != nil {
		csrfToken := GetCSRFToken(r)
		h.tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"CSRFToken": csrfToken,
			"Error":     "用户名或密码错误",
			"Username":  username,
		})
		return
	}

	h.session.CreateSession(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	if userID, _ := h.session.GetSession(r); userID != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	csrfToken := GetCSRFToken(r)
	h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
		"CSRFToken": csrfToken,
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "请求解析失败", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	csrfToken := GetCSRFToken(r)
	data := map[string]interface{}{
		"CSRFToken": csrfToken,
		"Username":  username,
	}

	if password != confirm {
		data["Error"] = "两次密码不一致"
		h.tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	user, err := h.userSvc.Register(username, password)
	if err != nil {
		data["Error"] = err.Error()
		h.tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	h.session.CreateSession(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.session.DestroySession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ========== 纪念日 Handler ==========

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	user, _ := h.userSvc.GetByID(userID)

	views, err := h.service.List(userID)
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	csrfToken := GetCSRFToken(r)

	data := struct {
		Views     interface{}
		Now       string
		CSRFToken string
		Username  string
		IsAdmin   bool
	}{
		Views:     views,
		Now:       time.Now().Format("2006年1月2日"),
		CSRFToken: csrfToken,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
	}

	if err := h.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

func (h *Handler) AddForm(w http.ResponseWriter, r *http.Request) {
	csrfToken := GetCSRFToken(r)
	data := map[string]interface{}{
		"CSRFToken": csrfToken,
	}
	if err := h.tmpl.ExecuteTemplate(w, "add.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	name := strings.TrimSpace(r.FormValue("name"))
	year := strings.TrimSpace(r.FormValue("year"))
	month := strings.TrimSpace(r.FormValue("month"))
	day := strings.TrimSpace(r.FormValue("day"))
	desc := strings.TrimSpace(r.FormValue("description"))

	date := year + "-" + month + "-" + day

	_, err := h.service.Create(userID, name, date, desc)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) EditForm(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	view, err := h.service.Get(userID, id)
	if err != nil {
		h.handleError(w, err, http.StatusNotFound)
		return
	}

	csrfToken := GetCSRFToken(r)

	parts := strings.Split(view.Date, "-")
	year, month, day := "", "", ""
	if len(parts) == 3 {
		year = parts[0]
		month = parts[1]
		day = parts[2]
	}

	data := map[string]interface{}{
		"ID":          view.ID,
		"Name":        view.Name,
		"Year":        year,
		"Month":       month,
		"Day":         day,
		"Description": view.Description,
		"CSRFToken":   csrfToken,
	}

	if err := h.tmpl.ExecuteTemplate(w, "edit.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	id := chi.URLParam(r, "id")
	name := strings.TrimSpace(r.FormValue("name"))
	year := strings.TrimSpace(r.FormValue("year"))
	month := strings.TrimSpace(r.FormValue("month"))
	day := strings.TrimSpace(r.FormValue("day"))
	desc := strings.TrimSpace(r.FormValue("description"))

	date := year + "-" + month + "-" + day

	err := h.service.Update(userID, id, name, date, desc)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	err := h.service.Delete(userID, id)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) APIReminders(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	views, err := h.service.List(userID)
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

func (h *Handler) APIStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "3.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ========== 管理员 Handler ==========

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userSvc.ListAll()
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	currentID := GetUserID(r)

	type UserView struct {
		ID        string
		Username  string
		IsAdmin   bool
		CreatedAt string
		Count     int
	}

	var userViews []UserView
	for _, u := range users {
		count := 0
		if anns, err := h.service.List(u.ID); err == nil {
			count = len(anns)
		}
		userViews = append(userViews, UserView{
			ID:        u.ID,
			Username:  u.Username,
			IsAdmin:   u.IsAdmin,
			CreatedAt: u.CreatedAt,
			Count:     count,
		})
	}

	csrfToken := GetCSRFToken(r)
	data := map[string]interface{}{
		"Users":     userViews,
		"CurrentID": currentID,
		"CSRFToken": csrfToken,
	}

	if err := h.tmpl.ExecuteTemplate(w, "admin.html", data); err != nil {
		h.logger.Printf("模板渲染失败: %v", err)
	}
}

func (h *Handler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "id")
	currentID := GetUserID(r)

	if err := h.userSvc.DeleteUser(targetID, currentID); err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleError(w http.ResponseWriter, err error, statusCode int) {
	h.logger.Printf("错误: %v", err)
	http.Error(w, http.StatusText(statusCode), statusCode)
}
