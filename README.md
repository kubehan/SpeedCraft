# ⚡ 速创社 (SpeedCraft)

> **专业 DevOps & 云原生咨询门户** · 把创意快速变成产品

[![Build](https://github.com/kubehan/SpeedCraft/actions/workflows/deploy.yml/badge.svg)](https://github.com/kubehan/SpeedCraft/actions/workflows/deploy.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubehan/speedcraft)](https://hub.docker.com/r/kubehan/speedcraft)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

一站式个人品牌 / 工作室门户，专为资深技术从业者设计。展示服务、案例、开源、博客，自带强大后台 + 营销工具（广告位 / SEO / 多主题），帮助接单变现。

---

## ✨ 核心亮点

- **⚡ 单二进制**：纯 Go，无 CGO，~15MB Docker 镜像，启动 < 100ms
- **🎨 5 套主题**：靛蓝 / 翡翠 / 玫瑰 / 琥珀 / 石板，后台一键切换
- **📦 SQLite 单文件**：零依赖部署，自动迁移 + 种子数据
- **🚀 多架构镜像**：linux/amd64 + linux/arm64（树莓派/M 系列 Mac/ARM 云服务器全支持）
- **🔒 CSRF + Session**：双层防护，Cookie HttpOnly + 表单 Token
- **📊 SEO 全套**：sitemap.xml / robots.txt / JSON-LD / OG tags / canonical
- **💰 广告位系统**：8 个预设投放位置，支持图片 + AdSense，自带点击/曝光统计

---

## 🧱 技术栈

| 层级 | 技术 |
|------|------|
| **后端** | Go 1.22+ · net/http · html/template |
| **DB** | SQLite (modernc.org/sqlite, 纯 Go) |
| **前端** | Tailwind CSS (CDN) · Alpine.js 3.x · SortableJS |
| **Markdown** | goldmark (GFM / TaskList / Footnote) |
| **部署** | Docker · BuildKit 多架构 · GitHub Actions |

---

## 📋 功能清单

### 前台页面
| 模块 | 路径 | 说明 |
|------|------|------|
| 首页 | `/` | Hero + 统计 + 技能 + 服务预览 + CTA |
| 服务 | `/services` | 18 项技术服务 + 报价 + 合作流程 |
| 案例 | `/portfolio` | 过往项目展示 |
| 开源 | `/opensource` | GitHub 开源项目 + Featured 推荐 |
| 博客 | `/blog` · `/blog/{slug}` | Markdown 文章 + 阅读计数 + JSON-LD |
| 关于 | `/about` | 个人介绍 + 在线联系表单（WeChat/邮件通知） |
| **自定义页面** | `/page/{slug}` | 后台创建任意页面（嵌入 / 独立 HTML 两种模式） |
| **SEO** | `/sitemap.xml` · `/robots.txt` | 自动生成 |

### 管理后台（`/admin`）
- 仪表盘（数据总览）
- **服务管理** — CRUD + 行内发布切换 + 批量操作 + 分页
- **文章管理** — Markdown 编辑器 + 弹窗预览 + 标签选择器 + 图片选择器
- **页面管理** — 嵌入模式（Markdown/HTML 片段）+ 独立 HTML 模式（iframe 隔离，整页落地页）
- **项目管理** — 案例 CRUD + 封面图选择
- **开源项目** — GitHub 项目展示管理
- **导航管理** — 拖拽排序 + 手动保存 + 关联已有页面
- **技能管理** — 编辑 + 排序 + 显示/隐藏
- **标签管理** — 博客标签 CRUD
- **广告管理** — 8 个投放位置 + 图片/HTML 两种类型 + 有效期 + 点击/曝光统计
- **站点设置** — Hero / 统计 / 关于我 / SEO 关键词 / SMTP / WeChat Webhook / **主题切换**
- **文件管理** — 拖拽上传 + 文件浏览器 + 一键复制（链接 / Markdown / HTML）+ 搜索过滤
- **留言管理** — 搜索 + 详情弹窗 + 状态变更 + CSV 导出 + 测试通知

### 通知集成
- 访客提交联系表单 → 企业微信机器人 + SMTP 邮件并行通知
- 异步发送（goroutine），不阻塞 API 响应
- 后台一键「测试通知」验证配置

### 主题系统
| 主题 | 主色 | 适合 |
|------|------|------|
| 靛蓝（默认） | `#4f46e5` | 专业可靠 |
| 翡翠 | `#059669` | 自然清新 |
| 玫瑰 | `#e11d48` | 温暖活力 |
| 琥珀 | `#d97706` | 友好亲和 |
| 石板 | `#0f172a` | 极简黑白 |

CSS 变量驱动，保存设置后全站立即生效。

### 动效系统
- Hero 区**科技感动态背景**：流动网格 + 漂浮光球 + 扫描线 + 脉动点阵（纯 CSS，支持 `prefers-reduced-motion`）
- IntersectionObserver 滚动触发动画：`reveal` / `reveal-stagger` / `reveal-zoom`
- 卡片悬停微动效、按钮波纹反馈
- 移动端右侧抽屉式导航（带遮罩 + ESC 关闭 + 滚动锁定）

---

## 🚀 快速开始

### 方式 1：Docker（推荐）

```bash
docker run -d \
  --name speedcraft \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /data/speedcraft/data:/app/data \
  -v /data/speedcraft/static/uploads:/app/static/uploads \
  kubehan/speedcraft:latest
```

访问 `http://localhost:8080`，后台 `http://localhost:8080/admin`

**默认管理员**：`admin` / `admin888`（**生产环境务必修改**，见下文「安全」章节）

### 方式 2：docker-compose

```yaml
version: '3.8'
services:
  speedcraft:
    image: kubehan/speedcraft:latest
    container_name: speedcraft
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./data:/app/data
      - ./uploads:/app/static/uploads
    environment:
      - PORT=8080
      - DB_PATH=/app/data/speedcraft.db
```

### 方式 3：本地开发

```bash
git clone https://github.com/kubehan/SpeedCraft.git
cd SpeedCraft
go mod download
go run .
```

数据库 / 模板 / 静态资源全自动初始化，首次启动会写入种子数据（18 个服务、10 个技能、4 篇文章等）。

---

## ⚙️ 配置

通过环境变量配置：

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `PORT` | `8080` | HTTP 监听端口 |
| `DB_PATH` | `data/speedcraft.db` | SQLite 数据库路径 |
| `SITE_NAME` | `速创社` | 站点名（也可在后台设置中改） |
| `ADMIN_USER` | `admin` | 后台管理员用户名 |
| `ADMIN_PASS` | `admin888` | 后台管理员密码 |

**所有内容**（首页文案、服务列表、SMTP、Webhook、SEO 关键词、主题）都在**后台站点设置**里配置，无需重启。

---

## 🌐 Nginx 反代部署

```nginx
upstream speedcraft {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate     /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    client_max_body_size 20M;

    # 静态资源直接交给 nginx（更快）
    location /static/ {
        alias /data/speedcraft/static/;
        expires 7d;
        access_log off;
        add_header Cache-Control "public, immutable";
    }

    location / {
        proxy_pass http://speedcraft;
        proxy_http_version 1.1;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

获取 SSL 证书：
```bash
sudo certbot --nginx -d your-domain.com
```

---

## 🔒 安全

### 必做

1. **修改默认密码** — 通过环境变量 `ADMIN_PASS` 或后台修改
2. **绑定到 127.0.0.1** — Docker 部署用 `-p 127.0.0.1:8080:8080`，强制走 nginx
3. **启用 HTTPS** — 用 certbot 申请 Let's Encrypt 证书
4. **定期备份** — `data/speedcraft.db` 是全部数据

### 已内置防护

- ✅ CSRF Token（所有 POST 表单 + AJAX）
- ✅ Session 服务端存储（HttpOnly Cookie）
- ✅ SQL 参数化查询（防 SQL 注入）
- ✅ `html/template` 自动 XSS 转义
- ✅ 路径穿越防护（文件删除时校验 `..` / `/`）
- ✅ 文件上传白名单（仅 `static/uploads/` 目录）
- ✅ 安全响应头（X-Content-Type-Options / X-Frame-Options / X-XSS-Protection）

---

## 🛠 开发指南

### 项目结构

```
SpeedCraft/
├── main.go              # 入口 + 路由注册
├── config/              # 配置加载
├── database/db.go       # SQLite 初始化 + 迁移 + 种子数据
├── models/models.go     # 所有数据模型 + CRUD
├── handlers/            # HTTP handler
│   ├── home.go          # 首页
│   ├── services.go      # 服务/案例/页面
│   ├── blog.go          # 博客 + Markdown 渲染
│   ├── contact.go       # 联系表单 + 通知
│   ├── admin.go         # 后台登录/中间件/Dashboard
│   ├── admin_manage.go  # 后台所有 CRUD（最长，包含全部资源）
│   ├── seo.go           # sitemap.xml / robots.txt
│   └── helpers.go       # 模板渲染 + FuncMap + Session
├── templates/
│   ├── base.html        # 公共页面 layout
│   ├── index.html       # 首页
│   ├── services.html    # 服务
│   ├── portfolio.html   # 案例
│   ├── opensource.html  # 开源
│   ├── blog.html        # 博客列表
│   ├── blog_post.html   # 博客详情
│   ├── about.html       # 关于
│   ├── page.html        # 自定义页面（嵌入模式）
│   ├── page_standalone.html  # 自定义页面（独立 iframe 模式）
│   ├── ad_slot.html     # 广告位通用模板
│   └── admin/           # 后台所有模板
└── static/              # CSS / JS / 上传文件
    └── uploads/         # 用户上传文件（gitignored）
```

### 构建

```bash
# 本地编译
go build -o speedcraft .

# 多架构 Docker 镜像（本地需 buildx）
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t kubehan/speedcraft:latest \
  --push .
```

### 测试

```bash
go vet ./...
go test ./...
```

---

## 🤖 CI/CD

`.github/workflows/deploy.yml` 已配置好：

- **触发**：push 到 `main` 分支 / push tag `v*` / 手动触发
- **流程**：`go vet` → `go build` → 多架构 Docker 构建（amd64 + arm64）→ 推送 Docker Hub
- **缓存**：GHA cache 加速重复构建
- **Tag 策略**：`latest` / `v1.0.0` / `1.0` / `sha-xxxxxxx`

### 需配置的 Secrets

到 `https://github.com/kubehan/SpeedCraft/settings/secrets/actions` 添加：

| Secret | 说明 |
|--------|------|
| `DOCKERHUB_USERNAME` | Docker Hub 用户名（`kubehan`） |
| `DOCKERHUB_TOKEN` | Docker Hub Access Token（**不是登录密码**，到 `https://hub.docker.com/settings/security` 创建） |

---

## 📈 SEO 优化

- ✅ 每页动态 `meta description` / `keywords`
- ✅ 完整 Open Graph 标签（`og:title` / `og:description` / `og:image` / `og:type`）
- ✅ Twitter Card 支持
- ✅ 博客文章注入 JSON-LD `BlogPosting` 结构化数据（含 author / publisher / datePublished / keywords）
- ✅ Canonical URL 防重复内容
- ✅ `/sitemap.xml` 自动生成（含所有静态页 + 博客文章 + 自定义页面）
- ✅ `/robots.txt` 自动生成，屏蔽 `/admin/` 和 `/api/`
- ✅ 语义化 HTML（`<article>` / `<nav>` / `<main>` / `<footer>`）

---

## 💰 广告位使用

后台 → 广告管理 → 新建广告：

| 投放位置 | 说明 |
|----------|------|
| `home_hero` | 首页 Hero 下方 |
| `home_middle` | 首页板块之间 |
| `home_bottom` | 首页页脚上方 |
| `blog_top` | 博客列表顶部 |
| `blog_post_top` | 文章详情顶部 |
| `blog_post_bottom` | 文章详情底部 |
| `sidebar` | 侧边栏 |
| `global_top` | 全站顶部条幅 |

**两种广告类型**：
- **图片广告** — 上传图 + 跳转链接，自动通过 `/ad/click/{id}` 中转记录点击
- **HTML/脚本** — 可粘贴 Google AdSense、阿里妈妈、自定义 HTML

**有效期控制**：设置开始/结束时间，过期自动下线
**统计**：实时展示曝光数 / 点击数

---

## 🗺 Roadmap

- [x] 主题切换系统
- [x] 自定义页面（嵌入 / 独立两种模式）
- [x] 广告位管理
- [x] SEO 完整支持
- [x] 多架构 Docker 镜像
- [ ] 评论系统（Giscus / Disqus 集成）
- [ ] 多语言 i18n
- [ ] 数据导出 / 备份恢复
- [ ] WebHook 触发器（文章发布通知 / 留言通知扩展）
- [ ] 访问统计仪表盘（接入 Plausible / Umami）

---

## 📄 License

MIT © 2024 [kubehan](https://github.com/kubehan)

---

## 🙏 致谢

- [Alpine.js](https://alpinejs.dev/) — 轻量交互
- [Tailwind CSS](https://tailwindcss.com/) — 实用类 CSS
- [goldmark](https://github.com/yuin/goldmark) — Markdown 渲染
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) — 纯 Go SQLite
- [SortableJS](https://sortablejs.github.io/Sortable/) — 拖拽排序

---

**Star ⭐ 一下支持作者继续维护**
