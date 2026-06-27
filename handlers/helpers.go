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
	Title   string
	Site    *config.Config
	Data    interface{}
	Current string
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
		"adminSidebar": func() []NavItem {
			return []NavItem{
				{Key: "dashboard", Label: "仪表盘", URL: "/admin", Icon: "📊"},
				{Key: "services", Label: "服务管理", URL: "/admin/services", Icon: "🔧"},
				{Key: "posts", Label: "文章管理", URL: "/admin/posts", Icon: "📝"},
				{Key: "projects", Label: "项目管理", URL: "/admin/projects", Icon: "📁"},
				{Key: "opensource", Label: "开源项目", URL: "/admin/opensource", Icon: "⭐"},
				{Key: "navigation", Label: "导航管理", URL: "/admin/navigation", Icon: "🗺️"},
				{Key: "skills", Label: "技能管理", URL: "/admin/skills", Icon: "🎯"},
				{Key: "settings", Label: "站点设置", URL: "/admin/settings", Icon: "⚙️"},
				{Key: "messages", Label: "留言管理", URL: "/admin/messages", Icon: "💬"},
				{Key: "upload", Label: "文件上传", URL: "/admin/upload", Icon: "📎"},
			}
		},
		"add": func(a, b int) int { return a + b },
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"sub": func(a, b int) int { return a - b },
		"trim": func(s string) string { return strings.TrimSpace(s) },
		"split": func(s, sep string) []string { return strings.Split(s, sep) },
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

func render(w http.ResponseWriter, tmpl string, data interface{}) {
	tpl, err := baseTmpl.Clone()
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	pageName := filepath.Base(tmpl)
	if tpl.Lookup(pageName) == nil {
		pagePath := filepath.Join("templates", tmpl)
		if _, err := os.Stat(pagePath); err == nil {
			tpl, err = tpl.ParseFiles(pagePath)
			if err != nil {
				http.Error(w, "Template parse error", http.StatusInternalServerError)
				return
			}
		}
	}

	tname := "layout"
	if strings.HasSuffix(tmpl, "login.html") {
		tname = pageName
	} else if tpl.Lookup("admin_layout") != nil {
		tname = "admin_layout"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, tname, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
