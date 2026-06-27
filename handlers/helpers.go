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
