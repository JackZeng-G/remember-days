# Remember Days - 纪念日管理应用

一个简洁的纪念日管理 Web 应用，帮助你记住每一个重要的日子。

## 项目结构

```
remember-days/
├── cmd/
│   └── server/
│       └── main.go          # 入口点
├── internal/
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── handler/
│   │   ├── handler.go       # HTTP 处理器
│   │   ├── middleware.go    # 中间件
│   │   └ template.go        # 模板渲染器
│   ├── model/
│   │   └ anniversary.go     # 数据模型
│   ├── service/
│   │   ├── anniversary.go   # 业务逻辑
│   │   ├── errors.go        # 业务错误
│   │   └ validation.go      # 输入验证
│   └── store/
│   │   ├── store.go         # 存储接口
│   │   ├── json_store.go    # JSON 文件实现
│   │   └ memory_store.go    # 内存实现（用于测试）
├── web/
│   ├── templates/           # HTML 模板
│   └ static/                # 静态资源
├── data/                    # 数据文件（运行时）
├── docs/                    # 文档
├── go.mod
├── go.sum
└── README.md
```

## 运行

```bash
# 直接运行
go run ./cmd/server

# 或编译后运行
go build -o anniversary.exe ./cmd/server
./anniversary.exe
```

访问 http://localhost:8080

## 配置

通过环境变量配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | 服务端口 | `8080` |
| `DATA_DIR` | 数据目录 | `data` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `CSRF_KEY` | CSRF 密钥 | （生产环境请修改） |

## API

- `GET /api/reminders` - 获取所有纪念日
- `GET /api/status` - 服务状态

## 测试

```bash
go test ./... -v
go test -cover ./...
```

## 安全特性

- CSRF 保护（所有表单）
- XSS 防护（模板自动转义）
- 文件权限 0600（数据文件）
- 日志输入清理

## 技术栈

- Go 1.21+
- chi router
- caarlos0/env 配置库
