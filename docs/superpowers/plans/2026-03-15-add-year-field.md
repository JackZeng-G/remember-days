# 纪念日增加年份字段 实施计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为纪念日系统增加年份字段，支持记录完整日期并显示已过年数

**Architecture:** 修改数据模型支持 YYYY-MM-DD 格式，通过 UnmarshalJSON 实现旧数据自动迁移，更新表单和显示界面

**Tech Stack:** Go 1.21, HTML templates

---

## Chunk 1: 后端数据模型修改

### Task 1: 修改 models.go 数据模型

**Files:**
- Modify: `models.go`

- [ ] **Step 1: 添加 calculateYearsSince 函数**

在 `models.go` 文件末尾添加年数计算函数：

```go
// calculateYearsSince 计算从原始日期到今年已经过去的年数
func calculateYearsSince(originalDate time.Time) int {
	now := time.Now()
	years := now.Year() - originalDate.Year()

	// 处理今年纪念日，需要考虑闰年2月29日的情况
	month := originalDate.Month()
	day := originalDate.Day()

	// 创建今年纪念日的日期
	thisYearOccurrence := time.Date(now.Year(), month, day, 0, 0, 0, 0, time.Local)

	// 如果今年纪念日还没到，年数减1
	if thisYearOccurrence.After(now) {
		years--
	}
	return years
}
```

- [ ] **Step 2: 添加 UnmarshalJSON 方法**

在 `Anniversary` 结构体定义后添加：

```go
// UnmarshalJSON 实现自定义 JSON 解析，支持旧格式数据迁移
func (a *Anniversary) UnmarshalJSON(data []byte) error {
	type Alias Anniversary
	tmp := struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	// 检测日期格式并迁移
	if len(tmp.Date) == 5 && strings.Contains(tmp.Date, "-") {
		// MM-DD 格式，自动补今年份
		currentYear := time.Now().Year()
		a.Date = fmt.Sprintf("%d-%s", currentYear, tmp.Date)
	} else {
		a.Date = tmp.Date
	}
	return nil
}
```

- [ ] **Step 3: 更新 GetAnniversaryView 函数**

替换整个 `GetAnniversaryView` 函数：

```go
// GetAnniversaryView 计算纪念日的视图数据
func GetAnniversaryView(anniversaries []Anniversary) []AnniversaryView {
	var views []AnniversaryView
	now := time.Now()

	for _, a := range anniversaries {
		// 解析日期 - 新格式 YYYY-MM-DD
		date, err := time.Parse("2006-01-02", a.Date)
		if err != nil {
			continue
		}

		// 计算下次纪念日
		nextOccurrence := time.Date(now.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		if nextOccurrence.Before(now) {
			nextOccurrence = time.Date(now.Year()+1, date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		}

		// 计算距离天数
		daysUntil := int(time.Until(nextOccurrence).Hours() / 24)

		// 计算年数
		yearsSince := calculateYearsSince(date)

		views = append(views, AnniversaryView{
			Anniversary:    a,
			DaysUntil:      daysUntil,
			IsUpcoming:     daysUntil <= 7 && daysUntil > 0,
			NextOccurrence: nextOccurrence,
			YearsSince:     yearsSince,
		})
	}

	return views
}
```

- [ ] **Step 4: 更新 AnniversaryView 结构体**

在 `AnniversaryView` 结构体中添加 `YearsSince` 字段：

```go
// AnniversaryView 纪念日视图（包含计算字段）
type AnniversaryView struct {
	Anniversary
	DaysUntil      int       `json:"days_until"`
	IsUpcoming     bool      `json:"is_upcoming"`
	NextOccurrence time.Time `json:"next_occurrence"`
	YearsSince     int       `json:"years_since"` // 新增：已过去的年数
}
```

- [ ] **Step 5: 验证并更新 imports**

