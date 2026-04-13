# 纪念日增加年份字段设计文档

**日期**: 2026-03-15
**状态**: 设计中

## 1. 需求概述

为纪念日录入功能增加年份字段，使系统能够：
- 记录具体的完整日期（年月日）
- 计算并显示已过去的年数（如"第5年"）
- 支持历史事件纪念日（如结婚纪念日、首次发生事件等）

### 用户需求
- **必填年份**：所有纪念日必须包含年份
- **数据迁移**：现有数据自动补今年份
- **显示效果**：显示年数/周年数

## 2. 数据模型设计

### Anniversary 结构体
```go
type Anniversary struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Date        string `json:"date"` // 格式从 MM-DD 改为 YYYY-MM-DD
    Description string `json:"description"`
    CreatedAt   string `json:"created_at"`
}

// 实现 JSON Unmarshal 以支持旧格式数据迁移
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

### AnniversaryView 结构体（新增字段）

```go
type AnniversaryView struct {
    Anniversary
    DaysUntil      int       `json:"days_until"`
    IsUpcoming     bool      `json:"is_upcoming"`
    NextOccurrence time.Time `json:"next_occurrence"`
    YearsSince     int       `json:"years_since"`  // 新增：已过去的年数
}
```

### 数据迁移策略

通过实现 `UnmarshalJSON` 方法，在加载 JSON 数据时自动检测并迁移旧格式：

- 检测日期长度为 5 且包含 "-" 时，判定为 `MM-DD` 格式
- 自动拼接当前年份转为 `YYYY-MM-DD`
- 保存时统一使用 `YYYY-MM-DD` 格式

## 3. 表单界面设计

### 添加页面 (templates/add.html)
在月份选择器之前新增年份选择器：

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
                <!-- JavaScript 动态生成 1900-当前年份+1 -->
            </select>
        </div>
    </div>
    <!-- 现有的月份和日期选择器 -->
</div>
```

### 编辑页面 (templates/edit.html)

需要修改传递给模板的数据结构，增加 `Year` 字段：

```go
data := struct {
    ID          string
    Name        string
    Year        string  // 新增：年份
    Month       string
    Day         string
    Description string
}{
    ID:          ann.ID,
    Name:        ann.Name,
    Year:        year,   // 从完整日期中提取
    Month:       month,
    Day:         day,
    Description: ann.Description,
}
```

- 加载时从完整日期中提取年份并预选
- 支持修改年份

### 年份范围

1900年到当前年份+1（包含上下界）

### 表单提交方式

表单 POST 提交三个独立字段：`year`、`month`、`day`

后端接收后组合成 `YYYY-MM-DD` 格式存储。

## 4. 显示设计

### 列表显示格式
```
纪念日名称
YYYY年MM月DD日 · 第X年
还有 XX 天
```

### 显示示例
- "结婚纪念日 · 2020年05月20日 · 第5年"
- "小李生日 · 2000年03月15日 · 第25年"

### 年数计算逻辑（含闰年处理）

```go
func calculateYearsSince(originalDate time.Time) int {
    now := time.Now()
    years := now.Year() - originalDate.Year()

    // 处理今年纪念日，需要考虑闰年2月29日的情况
    month := originalDate.Month()
    day := originalDate.Day()

    // 创建今年纪念日的日期
    thisYearOccurrence := time.Date(now.Year(), month, day, 0, 0, 0, 0, time.Local)

    // 如果创建失败（如2月29日在非闰年），time.Date会自动调整为3月1日
    // 我们需要根据调整后的日期判断
    if thisYearOccurrence.After(now) {
        years--
    }
    return years
}
```

## 5. 后端处理逻辑

### 日期解析（兼容旧格式）

```go
func parseDate(dateStr string) (time.Time, error) {
    // 尝试完整格式 YYYY-MM-DD
    if t, err := time.Parse("2006-01-02", dateStr); err == nil {
        return t, nil
    }
    // 兼容旧格式 MM-DD，自动补今年份
    currentYear := time.Now().Year()
    return time.Parse("2006-01-02", fmt.Sprintf("%d-%s", currentYear, dateStr))
}
```

### GetAnniversaryView 修改

```go
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

### 需要修改的函数

| 文件 | 函数 | 修改内容 |
|------|------|----------|
| models.go | GetAnniversaryView() | 更新日期解析格式为 "2006-01-02"，增加 YearsSince 计算 |
| models.go | Anniversary.UnmarshalJSON() | 新增方法处理数据迁移 |
| main.go | addHandler() | 接收 year/month/day 参数，组合成 YYYY-MM-DD 格式 |
| main.go | editHandler() | 从 YYYY-MM-DD 中提取 year/month/day 传递给模板，接收编辑后的参数 |

### addHandler 详细修改

```go
// POST 处理
year := strings.TrimSpace(r.FormValue("year"))
month := strings.TrimSpace(r.FormValue("month"))
day := strings.TrimSpace(r.FormValue("day"))

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
```

### editHandler 详细修改

```go
// GET 处理 - 解析现有日期
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

// POST 处理 - 与 addHandler 相同的组合逻辑
```

### index.html 显示修改

将原有的 `每年 {{.Date}}` 替换为显示完整日期和年数：

```html
<!-- 原来显示：每年 05-20 -->
<!-- 修改为：2020年05月20日 -->
<div class="anniversary-date">{{.NextOccurrence.Format "2006年01月02日"}}</div>
<div class="anniversary-years">第 {{.YearsSince}} 年</div>
```

## 6. 测试计划

1. **数据迁移测试**
   - 验证现有 `MM-DD` 格式数据自动补今年份
   - 验证新数据保存为 `YYYY-MM-DD` 格式
   - 验证数据保存后 JSON 文件格式正确

2. **年数计算测试**
   - 验证今年纪念日已过的情况（年数 = 当前年 - 原始年）
   - 验证今年纪念日未到的情况（年数 = 当前年 - 原始年 - 1）
   - 验证今天就是纪念日的情况（年数 = 当前年 - 原始年）
   - 验证闰年2月29日的情况（2020年2月29日在2025年显示为第5年，下个纪念日为2028年2月29日）
   - 验证非闰年2月29日调整情况

3. **表单功能测试**
   - 添加新纪念日（含年份）
   - 编辑现有纪念日（修改年份）
   - 表单验证（年份必填）
   - 验证年份边界：1899年拒绝，1900年接受，当前年+2拒绝，当前年+1接受

4. **显示测试**
   - 验证列表中年数显示正确
   - 验证完整日期显示正确
   - 验证 API 返回的 JSON 包含 YearsSince 字段

## 7. API 兼容性

`/api/reminders` 端点的 `Date` 字段格式从 `MM-DD` 变更为 `YYYY-MM-DD`，这是破坏性变更。

### 影响

- 任何依赖此 API 的外部程序需要更新日期解析逻辑
- 新的 `YearsSince` 字段会被自动包含在响应中

### 缓解措施

由于这是个人项目，无外部依赖者，可以接受此破坏性变更。

## 8. 实施步骤

1. 修改 `models.go` 中的数据结构和计算逻辑
2. 修改 `main.go` 中的处理器函数
3. 修改 `templates/add.html` 和 `templates/edit.html`
4. 修改 `templates/index.html` 的显示部分
5. 测试数据迁移功能
6. 全面测试
