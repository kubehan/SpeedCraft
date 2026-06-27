package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(24 * time.Hour)

	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	log.Println("[DB] 数据库初始化完成")
	return nil
}

func migrate() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			phone TEXT DEFAULT '',
			company TEXT DEFAULT '',
			service_type TEXT DEFAULT '',
			budget TEXT DEFAULT '',
			message TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			notified INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			summary TEXT NOT NULL,
			content TEXT NOT NULL,
			category TEXT NOT NULL,
			tech_stack TEXT NOT NULL,
			image_url TEXT DEFAULT '',
			client_name TEXT DEFAULT '',
			client_url TEXT DEFAULT '',
			is_published INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS blog_posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			summary TEXT NOT NULL,
			content TEXT NOT NULL,
			content_type TEXT DEFAULT 'markdown',
			tags TEXT DEFAULT '',
			is_published INTEGER DEFAULT 0,
			views INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			icon TEXT DEFAULT '',
			description TEXT NOT NULL,
			features TEXT DEFAULT '',
			pricing TEXT DEFAULT '',
			sort_order INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS open_source_projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			url TEXT DEFAULT '',
			github_url TEXT DEFAULT '',
			stars INTEGER DEFAULT 0,
			language TEXT DEFAULT '',
			license_type TEXT DEFAULT '',
			is_featured INTEGER DEFAULT 0,
			sort_order INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS navigation_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			label TEXT NOT NULL,
			url TEXT NOT NULL,
			icon TEXT DEFAULT '',
			parent_id INTEGER DEFAULT 0,
			sort_order INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS site_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE NOT NULL,
			setting_value TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			icon TEXT DEFAULT '',
			name TEXT NOT NULL,
			level TEXT DEFAULT '',
			category TEXT DEFAULT '',
			sort_order INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			content TEXT NOT NULL,
			content_type TEXT DEFAULT 'markdown',
			render_mode TEXT DEFAULT 'embed',
			is_published INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slot TEXT NOT NULL,
			ad_type TEXT DEFAULT 'image',
			image_url TEXT DEFAULT '',
			link_url TEXT DEFAULT '',
			html_content TEXT DEFAULT '',
			alt_text TEXT DEFAULT '',
			start_at DATETIME,
			end_at DATETIME,
			sort_order INTEGER DEFAULT 0,
			click_count INTEGER DEFAULT 0,
			view_count INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS friendlinks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			logo_url TEXT DEFAULT '',
			description TEXT DEFAULT '',
			category TEXT DEFAULT '',
			status TEXT DEFAULT 'pending',
			submitter_email TEXT DEFAULT '',
			submitter_note TEXT DEFAULT '',
			sort_order INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range tables {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}

	// Add views column if missing (v2 migration)
	DB.Exec("ALTER TABLE blog_posts ADD COLUMN views INTEGER DEFAULT 0")
	// Add render_mode column if missing (v3 migration)
	DB.Exec("ALTER TABLE pages ADD COLUMN render_mode TEXT DEFAULT 'embed'")

	if err := seedDefaults(); err != nil {
		return fmt.Errorf("seed defaults: %w", err)
	}
	return nil
}

