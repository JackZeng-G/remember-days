package handler

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

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
	"formatDate": func(dateStr string) string {
		if len(dateStr) == 10 {
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				return t.Format("2006年1月2日")
			}
		}
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

// TemplateRenderer 模板渲染器
type TemplateRenderer struct {
	templates *template.Template
}

// NewTemplateRendererFromFS 从 embed.FS 创建模板渲染器
func NewTemplateRendererFromFS(fsys fs.FS) (*TemplateRenderer, error) {
	sub, err := fs.Sub(fsys, "web/templates")
	if err != nil {
		return nil, fmt.Errorf("访问模板目录失败: %w", err)
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFS(sub, "*.html")
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
