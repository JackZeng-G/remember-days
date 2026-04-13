# 项目：记得

# 纪念日提醒服务

一个温馨优雅的纪念日管理 Web 应用，帮助您记录和追踪生命中每一个珍贵的时刻。

## 功能特性

- **纪念日管理** - 添加、编辑、删除纪念日事件
- **智能计算** - 自动计算距离下一个纪念日的天数
- **时光统计** - 显示已经相伴多少天、第几个纪念日
- **视觉提醒** - 即将到来（7天内）和当日的纪念日特殊高亮显示
- **JSON API** - 提供 RESTful API 接口，方便集成
- **优雅设计** - 精心设计的 UI，温暖的视觉风格

## 技术栈

- **Go 1.21+** - 后端服务
- **HTML Templates** - 服务端渲染
- **CSS3** - 现代 CSS 变量和动画
- **JSON Storage** - 本地文件存储，无需数据库

## 快速开始

### 编译运行

```bash
# 克隆项目
git clone <repository-url>

# 进入项目目录
cd 纪念日

# 编译
go build -o anniversary.exe

# 运行
./anniversary.exe
```

服务启动后，访问 http://localhost:8080

### 直接运行

```bash
go run .
```

## 项目结构

```
纪念日/
├── main.go          # 主程序入口和路由处理
├── models.go        # 数据模型和业务逻辑
├── storage.go       # 数据持久化
├── go.mod           # Go 模块定义
├── go.sum           # 依赖版本锁定
├── templates/       # HTML 模板
│   ├── index.html   # 主页模板
│   ├── add.html     # 添加页面
│   └── edit.html    # 编辑页面
├── static/          # 静态资源
│   └── style.css    # 样式文件
└── data/            # 数据存储目录
    └── anniversaries.json  # 纪念日数据
```

## API 接口

| 路径               | 方法     | 说明                    |
| ------------------ | -------- | ----------------------- |
| `/`              | GET      | 主页，显示所有纪念日    |
| `/add`           | GET/POST | 添加纪念日页面          |
| `/edit/{id}`     | GET/POST | 编辑纪念日页面          |
| `/delete/{id}`   | POST     | 删除纪念日              |
| `/api/reminders` | GET      | JSON 格式返回纪念日列表 |
| `/api/status`    | GET      | 服务状态检查            |

## 数据格式

纪念日数据存储在 `data/anniversaries.json`：

```json
{
  "anniversaries": [
    {
      "id": "abc12345",
      "name": "结婚纪念日",
      "date": "2020-05-20",
      "description": "美好的开始",
      "created_at": "2024-01-01 10:00:00"
    }
  ]
}
```

## 配置

在 `main.go` 中可修改以下配置：

```go
const (
    Port        = 8080            // 服务端口
    ServiceName = "纪念日提醒服务"  // 服务名称
)
```

## 界面预览

应用采用温暖优雅的设计风格：

- 渐变背景动画效果
- 时间线式纪念日展示
- 倒计时徽章突出显示
- 即将到来的纪念日高亮提醒
- 当日纪念日特殊动画效果

## 依赖

- [github.com/google/uuid](https://github.com/google/uuid) - UUID 生成

## 日志

运行日志记录在 `anniversary.log` 文件中，包括：

- 服务启动信息
- 纪念日添加/编辑/删除记录
- 错误信息

## LICENSE

GNU GPLv3
