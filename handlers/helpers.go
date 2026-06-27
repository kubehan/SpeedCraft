package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"speedcraft/config"
	"speedcraft/models"
)

type PageData struct {
	Title           string
	Site            *config.Config
	Data            interface{}
	Current         string
	CSRFToken       string
	MetaDescription string
	MetaKeywords    string
	OGType          string
	OGImage         string
	CanonicalURL    string
	JSONLD          string
}

type NavItem struct {
	Key   string
	Label string
	URL   string
	Icon  string
}

var hardcodedNav = []NavItem{
	{Key: "home", Label: "首页", URL: "/", Icon: "🏠"},
	{Key: "services", Label: "服务", URL: "/services", Icon: "🔧"},
	{Key: "portfolio", Label: "案例", URL: "/portfolio", Icon: "📁"},
	{Key: "opensource", Label: "开源", URL: "/opensource", Icon: "⭐"},
	{Key: "blog", Label: "博客", URL: "/blog", Icon: "📝"},
	{Key: "about", Label: "关于", URL: "/about", Icon: "👤"},
}

var baseTmpl *template.Template

func InitTemplates() error {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04")
		},
		"slice": func(s string, n int) string {
			runes := []rune(s)
			if len(runes) <= n {
				return s
			}
			return string(runes[:n]) + "..."
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"navItems": func() []NavItem {
			if items, err := models.GetPublishedNavigation(); err == nil && len(items) > 0 {
				result := make([]NavItem, len(items))
				for i, item := range items {
					result[i] = NavItem{Key: item.URL, Label: item.Label, URL: item.URL, Icon: item.Icon}
				}
				return result
			}
			return hardcodedNav
		},
		"listSkills": func() []models.Skill {
			if skills, err := models.GetPublishedSkills(); err == nil {
				return skills
			}
			return []models.Skill{}
		},
		"previewServices": func() []models.Service {
			services, err := models.GetPublishedServices()
			if err != nil {
				return []models.Service{}
			}
			if len(services) > 3 {
				return services[:3]
			}
			return services
		},
		"getSetting": func(key string) string {
			return models.GetSetting(key)
		},
		"friendLinks": func() []models.FriendLink {
			list, err := models.GetPublishedFriendLinks()
			if err != nil {
				return []models.FriendLink{}
			}
			return list
		},
		"socialAccounts": func() []models.SocialAccount {
			list, err := models.GetPublishedSocialAccounts()
			if err != nil {
				return []models.SocialAccount{}
			}
			return list
		},
		"socialPlatforms": func() []map[string]string {
			return []map[string]string{
				{"Key": "wechat_official", "Label": "微信公众号", "Icon": "📢", "ShortLabel": "公众号"},
				{"Key": "wechat_channel", "Label": "微信视频号", "Icon": "🎬", "ShortLabel": "视频号"},
				{"Key": "wechat_mini", "Label": "微信小程序", "Icon": "🧩", "ShortLabel": "小程序"},
				{"Key": "wechat_personal", "Label": "微信号", "Icon": "💬", "ShortLabel": "微信"},
				{"Key": "xiaohongshu", "Label": "小红书", "Icon": "📕", "ShortLabel": "小红书"},
				{"Key": "douyin", "Label": "抖音", "Icon": "🎵", "ShortLabel": "抖音"},
				{"Key": "kuaishou", "Label": "快手", "Icon": "⚡", "ShortLabel": "快手"},
				{"Key": "weibo", "Label": "微博", "Icon": "🐦", "ShortLabel": "微博"},
				{"Key": "bilibili", "Label": "B站", "Icon": "📺", "ShortLabel": "B站"},
				{"Key": "zhihu", "Label": "知乎", "Icon": "💡", "ShortLabel": "知乎"},
				{"Key": "github", "Label": "GitHub", "Icon": "🐙", "ShortLabel": "GitHub"},
				{"Key": "twitter", "Label": "Twitter / X", "Icon": "🐤", "ShortLabel": "X"},
				{"Key": "youtube", "Label": "YouTube", "Icon": "▶️", "ShortLabel": "YouTube"},
				{"Key": "qq_group", "Label": "QQ 群", "Icon": "🐧", "ShortLabel": "QQ群"},
				{"Key": "telegram", "Label": "Telegram", "Icon": "✈️", "ShortLabel": "TG"},
				{"Key": "custom", "Label": "自定义", "Icon": "🔗", "ShortLabel": "其他"},
			}
		},
		"platformIcon": func(key string) string {
			icons := map[string]string{
				"wechat_official": "📢", "wechat_channel": "🎬", "wechat_mini": "🧩", "wechat_personal": "💬",
				"xiaohongshu": "📕", "douyin": "🎵", "kuaishou": "⚡", "weibo": "🐦",
				"bilibili": "📺", "zhihu": "💡", "github": "🐙", "twitter": "🐤",
				"youtube": "▶️", "qq_group": "🐧", "telegram": "✈️", "custom": "🔗",
			}
			if v, ok := icons[key]; ok {
				return v
			}
			return "🔗"
		},
		"platformSVG": func(key string) template.HTML {
			// Recognizable platform marks — distinctive shapes/letters, not exact trademark reproductions
			svgs := map[string]string{
				// 微信 — chat bubble shape
				"wechat_official": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M8.5 6C4.91 6 2 8.46 2 11.5c0 1.74.96 3.29 2.46 4.31L4 18l2.16-1.18c.74.2 1.53.31 2.34.31.2 0 .39-.01.58-.02-.13-.42-.21-.85-.21-1.31 0-2.94 2.91-5.3 6.5-5.3.39 0 .77.03 1.14.09C16.05 7.42 12.6 6 8.5 6zM6 9.75a.75.75 0 110 1.5.75.75 0 010-1.5zm5 0a.75.75 0 110 1.5.75.75 0 010-1.5zm9.5 1.75c-2.99 0-5.5 2.02-5.5 4.5s2.51 4.5 5.5 4.5c.62 0 1.22-.09 1.78-.24L24 21l-.4-1.4C24.84 18.76 26 17.21 26 15.5l-.06-.41C25.5 12.92 23.21 11.5 20.5 11.5z" transform="scale(0.85) translate(2 2)"/></svg>`,

				// 视频号 — play triangle in chat bubble
				"wechat_channel": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M12 2C6.48 2 2 6.04 2 11c0 2.42 1.07 4.6 2.8 6.18L4 22l4.5-2.4c1.1.26 2.27.4 3.5.4 5.52 0 10-4.04 10-9s-4.48-9-10-9zm-2 12V8l5 3-5 3z"/></svg>`,

				// 小程序 — diamond
				"wechat_mini": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M12 2L2 12l10 10 10-10L12 2zm0 4l6 6-6 6-6-6 6-6z"/></svg>`,

				// 微信号 (personal) — speech bubble
				"wechat_personal": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zM8 11.5a1.5 1.5 0 110-3 1.5 1.5 0 010 3zm8 0a1.5 1.5 0 110-3 1.5 1.5 0 010 3z"/></svg>`,

				// 小红书 — letter mark in rounded square
				"xiaohongshu": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><rect x="2" y="2" width="20" height="20" rx="4" fill="currentColor" opacity="0.15"/><text x="12" y="16" font-size="11" font-weight="900" text-anchor="middle" fill="currentColor" font-family="-apple-system,sans-serif">小</text></svg>`,

				// 抖音 — musical note
				"douyin": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M12.5 2v12.5a3 3 0 11-3-3c.34 0 .67.06.97.17V8.55a6 6 0 105.03 5.95V8.2c.99.71 2.19 1.13 3.5 1.13V6.32a4.32 4.32 0 01-4.32-4.32H12.5z"/></svg>`,

				// 快手 — lightning bolt
				"kuaishou": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M13 2L4 14h7l-2 8 9-12h-7l2-8z"/></svg>`,

				// 微博 — letter W
				"weibo": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><rect x="2" y="2" width="20" height="20" rx="4" fill="currentColor" opacity="0.15"/><text x="12" y="16" font-size="11" font-weight="900" text-anchor="middle" fill="currentColor" font-family="-apple-system,sans-serif">微</text></svg>`,

				// B站 — TV with antenna
				"bilibili": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M7.5 2.5L9 4h6l1.5-1.5 1.4 1.4L16.4 5H19a2 2 0 012 2v11a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2.6L6.1 3.9l1.4-1.4zM7 9v6h2v-6H7zm8 0v6h2v-6h-2z"/></svg>`,

				// 知乎 — letter Z
				"zhihu": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><rect x="2" y="2" width="20" height="20" rx="4" fill="currentColor" opacity="0.15"/><text x="12" y="16" font-size="11" font-weight="900" text-anchor="middle" fill="currentColor" font-family="-apple-system,sans-serif">知</text></svg>`,

				// GitHub — Octocat-style head
				"github": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M12 2C6.48 2 2 6.48 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.87 1.52 2.34 1.07 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.94 0-1.1.39-1.99 1.03-2.69-.1-.25-.45-1.27.1-2.65 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.38.2 2.4.1 2.65.64.7 1.03 1.59 1.03 2.69 0 3.84-2.34 4.69-4.57 4.94.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12c0-5.52-4.48-10-10-10z"/></svg>`,

				// Twitter / X — letter X
				"twitter": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M18 2h3l-7.5 8.6L22 22h-7l-5.5-7.2L3 22H0l8-9.2L0 2h7l5 6.5L18 2z"/></svg>`,

				// YouTube — play in rounded rect
				"youtube": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M23 7.5a3 3 0 00-2.1-2.1C19 5 12 5 12 5s-7 0-8.9.4A3 3 0 001 7.5C.6 9.4.6 12 .6 12s0 2.6.4 4.5a3 3 0 002.1 2.1C5 19 12 19 12 19s7 0 8.9-.4a3 3 0 002.1-2.1c.4-1.9.4-4.5.4-4.5s0-2.6-.4-4.5zM9.6 15.5v-7l6 3.5-6 3.5z"/></svg>`,

				// QQ — penguin silhouette
				"qq_group": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M12 2C8.5 2 6 4.5 6 8c0 1.5.5 3 1.5 4-.5.5-2.5 2.5-2.5 5 0 .5.5 1 1 1h12c.5 0 1-.5 1-1 0-2.5-2-4.5-2.5-5 1-1 1.5-2.5 1.5-4 0-3.5-2.5-6-6-6zm-2 6c.6 0 1 .4 1 1s-.4 1-1 1-1-.4-1-1 .4-1 1-1zm4 0c.6 0 1 .4 1 1s-.4 1-1 1-1-.4-1-1 .4-1 1-1z"/></svg>`,

				// Telegram — paper plane
				"telegram": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M22 2L2 11l6 2 2 7 4-4 6 5 2-19zM10 14l8-7-10 8 2-1z"/></svg>`,

				// Custom — generic link
				"custom": `<svg viewBox="0 0 24 24" fill="currentColor" class="w-full h-full"><path d="M10.6 13.4a4 4 0 005.7 0l3.5-3.5a4 4 0 00-5.7-5.7l-1.4 1.4 1.4 1.4 1.4-1.4a2 2 0 012.9 2.9l-3.5 3.5a2 2 0 01-2.9 0l-1.4 1.4zm2.8-2.8a4 4 0 00-5.7 0l-3.5 3.5a4 4 0 005.7 5.7l1.4-1.4-1.4-1.4-1.4 1.4a2 2 0 01-2.9-2.9l3.5-3.5a2 2 0 012.9 0l1.4-1.4z"/></svg>`,
			}
			if v, ok := svgs[key]; ok {
				return template.HTML(v)
			}
			return template.HTML(svgs["custom"])
		},
		"platformLabel": func(key string) string {
			labels := map[string]string{
				"wechat_official": "微信公众号", "wechat_channel": "微信视频号", "wechat_mini": "微信小程序", "wechat_personal": "微信",
				"xiaohongshu": "小红书", "douyin": "抖音", "kuaishou": "快手", "weibo": "微博",
				"bilibili": "B站", "zhihu": "知乎", "github": "GitHub", "twitter": "Twitter",
				"youtube": "YouTube", "qq_group": "QQ群", "telegram": "Telegram", "custom": "自定义",
			}
			if v, ok := labels[key]; ok {
				return v
			}
			return key
		},
		"platformColor": func(key string) string {
			// Brand colors for each platform
			colors := map[string]string{
				"wechat_official": "#07c160", "wechat_channel": "#07c160", "wechat_mini": "#07c160", "wechat_personal": "#07c160",
				"xiaohongshu":     "#ff2442",
				"douyin":          "#000000",
				"kuaishou":        "#ff4906",
				"weibo":           "#e6162d",
				"bilibili":        "#fb7299",
				"zhihu":           "#0084ff",
				"github":          "#24292e",
				"twitter":         "#1da1f2",
				"youtube":         "#ff0000",
				"qq_group":        "#1296db",
				"telegram":        "#0088cc",
				"custom":          "#6366f1",
			}
			if v, ok := colors[key]; ok {
				return v
			}
			return "#6366f1"
		},
		"linkColor": func(i interface{}) string {
			// Tasteful palette that works on dark footer backgrounds
			palette := []string{
				"#60a5fa", // blue-400
				"#4ade80", // green-400
				"#f472b6", // pink-400
				"#fbbf24", // amber-400
				"#a78bfa", // violet-400
				"#22d3ee", // cyan-400
				"#fb923c", // orange-400
				"#34d399", // emerald-400
				"#f87171", // red-400
				"#c084fc", // purple-400
				"#facc15", // yellow-400
				"#2dd4bf", // teal-400
			}
			var idx int
			switch v := i.(type) {
			case int:
				idx = v
			case int64:
				idx = int(v)
			case string:
				// Deterministic hash by string content
				for _, c := range v {
					idx += int(c)
				}
			}
			if idx < 0 {
				idx = -idx
			}
			return palette[idx%len(palette)]
		},
		"adsBySlot": func(slot string) []models.Ad {
			ads, err := models.GetActiveAdsBySlot(slot)
			if err != nil {
				return []models.Ad{}
			}
			// Increment view count async (fire & forget)
			for _, a := range ads {
				go models.IncrementAdView(a.ID)
			}
			return ads
		},
		"adSlots": func() []map[string]string {
			return []map[string]string{
				{"Key": "home_hero", "Label": "首页 Hero 下方"},
				{"Key": "home_middle", "Label": "首页板块之间"},
				{"Key": "home_bottom", "Label": "首页页脚上方"},
				{"Key": "blog_top", "Label": "博客列表顶部"},
				{"Key": "blog_post_top", "Label": "文章详情顶部"},
				{"Key": "blog_post_bottom", "Label": "文章详情底部"},
				{"Key": "sidebar", "Label": "侧边栏"},
				{"Key": "global_top", "Label": "全站顶部条幅"},
			}
		},
		"themePresets": func() []map[string]string {
			return []map[string]string{
				{"Key": "indigo", "Label": "靛蓝", "Color": "#4f46e5", "Desc": "专业可靠"},
				{"Key": "emerald", "Label": "翡翠", "Color": "#059669", "Desc": "自然清新"},
				{"Key": "rose", "Label": "玫瑰", "Color": "#e11d48", "Desc": "温暖活力"},
				{"Key": "amber", "Label": "琥珀", "Color": "#d97706", "Desc": "友好亲和"},
				{"Key": "slate", "Label": "石板", "Color": "#0f172a", "Desc": "极简黑白"},
			}
		},
		"themeCSS": func() template.CSS {
			preset := models.GetSetting("theme_preset")
			if preset == "" {
				preset = "indigo"
			}
			themes := map[string]string{
				"indigo": `--color-primary:#4f46e5;--color-primary-hover:#4338ca;--color-primary-light:#eef2ff;--color-primary-text:#4338ca;--color-hero-bg:#0f172a;--color-hero-accent:#818cf8;`,
				"emerald": `--color-primary:#059669;--color-primary-hover:#047857;--color-primary-light:#ecfdf5;--color-primary-text:#047857;--color-hero-bg:#064e3b;--color-hero-accent:#6ee7b7;`,
				"rose": `--color-primary:#e11d48;--color-primary-hover:#be123c;--color-primary-light:#fff1f2;--color-primary-text:#be123c;--color-hero-bg:#1f1115;--color-hero-accent:#fda4af;`,
				"amber": `--color-primary:#d97706;--color-primary-hover:#b45309;--color-primary-light:#fffbeb;--color-primary-text:#b45309;--color-hero-bg:#1c1410;--color-hero-accent:#fcd34d;`,
				"slate": `--color-primary:#0f172a;--color-primary-hover:#1e293b;--color-primary-light:#f1f5f9;--color-primary-text:#0f172a;--color-hero-bg:#0f172a;--color-hero-accent:#94a3b8;`,
			}
			vars, ok := themes[preset]
			if !ok {
				vars = themes["indigo"]
			}
			return template.CSS(":root{" + vars + "}")
		},
		"adminSidebar": func() []NavItem {
			return []NavItem{
				{Key: "dashboard", Label: "仪表盘", URL: "/admin", Icon: "📊"},
				{Key: "services", Label: "服务管理", URL: "/admin/services", Icon: "🔧"},
				{Key: "posts", Label: "文章管理", URL: "/admin/posts", Icon: "📝"},
				{Key: "pages", Label: "页面管理", URL: "/admin/pages", Icon: "📄"},
				{Key: "projects", Label: "项目管理", URL: "/admin/projects", Icon: "📁"},
				{Key: "opensource", Label: "开源项目", URL: "/admin/opensource", Icon: "⭐"},
				{Key: "navigation", Label: "导航管理", URL: "/admin/navigation", Icon: "🗺️"},
				{Key: "skills", Label: "技能管理", URL: "/admin/skills", Icon: "🎯"},
				{Key: "tags", Label: "标签管理", URL: "/admin/tags", Icon: "🏷️"},
				{Key: "ads", Label: "广告管理", URL: "/admin/ads", Icon: "📢"},
				{Key: "friendlinks", Label: "友链管理", URL: "/admin/friendlinks", Icon: "🔗"},
				{Key: "social", Label: "社交账号", URL: "/admin/social", Icon: "📱"},
				{Key: "settings", Label: "站点设置", URL: "/admin/settings", Icon: "⚙️"},
				{Key: "messages", Label: "留言管理", URL: "/admin/messages", Icon: "💬"},
				{Key: "upload", Label: "文件上传", URL: "/admin/upload", Icon: "📎"},
			}
		},
		"add": func(a, b int) int { return a + b },
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"sub": func(a, b int) int { return a - b },
		"trim": func(s string) string { return strings.TrimSpace(s) },
		"trimPrefix": func(s, prefix string) string { return strings.TrimPrefix(s, prefix) },
		"div": func(a, b int64) int64 { return a / b },
		"split": func(s, sep string) []string { return strings.Split(s, sep) },
		"dict": func(values ...interface{}) map[string]interface{} {
			d := make(map[string]interface{})
			for i := 0; i < len(values)-1; i += 2 {
				if k, ok := values[i].(string); ok {
					d[k] = values[i+1]
				}
			}
			return d
		},
		"match": func(s, pattern string) bool {
			for _, p := range strings.Split(pattern, "|") {
				if strings.EqualFold(s, p) {
					return true
				}
			}
			return false
		},
	}

	baseTmpl = template.New("layout").Funcs(funcMap)
	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			_, err := baseTmpl.ParseFiles(path)
			return err
		}
		return nil
	})
	return err
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	tpl, err := baseTmpl.Clone()
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	pagePath := filepath.Join("templates", tmpl)
	if _, err := os.Stat(pagePath); err == nil {
		tpl, err = tpl.ParseFiles(pagePath)
		if err != nil {
			http.Error(w, "Template parse error", http.StatusInternalServerError)
			return
		}
	}

	if r != nil {
		if pd, ok := data.(PageData); ok {
			if cookie, err := r.Cookie("admin_token"); err == nil {
				sessionsMu.RLock()
				session := sessions[cookie.Value]
				sessionsMu.RUnlock()
				if session != nil {
					pd.CSRFToken = session.CSRFToken
				}
			}
			data = pd
		}
	}

	tname := "layout"
	if strings.HasSuffix(tmpl, "login.html") {
		tname = filepath.Base(tmpl)
	} else if strings.HasPrefix(tmpl, "admin/") {
		tname = "admin_layout"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, tname, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