func seedDefaults() error {
	defaults := map[string]string{
		"site_name":             "速创社",
		"site_description":      "把创意快速变成产品 · 专业 DevOps & MVP 开发",
		"site_keywords":         "MVP开发,DevOps,云原生,快速原型,技术顾问,创业技术合伙人",
		"contact_email":         "hello@speedcraft.dev",
		"contact_wechat":        "SpeedCraft",
		"wechat_official":       "",
		"wechat_qr":             "",
		"icp_beian":             "",
		"icp_beian_url":         "https://beian.miit.gov.cn/",
		"police_beian":          "",
		"police_beian_url":      "",
		"copyright":             "",
		"footer_extra_html":     "",
		"default_project_cover": "",
		"wechat_webhook":        "",
		"smtp_host":             "",
		"smtp_port":             "587",
		"smtp_user":             "",
		"smtp_pass":             "",
		"notify_email":          "",
		"about_me":              `<div class="space-y-4"><p class="text-lg">9 年资深全栈运维开发，专注<strong>快速落地</strong>。我相信：好的技术不是炫技，而是帮你在最短时间内把想法变成可用的产品。</p><p>从 MVP 极速开发到云原生架构，从 CI/CD 自动化到安全合规，我帮你扫清技术障碍，让你的团队聚焦业务创新。</p></div>`,
		"hero_title":            "把创意快速变成产品",
		"hero_subtitle":         "MVP 极速开发 · 云原生架构 · DevOps 自动化 — 我帮初创企业和技术团队用最低成本、最快速度把想法落地。",
		"hero_badge":            "⚡ 专注 MVP 快速落地 · 9 年全栈经验",
		"stats_years":           "9+",
		"stats_projects":        "50+",
		"stats_uptime":          "99.9%",
		"stats_response":        "24h",
		"stats_years_label":     "年经验",
		"stats_projects_label":  "交付项目",
		"stats_uptime_label":    "系统可用性",
		"stats_response_label":  "响应时效",
		"philosophy_title":      "我们的理念",
		"philosophy_desc":       "快速 · 落地 · 交付 — 不做重理论，只做能跑的代码、能用的系统。你出想法，我出技术和执行力。",
		"theme_preset":          "indigo",
	}

	for key, value := range defaults {
		_, err := DB.Exec(
			`INSERT OR IGNORE INTO site_settings (setting_key, setting_value) VALUES (?, ?)`,
			key, value,
		)
		if err != nil {
			return err
		}
	}

	var count int
	DB.QueryRow("SELECT COUNT(*) FROM services").Scan(&count)
	if count == 0 {
		services := []struct {
			title, icon, desc, features, pricing string
			order int
		}{
			{"MVP 极速开发", "🚀", "从idea到可演示的MVP，最快3天交付。适合创业团队验证想法、融早期融资。", "需求梳理与原型设计\n全栈快速开发\n敏捷迭代 3天/Sprint\nCI/CD 自动部署\n用户反馈闭环", "¥5,000/起", 1},
			{"云架构设计", "☁️", "提供 AWS / Azure / 阿里云 全栈云架构设计与迁移方案，确保高可用、高安全、低成本。", "混合云架构设计\n云原生改造\n成本优化分析\n安全合规审计\n容灾与备份方案", "¥3,000/起", 2},
			{"CI/CD 流水线", "🔧", "定制企业级 CI/CD 流水线，从代码提交到生产发布全自动化，提升交付效率 10 倍。", "GitLab CI / Jenkins / GitHub Actions\n多环境自动部署\n质量门禁集成\n灰度发布 & 回滚\n流水线可视化", "¥2,000/起", 3},
			{"Kubernetes 运维", "🐳", "K8s 集群搭建、运维、排障，让您的容器化应用稳定运行。", "集群搭建 & 升级\n监控 & 告警\n弹性伸缩策略\n服务网格 (Istio)\n安全策略 (OPA)", "¥4,000/起", 4},
			{"监控体系建设", "📊", "Prometheus + Grafana + ELK 全栈监控，从基础设施到业务层全覆盖。", "Prometheus 生态搭建\nGrafana 看板定制\n日志采集 & 分析\n链路追踪 (OpenTelemetry)\n告警规则 & 值班", "¥2,500/起", 5},
			{"基础设施即代码", "📦", "Terraform / Pulumi / Ansible 实现基础设施全量代码化管理，环境可追溯、可复现。", "Terraform 多云编排\nAnsible 配置管理\n不可变基础设施\nGitOps 工作流\n合规即代码", "¥3,000/起", 6},
			{"DevOps 咨询与培训", "🎓", "团队 DevOps 成熟度评估、转型路线规划、技术培训，助力企业高效交付。", "现状评估 & 改进方案\n工具链选型 & 落地\n团队培训 & 工作坊\nSRE 体系建设\n运维自动化方案", "¥5,000/天", 7},
			{"代码开发", "💻", "全栈开发能力，Go / Python / Node.js / React，从后端到前端全覆盖。", "业务系统开发\nAPI 设计与开发\n第三方集成\n性能优化\n代码审查", "¥500/小时", 8},
			{"系统设计", "🏗️", "高可用、高并发系统架构设计，确保系统可扩展、可维护。", "系统架构设计\n技术选型评估\n容量规划\n性能评估\n技术方案文档", "¥3,000/起", 9},
			{"方案设计", "📋", "从业务需求到技术方案的完整转换，输出可落地的详细设计方案。", "需求分析与梳理\n技术方案编写\n架构评审\n实施路线图\n风险评估", "¥2,000/起", 10},
			{"信创适配", "🇨🇳", "国产化信创适配改造，支持鲲鹏、飞腾、麒麟、达梦等国产平台。", "信创环境评估\n应用适配改造\n国产数据库迁移\n性能调优\n兼容性测试", "¥5,000/起", 11},
			{"安全等保 & 基线核查", "🔒", "等保2.0合规咨询、安全基线核查、安全加固，帮助企业通过等保测评。", "等保差距分析\n安全整改方案\n基线核查与加固\n渗透测试\n安全应急响应", "¥4,000/起", 12},
			{"服务上云 & 托管维护", "☁️", "传统应用上云迁移、云原生改造、长期托管运维，让您专注于业务。", "上云评估与规划\n应用迁移实施\n7x24 监控告警\n定期巡检报告\n故障应急处理", "¥3,000/月", 13},
			{"脚本开发", "📜", "自动化脚本编写，批量运维、数据处理、定时任务，解决重复性工作。", "Shell/Python/Go 脚本\n批量运维自动化\n数据处理 & 清洗\n定时任务编排\n日志分析脚本", "¥500/起", 14},
			{"服务迁移", "📤", "服务器迁移、数据库迁移、应用迁移，无忧切换到新环境。", "跨云迁移方案\n数据库迁移 & 同步\n应用零停机迁移\n数据校验 & 回滚\n性能压测验证", "¥800/起", 15},
			{"服务器维护", "🛠️", "Linux 服务器日常维护、安全加固、性能优化，保障系统稳定运行。", "系统安全加固\n内核参数优化\n日志轮转配置\n定期备份策略\n故障排查 & 恢复", "¥300/次", 16},
			{"域名 & SSL 配置", "🌐", "域名解析、CDN 加速、SSL 证书配置与管理，一站式搞定。", "DNS 解析配置\nCDN 加速接入\nSSL 证书申请 & 续期\nHTTPS 强制跳转\n泛域名接入", "¥200/次", 17},
			{"Docker 容器化", "🐋", "现有应用容器化改造，提供 Dockerfile、docker-compose、K8s 部署配置。", "Dockerfile 编写\ndocker-compose 编排\n镜像构建 & 优化\n私有仓库搭建\nCI/CD 集成", "¥500/起", 18},
		}
		for _, s := range services {
			DB.Exec("INSERT INTO services (title, icon, description, features, pricing, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
				s.title, s.icon, s.desc, s.features, s.pricing, s.order)
		}
	}

	DB.QueryRow("SELECT COUNT(*) FROM navigation_items").Scan(&count)
	if count == 0 {
		navs := []struct{ label, url, icon string; order int }{
			{"首页", "/", "🏠", 1},
			{"服务", "/services", "🔧", 2},
			{"案例", "/portfolio", "📁", 3},
			{"开源", "/opensource", "⭐", 4},
			{"博客", "/blog", "📝", 5},
			{"关于", "/about", "👤", 6},
		}
		for _, n := range navs {
			DB.Exec("INSERT INTO navigation_items (label, url, icon, sort_order, is_published) VALUES (?, ?, ?, ?, 1)",
				n.label, n.url, n.icon, n.order)
		}
	}

	DB.QueryRow("SELECT COUNT(*) FROM skills").Scan(&count)
	if count == 0 {
		skills := []struct{ icon, name, level, category string; order int }{
			{"☁️", "AWS / Azure / 阿里云", "Expert", "Cloud", 1},
			{"🐳", "Docker / Kubernetes", "Expert", "Container", 2},
			{"📦", "Terraform / Ansible", "Expert", "IaC", 3},
			{"🔧", "Jenkins / GitLab CI", "Expert", "CI/CD", 4},
			{"📊", "Prometheus / Grafana", "Expert", "Monitor", 5},
			{"💻", "Go / Python / Node.js", "Expert", "Code", 6},
			{"⚛️", "React / Vue / 小程序", "Advanced", "Frontend", 7},
			{"🗄️", "MySQL / PostgreSQL / Redis", "Expert", "Database", 8},
			{"🌐", "Istio / Envoy / Nginx", "Advanced", "Network", 9},
			{"📝", "ELK / Loki / OpenTelemetry", "Advanced", "Observability", 10},
		}
		for _, s := range skills {
			DB.Exec("INSERT INTO skills (icon, name, level, category, sort_order) VALUES (?, ?, ?, ?, ?)",
				s.icon, s.name, s.level, s.category, s.order)
		}
	}

	DB.QueryRow("SELECT COUNT(*) FROM blog_posts").Scan(&count)
	if count == 0 {
		posts := []struct {
			title, slug, summary, content, tags string
			published int
		}{
			{"9年运维经验总结：如何构建高可用架构", "high-availability-architecture",
				"本文总结了9年运维生涯中关于高可用架构设计的关键原则和实践经验。",
				"## 引言\n\n高可用架构是每个运维工程师的必修课。经过9年的实战积累，我想分享一些核心原则。\n\n## 1. 冗余设计\n\n消除单点故障是高可用的第一步。从网络层到应用层，每一层都需要冗余设计。\n\n## 2. 故障隔离\n\n使用熔断器、舱壁模式等技术防止故障级联扩散。\n\n## 3. 容量规划\n\n基于历史数据和业务预测进行容量规划，预留20%-30%的Buffer。\n\n## 4. 自动化运维\n\n所有重复性工作都应该自动化，减少人为失误。\n\n## 总结\n\n高可用不是一蹴而就的，需要持续迭代和改进。",
				"高可用,架构设计,SRE", 1},
			{"Kubernetes排障实战指南", "kubernetes-troubleshooting-guide",
				"整理了日常K8s运维中常见的故障场景和排障方法。",
				"## Pod异常排障\n\nCrashLoopBackOff、ImagePullBackOff、Pending状态的排查步骤...\n\n## 网络问题\n\nDNS解析异常、Service访问超时、Ingress配置错误的排查...\n\n## 存储故障\n\nPV/PVC绑定失败、存储性能问题的排查...",
				"Kubernetes,排障,运维", 1},
			{"Terraform最佳实践", "terraform-best-practices",
				"使用Terraform管理基础设施的最佳实践总结。",
				"## 项目结构\n\n推荐的分层目录结构，环境隔离策略...\n\n## 状态管理\n\n远程状态存储、锁机制、状态迁移...\n\n## 模块设计\n\n可复用模块的设计原则、版本管理...",
				"Terraform,IaC,最佳实践", 1},
			{"从零搭建MVP：创业者的技术选型指南", "mvp-tech-stack-guide",
				"如何用最低成本、最快速度搭建MVP？本文分享技术选型策略。",
				"## MVP 核心原则\n\n做最少的功能，验证最核心的假设。\n\n## 技术选型\n\n- 后端: Go / Node.js\n- 前端: React / Vue\n- 数据库: PostgreSQL\n- 部署: Docker + 云服务器\n\n## 推荐工具链\n\n...",
				"MVP,创业,技术选型", 1},
		}
		for _, p := range posts {
			DB.Exec("INSERT INTO blog_posts (title, slug, summary, content, content_type, tags, is_published) VALUES (?, ?, ?, ?, 'markdown', ?, ?)",
				p.title, p.slug, p.summary, p.content, p.tags, p.published)
		}
	}

	DB.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if count == 0 {
		projects := []struct {
			title, slug, summary, content, category, techStack, client string
			published int
		}{
			{"电商平台K8s迁移", "ecommerce-k8s-migration", "为某头部电商平台完成从传统VM架构到Kubernetes容器平台的全面迁移。", "<h2>项目背景</h2><p>该电商平台日活用户超过100万，原有VM架构面临扩容慢、资源利用率低等问题。</p><h2>成果</h2><p>部署效率提升10倍，资源利用率提高60%，3次大促零宕机。</p>", "Kubernetes", "Docker,Kubernetes,Helm,Terraform,Prometheus", "某电商平台", 1},
			{"CI/CD流水线重构", "cicd-pipeline-redesign", "为金融科技公司重构CI/CD流水线，实现多环境自动部署与安全合规门禁。", "<h2>解决方案</h2><p>GitLab CI + GitOps 工作流，多环境自动部署，安全扫描与合规检查集成。</p><h2>成果</h2><p>发布频率从周级降至日级，回滚时间从30分钟降至2分钟。</p>", "CI/CD", "GitLab CI,ArgoCD,Terraform", "某金融科技公司", 1},
			{"全链路监控平台建设", "full-stack-monitoring", "为物联网企业构建从基础设施到业务指标的全链路监控平台。", "<h2>方案</h2><p>Prometheus联邦集群 + Grafana统一看板 + ELK日志平台 + OpenTelemetry链路追踪</p>", "监控", "Prometheus,Grafana,ELK,OpenTelemetry", "某IoT企业", 1},
			{"SaaS平台MVP极速开发", "saas-mvp-development", "为创业团队在4周内完成SaaS平台MVP开发，快速验证市场。", "<h2>项目亮点</h2><p>4周从0到1交付完整MVP，支持用户注册、支付、管理后台等核心功能。</p>", "MVP开发", "Go,React,PostgreSQL,Docker", "某创业团队", 1},
		}
		for _, p := range projects {
			DB.Exec("INSERT INTO projects (title, slug, summary, content, category, tech_stack, client_name, is_published) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				p.title, p.slug, p.summary, p.content, p.category, p.techStack, p.client, p.published)
		}
	}

	DB.QueryRow("SELECT COUNT(*) FROM open_source_projects").Scan(&count)
	if count == 0 {
		projects := []struct {
			name, desc, url, github, lang, license string
			stars, featured, order int
		}{
			{"GoDeploy", "轻量级 Go 应用部署工具，支持 CI/CD 一键部署到服务器", "https://github.com/example/godeploy", "https://github.com/example/godeploy", "Go", "MIT", 128, 1, 1},
			{"K8sDash", "Kubernetes 集群可视化监控面板", "https://github.com/example/k8sdash", "https://github.com/example/k8sdash", "TypeScript", "Apache-2.0", 89, 1, 2},
			{"DevOpsKit", "一键搭建 DevOps 工具链的 CLI 工具集", "https://github.com/example/devopskit", "https://github.com/example/devopskit", "Python", "MIT", 256, 1, 3},
			{"CloudInit", "多云基础设施初始化工具", "https://github.com/example/cloudinit", "https://github.com/example/cloudinit", "Go", "MIT", 67, 0, 4},
			{"LogPilot", "轻量日志采集与分析工具，替代 Filebeat + Logstash", "https://github.com/example/logpilot", "https://github.com/example/logpilot", "Go", "Apache-2.0", 45, 0, 5},
		}
		for _, p := range projects {
			DB.Exec("INSERT INTO open_source_projects (name, description, url, github_url, stars, language, license_type, is_featured, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
				p.name, p.desc, p.url, p.github, p.lang, p.license, p.stars, p.featured, p.order)
		}
	}

	DB.QueryRow("SELECT COUNT(*) FROM tags").Scan(&count)
	if count == 0 {
		tagNames := []string{"高可用", "架构设计", "SRE", "Kubernetes", "排障", "运维", "Terraform", "IaC", "最佳实践", "MVP", "创业", "技术选型"}
		for _, name := range tagNames {
			slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			DB.Exec("INSERT INTO tags (name, slug) VALUES (?, ?)", name, slug)
		}
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