确认 `models.go` 顶部 import 块包含以下所有包：

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)
```

如果有缺失的包，请添加到 import 块中。

---

## Chunk 2: 后端处理器修改

### Task 2: 修改 main.go 处理器

**Files:**
- Modify: `main.go`

- [ ] **Step 1: 修改 addHandler POST 处理**

找到 `addHandler` 函数中的 POST 处理部分（约第 136-165 行），修改为：

```go
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

	// 组合完整日期
	date := fmt.Sprintf("%s-%s-%s", year, month, day)
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
```

- [ ] **Step 2: 修改 editHandler GET 处理**

找到 `editHandler` 函数中的 GET 处理部分（约第 196-226 行），修改日期解析和数据结构：

```go
if r.Method == "GET" {
	// 解析日期 - 新格式 YYYY-MM-DD
	parts := strings.Split(ann.Date, "-")
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
```

- [ ] **Step 3: 修改 editHandler POST 处理**

找到 POST 处理部分（约第 229-260 行），修改为：

```go
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

	// 组合完整日期
	date := fmt.Sprintf("%s-%s-%s", year, month, day)
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
```

- [ ] **Step 4: 确认 main.go imports**

确保 imports 包含 `fmt`：

```go
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
```

---

## Chunk 3: 前端表单修改

### Task 3: 修改 add.html 添加年份选择器

**Files:**
- Modify: `templates/add.html`

- [ ] **Step 1: 添加年份选择器 HTML**

找到第 66-102 行的日期选择区域，在月份选择器之前添加年份选择器：

将：
```html
<div class="form-row">
    <div class="form-group form-group-month">
```

改为：
```html
<div class="form-row">
    <div class="form-group form-group-year">
        <label for="year" class="form-label">
            <span class="label-text">年份</span>
            <span class="label-badge label-required">必填</span>
        </label>
        <div class="select-wrapper">
            <select id="year" name="year" required class="form-select">
                <option value="">选择年份</option>
            </select>
        </div>
    </div>

    <div class="form-group form-group-month">
```

- [ ] **Step 2: 添加年份生成 JavaScript**

在现有 `<script>` 标签内（约第 134 行后），添加年份生成逻辑：

```javascript
// 年份选择器
const yearSelect = document.getElementById('year');

function initYearSelect() {
    const currentYear = new Date().getFullYear();
    const maxYear = currentYear + 1;
    const minYear = 1900;

    for (let year = maxYear; year >= minYear; year--) {
        const option = document.createElement('option');
        option.value = year.toString();
        option.textContent = year + '年';
        yearSelect.appendChild(option);
    }
}

initYearSelect();
```

### Task 4: 修改 edit.html 添加年份选择器

**Files:**
- Modify: `templates/edit.html`

- [ ] **Step 1: 添加年份选择器 HTML**

找到第 67-101 行的日期选择区域，添加年份选择器：

将：
```html
<div class="form-row">
    <div class="form-group form-group-month">
```

改为：
```html
<div class="form-row">
    <div class="form-group form-group-year">
        <label for="year" class="form-label">
            <span class="label-text">年份</span>
            <span class="label-badge label-required">必填</span>
        </label>
        <div class="select-wrapper">
            <select id="year" name="year" required class="form-select">
            </select>
        </div>
    </div>

    <div class="form-group form-group-month">
```

- [ ] **Step 2: 更新 JavaScript 初始化**

修改现有 `<script>` 部分（约第 133-162 行），添加年份处理：

```javascript
<script>
    const yearSelect = document.getElementById('year');
    const monthSelect = document.getElementById('month');
    const daySelect = document.getElementById('day');
    const currentYear = "{{.Year}}";
    const currentDay = "{{.Day}}";

    const daysInMonth = {
        '01': 31, '02': 29, '03': 31, '04': 30,
        '05': 31, '06': 30, '07': 31, '08': 31,
        '09': 30, '10': 31, '11': 30, '12': 31
    };

    // 初始化年份选择器
    function initYearSelect() {
        const thisYear = new Date().getFullYear();
        const maxYear = thisYear + 1;
        const minYear = 1900;

        for (let year = maxYear; year >= minYear; year--) {
            const option = document.createElement('option');
            option.value = year.toString();
            option.textContent = year + '年';
            if (option.value === currentYear) {
                option.selected = true;
            }
            yearSelect.appendChild(option);
        }
    }

    function updateDays() {
        const month = monthSelect.value;
        const days = daysInMonth[month] || 31;

        daySelect.innerHTML = '';
        for (let i = 1; i <= days; i++) {
            const option = document.createElement('option');
            option.value = i.toString().padStart(2, '0');
            option.textContent = i + '日';
            if (option.value === currentDay) {
                option.selected = true;
            }
            daySelect.appendChild(option);
        }
    }

    initYearSelect();
    monthSelect.addEventListener('change', updateDays);
    updateDays();
</script>
```

---

## Chunk 4: 列表显示修改

### Task 5: 修改 index.html 显示年数

**Files:**
- Modify: `templates/index.html`

- [ ] **Step 1: 更新日期显示格式**

找到第 63 行的日期显示部分：

```html
<span>每年 {{.Date}}</span>
```

改为：
```html
<span>{{.NextOccurrence.Format "2006年01月02日"}}</span>
```

- [ ] **Step 2: 添加年数显示**

找到第 65-68 行：

```html
<span class="meta-divider">·</span>
<span class="meta-item">
    <span class="meta-highlight">{{.NextOccurrence.Format "2006年1月2日"}}</span>
</span>
```

改为：
```html
<span class="meta-divider">·</span>
<span class="meta-item">
    <span class="meta-highlight">第 {{.YearsSince}} 年</span>
</span>
```

---

## Chunk 5: 测试验证

### Task 6: 编译和测试

**Files:**
- None (验证步骤)

- [ ] **Step 1: 编译项目**

Run: `go build -o anniversary.exe`

Expected: 编译成功，无错误

- [ ] **Step 2: 启动服务并手动测试**

Run: `./anniversary.exe`

Expected: 服务启动在 http://localhost:8080

测试项目：
1. 访问首页，检查现有数据显示正常（日期自动补今年份）
2. 添加新纪念日，选择完整年月日
3. 编辑现有纪念日，检查年份预选正确
4. 检查列表显示年数正确

- [ ] **Step 3: 验证数据格式**

检查 `data/anniversaries.json` 文件，确认日期格式为 `YYYY-MM-DD`

---

## 文件修改摘要

| 文件 | 修改类型 | 主要变更 |
|------|----------|----------|
| models.go | 修改 | 添加 UnmarshalJSON、calculateYearsSince，更新 GetAnniversaryView，AnniversaryView 添加 YearsSince |
| main.go | 修改 | addHandler 和 editHandler 添加 year 参数处理 |
| templates/add.html | 修改 | 添加年份选择器 |
| templates/edit.html | 修改 | 添加年份选择器，预选现有年份 |
| templates/index.html | 修改 | 显示完整日期和年数 |
