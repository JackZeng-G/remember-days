<div align="center">

# 🕯️ Remember Days

**记录生命中每一个珍贵的时刻**

一个温馨优雅的纪念日管理 Web 应用

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPL--3.0-d4846a?style=flat-square)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Alpine-0db7ed?style=flat-square&logo=docker)](Dockerfile)

</div>

---

## ✨ 特性

- 温馨优雅的 UI 设计，时间线卡片布局
- 倒计时提醒，临近纪念日自动高亮
- CSRF 防护 · XSS 防御 · 数据文件加密存储
- 单二进制部署，Docker 一键启动

## 🚀 快速开始

```bash
go run .
```

打开 `http://localhost:8080` 即可使用。

## 🐳 Docker 部署

```bash
# 编译 + 构建 + 启动
GOOS=linux GOARCH=amd64 go build -o build/remember .
docker compose up -d --build
```

**运维**

```bash
docker compose logs -f         # 日志
docker compose restart         # 重启
docker compose down            # 停止
```

数据持久化在 `./data` 目录。

## ⚙️ 配置

| 变量 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `PORT` | 服务端口 | `8080` |
| `DATA_DIR` | 数据目录 | `data` |
| `CSRF_KEY` | CSRF 密钥 | ⚠️ 生产环境必须修改 |

## 📡 API

| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| `GET` | `/api/reminders` | 获取所有纪念日 |
| `GET` | `/api/status` | 服务状态 |

## 🛠 开发

```bash
go test ./... -v          # 运行测试
go test -cover ./...      # 覆盖率
```

## 📁 项目结构

```
├── main.go               # 入口
├── internal/             # 后端（Go 语言约定）
│   ├── config/           #   配置
│   ├── handler/          #   路由与中间件
│   ├── model/            #   数据模型
│   ├── service/          #   业务逻辑
│   └── store/            #   存储层
├── web/                  # 前端
│   ├── templates/        #   HTML 模板
│   └── static/           #   样式
├── Dockerfile
└── docker-compose.yml
```

## 📄 许可证

[GPL-3.0](LICENSE) © Remember Days
