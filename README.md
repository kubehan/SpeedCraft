# ⚡ 速创社 (SpeedCraft)

> 专业 DevOps & 云原生架构咨询门户 · 把创意快速变成产品

一站式个人品牌门户网站，专为资深 DevOps / SRE 工程师设计。展示技术服务、开源作品、博客文章，并自带管理后台，帮助自由职业者获客变现。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+, net/http, html/template |
| 数据库 | SQLite (via modernc.org/sqlite, 纯 Go, 无 CGO) |
| 前端 | Tailwind CSS (CDN), Alpine.js 3.x |
| Markdown | goldmark (GFM 扩展) |
| 部署 | Docker, docker-compose, GitHub Actions |

## 功能特性

### 前台页面
- **首页** — Hero 大屏、核心技能、服务预览、项目统计
- **服务** — 18 个技术服务详情 + 定价 (全栈开发/云原生/K8s/DevOps 等)
- **案例** — 过往项目展示，含技术栈/客户/成果
- **博客** — Markdown/HTML 内容发布，阅读计数器
- **开源** — GitHub 开源项目 showcase，Star/语言/许可证
- **关于** — 个人介绍 + 在线联系表单 (含 WeChat/邮件通知)

### 管理后台
- **独立登录** — 专属登录页，Secure Cookie 会话
- **仪表盘** — 统计数据总览
- **服务管理** — CRUD + 排序 + 发布控制
- **文章管理** — Markdown 实时预览 + 草稿/发布切换
- **项目管理** — 案例增删改
- **开源项目管理** — Featured 标记
- **导航管理** — 菜单项排序/隐藏
- **技能管理** — 技术栈项管理
- **站点设置** — Hero 文案、统计数字、关于我 HTML、SEO、SMTP/Webhook 配置
- **文件上传** — 拖拽上传 + 文件画廊 (复制 URL/Markdown)
- **留言管理** — 访客留言列表

### 通知集成
- 访客留言 → WeChat Work 机器人 webhook
- 访客留言 → SMTP 邮件通知 (可选)
- 异步发送，不阻塞 API 响应

## 快速开始

### 本地开发

```bash
# 克隆
git clone <repo-url> && cd speedcraft

# 编译运行 (自动初始化 SQLite + 种子数据)
go run .    # 或 go build && ./speedcraft

# 打开浏览器
open http://localhost:8080
```

默认管理员: `admin` / `admin888`

通过环境变量自定义密码:

```bash
ADMIN_PWD=mysecret go run .
```

### Docker 部署

```bash
# 构建并启动
docker compose up -d

# 自定义管理员密码
ADMIN_PWD=mysecret docker compose up -d

# 查看日志
docker compose logs -f
```

### 生产部署

1. 配置域名反代 (Nginx/Caddy) 到 `localhost:8080`
2. 设置环境变量: `ADMIN_PWD`, `SITE_URL`
3. 进后台 → 站点设置 → 配置 SMTP / WeChat Webhook
4. (可选) GitHub Actions 自动构建 Docker 镜像推送 GHCR

## 配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `PORT` | `8080` | 服务端口 |
| `DB_PATH` | `data/speedcraft.db` | SQLite 数据库路径 |
| `SITE_URL` | `http://localhost:8080` | 站点 URL (影响 SEO/通知) |
| `ADMIN_PWD` | `admin888` | 管理员登录密码 |

## 目录结构

```
├── main.go                 # 入口 + 路由注册 + 安全中间件
├── config/config.go        # 配置加载 (环境变量)
├── database/db.go          # SQLite 初始化 + 自动迁移 + 种子数据
├── handlers/
│   ├── helpers.go          # 模板引擎 + FuncMap + render/respondJSON
│   ├── admin.go            # 登录/登出 + Session + AdminMiddleware + 仪表盘/消息
│   ├── admin_manage.go     # CRUD: 服务/文章/项目/开源/导航/技能/设置/上传/预览
│   ├── home.go             # 首页
│   ├── services.go         # 服务页 + 开源页
│   ├── blog.go             # 博客列表 + 详情 (goldmark 渲染)
│   └── contact.go          # 关于页 + 留言 API + 通知
├── models/models.go        # 数据模型 + CRUD 方法
├── templates/
│   ├── base.html           # 公共布局
│   ├── admin/
│   │   ├── admin_base.html # 管理后台布局 (深色侧边栏)
│   │   ├── login.html      # 独立登录页
│   │   ├── dashboard.html  # 仪表盘
│   │   ├── services.html   # 服务列表
│   │   ├── service_form.html
│   │   ├── posts.html      # 文章列表
│   │   ├── post_form.html  # 写文章 (含 Alpine 编辑器)
│   │   ├── projects.html   # 项目列表
│   │   ├── project_form.html
│   │   ├── opensource.html
│   │   ├── opensource_form.html
│   │   ├── navigation.html
│   │   ├── skills.html
│   │   ├── settings.html
│   │   ├── messages.html
│   │   └── upload.html     # 文件上传 (拖拽 + 画廊)
│   ├── index.html           # 首页
│   ├── services.html        # 服务展示页
│   ├── portfolio.html       # 案例展示页
│   ├── blog.html            # 博客列表
│   ├── blog_post.html       # 博文详情
│   ├── opensource.html      # 开源项目页
│   └── about.html           # 关于页 + 联系表单
├── static/                  # 静态资源
│   └── uploads/             # 上传文件存储
├── scripts/
│   └── seed.sql             # 手动种子 SQL (仅参考)
├── Dockerfile               # 多阶段构建 (~15MB)
├── docker-compose.yml       # 一键部署
└── .github/workflows/       # CI/CD
```

## 数据库

第一次启动自动创建 SQLite 数据库并填充种子数据:
- 18 个技术服务 + 定价
- 4 篇博客文章 (Markdown)
- 4 个客户案例
- 5 个开源项目
- 10 个技能项
- 6 个导航菜单项
- 30+ 站点设置

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 首页 |
| GET | `/services` | 服务列表 |
| GET | `/portfolio` | 案例列表 |
| GET | `/opensource` | 开源项目 |
| GET | `/blog` | 博客列表 |
| GET | `/blog/{slug}` | 博文详情 |
| GET | `/about` | 关于页 |
| POST | `/api/message` | 提交留言 |
| POST | `/admin/preview` | Markdown 预览 (需登录) |
| POST | `/admin/upload` | 文件上传 (需登录) |

## License

MIT
