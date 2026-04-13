package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// TemplateRenderer 模板渲染器
type TemplateRenderer struct {
	templates *template.Template
}

// NewTemplateRenderer 创建模板渲染器
func NewTemplateRenderer(glob string) (*TemplateRenderer, error) {
	funcs := template.FuncMap{
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
		"jsEscape": template.JSEscaper,
	}

	tmpl, err := template.New("").Funcs(funcs).ParseGlob(glob)
	if err != nil {
		return nil, fmt.Errorf("解析模板失败: %w", err)
	}

	return &TemplateRenderer{templates: tmpl}, nil
}

// ExecuteTemplate 渲染指定模板
func (tr *TemplateRenderer) ExecuteTemplate(w http.ResponseWriter, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tr.templates.ExecuteTemplate(w, name, data)
}