package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 配置
const (
	Port        = 8080
	ServiceName = "纪念日提醒服务"
)

var (
	logFile *os.File
)

func main() {
	// 获取程序所在目录
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// 切换工作目录
	os.Chdir(exeDir)

	// 创建日志文件
	var err error
	logFile, err = os.OpenFile("anniversary.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}

	log.Printf("========================================")
	log.Printf("  %s 启动", ServiceName)
	log.Printf("  时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("========================================")

	// 创建必要的目录
	os.MkdirAll("data", 0755)

	// 设置路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/api/reminders", apiHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 启动服务器
	url := fmt.Sprintf("http://localhost:%d", Port)
	log.Printf("服务地址: %s", url)

	// 启动HTTP服务器
	if err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil); err != nil {
		log.Printf("启动服务器失败: %v", err)
		if logFile != nil {
			logFile.Close()
		}
	}
}

// 模板函数
var templateFuncs = template.FuncMap{
	"mul": func(a, b interface{}) float64 {
		var af, bf float64
		switch v := a.(type) {
		case int:
			af = float64(v)
		case float64:
			af = v
		}
		switch v := b.(type) {
		case int:
			bf = float64(v)
		case float64:
			bf = v
		}
		return af * bf
	},
	"iterate": func(count int) []int {
		var items []int
		for i := 1; i <= count; i++ {
			items = append(items, i)
		}
		return items
	},
	"formatDate": func(dateStr string) string {
		// 尝试解析 YYYY-MM-DD 格式
		if len(dateStr) == 10 {
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				return t.Format("2006年1月2日")
			}
		}
		// 尝试解析 MM-DD 格式（旧数据）
		if len(dateStr) == 5 {
			t, err := time.Parse("01-02", dateStr)
			if err == nil {
				return t.Format("1月2日")
			}
		}
		return dateStr
	},
}

// indexHandler 显示主页
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").Funcs(templateFuncs).ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	storage, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views := GetAnniversaryView(storage.Anniversaries)

	data := struct {
		Views []AnniversaryView
		Now   string
	}{
		Views: views,
		Now:   time.Now().Format("2006年1月2日"),
	}

	tmpl.Execute(w, data)
}

// addHandler 处理添加请求
func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.New("add.html").Funcs(templateFuncs).ParseFiles("templates/add.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		name := strings.TrimSpace(r.FormValue("name"))
		year := strings.TrimSpace(r.FormValue("year"))
		month := strings.TrimSpace(r.FormValue("month"))
		day := strings.TrimSpace(r.FormValue("day"))
		desc := strings.TrimSpace(r.FormValue("description"))

		if name == "" || year == "" || month == "" || day == "" {
			http.Error(w, "名称、年份、月份和日期不能为空", http.StatusBadRequest)
			return
		}

		date := year + "-" + month + "-" + day
		if _, err := time.Parse("2006-01-02", date); err != nil {
			http.Error(w, "日期格式错误", http.StatusBadRequest)
			return
		}

		storage, err := LoadData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		AddAnniversary(storage, name, date, desc)
		SaveData(storage)

		log.Printf("添加纪念日: %s (%s)", name, date)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// editHandler 处理编辑请求
func editHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/edit/")
	if id == "" {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	storage, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 查找纪念日
	var ann *Anniversary
	for i := range storage.Anniversaries {
		if storage.Anniversaries[i].ID == id {
			ann = &storage.Anniversaries[i]
			break
		}
	}

	if ann == nil {
		http.Error(w, "纪念日不存在", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		// 解析日期 (YYYY-MM-DD 或 MM-DD 格式)
		parts := strings.Split(ann.Date, "-")
		year := ""
		month := ""
		day := ""
		if len(parts) == 3 {
			year = parts[0]
			month = parts[1]
			day = parts[2]
		} else if len(parts) == 2 {
			month = parts[0]
			day = parts[1]
		}

		data := struct {
			ID          string
			Name        string
			Year        string
			Month       string
			Day         string
			Description string
		}{
			ID:          ann.ID,
			Name:        ann.Name,
			Year:        year,
			Month:       month,
			Day:         day,
			Description: ann.Description,
		}

		tmpl, err := template.New("edit.html").Funcs(templateFuncs).ParseFiles("templates/edit.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		name := strings.TrimSpace(r.FormValue("name"))
		year := strings.TrimSpace(r.FormValue("year"))
		month := strings.TrimSpace(r.FormValue("month"))
		day := strings.TrimSpace(r.FormValue("day"))
		desc := strings.TrimSpace(r.FormValue("description"))

		if name == "" || year == "" || month == "" || day == "" {
			http.Error(w, "名称、年份、月份和日期不能为空", http.StatusBadRequest)
			return
		}

		date := year + "-" + month + "-" + day
		if _, err := time.Parse("2006-01-02", date); err != nil {
			http.Error(w, "日期格式错误", http.StatusBadRequest)
			return
		}

		// 更新数据
		for i := range storage.Anniversaries {
			if storage.Anniversaries[i].ID == id {
				storage.Anniversaries[i].Name = name
				storage.Anniversaries[i].Date = date
				storage.Anniversaries[i].Description = desc
				break
			}
		}

		SaveData(storage)
		log.Printf("编辑纪念日: %s (%s)", name, date)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// deleteHandler 处理删除请求
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/delete/")
	if id == "" {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	storage, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	DeleteAnniversary(storage, id)
	SaveData(storage)

	log.Printf("删除纪念日: %s", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// apiHandler 返回JSON API
func apiHandler(w http.ResponseWriter, r *http.Request) {
	storage, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views := GetAnniversaryView(storage.Anniversaries)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

// statusHandler 返回服务状态
func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
