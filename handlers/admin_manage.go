package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"speedcraft/config"
	"speedcraft/models"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func AdminDashboardStats(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		total, pending := models.GetDashboardStats()
		services, _ := models.GetServices()
		posts, _ := models.GetAllPosts()
		projects, _ := models.GetAllProjects()
		opensource, _ := models.GetOpenSourceProjects()

		pubServices, privServices := 0, 0
		for _, s := range services {
			if s.IsPublished == 1 {
				pubServices++
			} else {
				privServices++
			}
		}
		pubPosts, privPosts := 0, 0
		for _, p := range posts {
			if p.IsPublished == 1 {
				pubPosts++
			} else {
				privPosts++
			}
		}
		pubProjects, privProjects := 0, 0
		for _, p := range projects {
			if p.IsPublished == 1 {
				pubProjects++
			} else {
				privProjects++
			}
		}
		pubOS, privOS := 0, 0
		for _, p := range opensource {
			if p.IsPublished == 1 {
				pubOS++
			} else {
				privOS++
			}
		}

		render(w, "admin/dashboard.html", PageData{
			Title:   "管理后台 · " + cfg.SiteName,
			Site:    cfg,
			Current: "dashboard",
			Data: map[string]interface{}{
				"totalMessages":   total,
				"pendingMessages": pending,
				"pubServices":     pubServices,
				"privServices":    privServices,
				"pubPosts":        pubPosts,
				"privPosts":       privPosts,
				"pubProjects":     pubProjects,
				"privProjects":    privProjects,
				"pubOS":           pubOS,
				"privOS":          privOS,
			},
		})
	})
}

// -------- Services CRUD --------
func AdminServices(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetServices()
		if list == nil {
			list = []models.Service{}
		}
		render(w, "admin/services.html", PageData{
			Title:   "服务管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "services",
			Data:    list,
		})
	})
}

func AdminServiceEdit(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		var svc *models.Service
		if idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			svc, _ = models.GetService(id)
		}
		if svc == nil {
			svc = &models.Service{IsPublished: 1}
			render(w, "admin/service_form.html", PageData{
				Title:   "新增服务 · " + cfg.SiteName,
				Site:    cfg,
				Current: "services",
				Data:    svc,
			})
		} else {
			render(w, "admin/service_form.html", PageData{
				Title:   "编辑服务 · " + cfg.SiteName,
				Site:    cfg,
				Current: "services",
				Data:    svc,
			})
		}
	})
}

func AdminServiceSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
		published := 0
		if r.FormValue("is_published") == "1" {
			published = 1
		}
		_, err := models.SaveService(&models.Service{
			ID:          id,
			Title:       r.FormValue("title"),
			Icon:        r.FormValue("icon"),
			Description: r.FormValue("description"),
			Features:    r.FormValue("features"),
			Pricing:     r.FormValue("pricing"),
			SortOrder:   sortOrder,
			IsPublished: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/services", http.StatusSeeOther)
	})
}

func AdminServiceDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteService(id)
		http.Redirect(w, r, "/admin/services", http.StatusSeeOther)
	})
}

// -------- Blog CRUD --------
func AdminPosts(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetAllPosts()
		if list == nil {
			list = []models.BlogPost{}
		}
		render(w, "admin/posts.html", PageData{
			Title:   "文章管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "posts",
			Data:    list,
		})
	})
}

func AdminPostEdit(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		var post *models.BlogPost
		if idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			post, _ = models.GetPost(id)
		}
		if post == nil {
			post = &models.BlogPost{ContentType: "markdown", IsPublished: 0}
		}
		render(w, "admin/post_form.html", PageData{
			Title:   "编辑文章 · " + cfg.SiteName,
			Site:    cfg,
			Current: "posts",
			Data:    post,
		})
	})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "?", "")
	s = strings.ReplaceAll(s, "&", "")
	s = strings.ReplaceAll(s, "=", "")
	s = strings.ReplaceAll(s, ".", "")
	return s
}

func AdminPostSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		published := 0
		if r.FormValue("is_published") == "1" {
			published = 1
		}

		slug := r.FormValue("slug")
		if slug == "" {
			slug = slugify(r.FormValue("title"))
		}

		_, err := models.SavePost(&models.BlogPost{
			ID:          id,
			Title:       r.FormValue("title"),
			Slug:        slug,
			Summary:     r.FormValue("summary"),
			Content:     r.FormValue("content"),
			ContentType: r.FormValue("content_type"),
			Tags:        r.FormValue("tags"),
			IsPublished: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
	})
}

func AdminPostDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeletePost(id)
		http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
	})
}

// -------- Projects CRUD --------
func AdminProjects(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetAllProjects()
		if list == nil {
			list = []models.Project{}
		}
		render(w, "admin/projects.html", PageData{
			Title:   "项目管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "projects",
			Data:    list,
		})
	})
}

func AdminProjectEdit(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		var p *models.Project
		if idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			p, _ = models.GetProject(id)
		}
		if p == nil {
			p = &models.Project{IsPublished: 0}
		}
		render(w, "admin/project_form.html", PageData{
			Title:   "编辑项目 · " + cfg.SiteName,
			Site:    cfg,
			Current: "projects",
			Data:    p,
		})
	})
}

func AdminProjectSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		published := 0
		if r.FormValue("is_published") == "1" {
			published = 1
		}
		slug := r.FormValue("slug")
		if slug == "" {
			slug = slugify(r.FormValue("title"))
		}

		_, err := models.SaveProject(&models.Project{
			ID:          id,
			Title:       r.FormValue("title"),
			Slug:        slug,
			Summary:     r.FormValue("summary"),
			Content:     r.FormValue("content"),
			Category:    r.FormValue("category"),
			TechStack:   r.FormValue("tech_stack"),
			ImageURL:    r.FormValue("image_url"),
			ClientName:  r.FormValue("client_name"),
			ClientURL:   r.FormValue("client_url"),
			IsPublished: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
	})
}

func AdminProjectDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteProject(id)
		http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
	})
}

// -------- Open Source CRUD --------
func AdminOpenSource(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetOpenSourceProjects()
		if list == nil {
			list = []models.OpenSourceProject{}
		}
		render(w, "admin/opensource.html", PageData{
			Title:   "开源项目管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "opensource",
			Data:    list,
		})
	})
}

func AdminOpenSourceEdit(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		var p *models.OpenSourceProject
		if idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			p, _ = models.GetOpenSourceProject(id)
		}
		if p == nil {
			p = &models.OpenSourceProject{IsPublished: 1}
		}
		render(w, "admin/opensource_form.html", PageData{
			Title:   "编辑开源项目 · " + cfg.SiteName,
			Site:    cfg,
			Current: "opensource",
			Data:    p,
		})
	})
}

func AdminOpenSourceSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		stars, _ := strconv.Atoi(r.FormValue("stars"))
		sort, _ := strconv.Atoi(r.FormValue("sort_order"))
		published := 0
		if r.FormValue("is_published") == "1" {
			published = 1
		}
		featured := 0
		if r.FormValue("is_featured") == "1" {
			featured = 1
		}

		_, err := models.SaveOpenSourceProject(&models.OpenSourceProject{
			ID:          id,
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			URL:         r.FormValue("url"),
			GithubURL:   r.FormValue("github_url"),
			Stars:       stars,
			Language:    r.FormValue("language"),
			LicenseType: r.FormValue("license_type"),
			IsFeatured:  featured,
			SortOrder:   sort,
			IsPublished: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/opensource", http.StatusSeeOther)
	})
}

func AdminOpenSourceDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteOpenSourceProject(id)
		http.Redirect(w, r, "/admin/opensource", http.StatusSeeOther)
	})
}

// -------- Navigation CRUD --------
func AdminNavigation(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetNavigation()
		if list == nil {
			list = []models.NavItem{}
		}
		render(w, "admin/navigation.html", PageData{
			Title:   "导航管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "navigation",
			Data:    list,
		})
	})
}

func AdminNavigationSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		sort, _ := strconv.Atoi(r.FormValue("sort_order"))
		published := 0
		if r.FormValue("is_published") == "1" {
			published = 1
		}

		_, err := models.SaveNavItem(&models.NavItem{
			ID:        id,
			Label:     r.FormValue("label"),
			URL:       r.FormValue("url"),
			Icon:      r.FormValue("icon"),
			SortOrder: sort,
			Published: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/navigation", http.StatusSeeOther)
	})
}

func AdminNavigationDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteNavItem(id)
		http.Redirect(w, r, "/admin/navigation", http.StatusSeeOther)
	})
}

// -------- Settings --------
func AdminSettings(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		settings, _ := models.GetAllSettings()
		if settings == nil {
			settings = make(map[string]string)
		}
		render(w, "admin/settings.html", PageData{
			Title:   "站点设置 · " + cfg.SiteName,
			Site:    cfg,
			Current: "settings",
			Data:    settings,
		})
	})
}

func AdminSettingsSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		settings := make(map[string]string)
		for key := range r.Form {
			if key != "id" {
				settings[key] = r.FormValue(key)
			}
		}
		if err := models.SaveSettings(settings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
	})
}

// -------- Skills CRUD --------
func AdminSkills(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetSkills()
		if list == nil {
			list = []models.Skill{}
		}
		render(w, "admin/skills.html", PageData{
			Title:   "技能管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "skills",
			Data:    list,
		})
	})
}

func AdminSkillSave(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		sort, _ := strconv.Atoi(r.FormValue("sort_order"))
		published := 0
		if r.FormValue("is_published") == "1" {
			published = 1
		}

		_, err := models.SaveSkill(&models.Skill{
			ID:        id,
			Icon:      r.FormValue("icon"),
			Name:      r.FormValue("name"),
			Level:     r.FormValue("level"),
			Category:  r.FormValue("category"),
			SortOrder: sort,
			Published: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/skills", http.StatusSeeOther)
	})
}

func AdminSkillDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteSkill(id)
		http.Redirect(w, r, "/admin/skills", http.StatusSeeOther)
	})
}

// -------- File Upload --------
func AdminUpload(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			uploadDir := "static/uploads"
			os.MkdirAll(uploadDir, 0755)

			files, _ := os.ReadDir(uploadDir)
			type fileInfo struct {
				Name string
				URL  string
				Size int64
				Ext  string
			}
			var fileList []fileInfo
			for _, f := range files {
				if !f.IsDir() {
					info, _ := f.Info()
					ext := filepath.Ext(f.Name())
					fileList = append(fileList, fileInfo{
						Name: f.Name(),
						URL:  "/static/uploads/" + f.Name(),
						Size: info.Size(),
						Ext:  ext,
					})
				}
			}

			render(w, "admin/upload.html", PageData{
				Title:   "文件上传 · " + cfg.SiteName,
				Site:    cfg,
				Current: "upload",
				Data:    fileList,
			})
			return
		}

		r.ParseMultipartForm(10 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "文件上传失败"})
			return
		}
		defer file.Close()

		ext := filepath.Ext(handler.Filename)
		uname := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(handler.Filename, ext), ext)
		uploadDir := "static/uploads"
		os.MkdirAll(uploadDir, 0755)

		dst, err := os.Create(filepath.Join(uploadDir, uname))
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "文件保存失败"})
			return
		}
		defer dst.Close()

		io.Copy(dst, file)

		respondJSON(w, http.StatusOK, map[string]string{
			"url": "/static/uploads/" + uname,
			"name": uname,
		})
	})
}

// -------- Markdown Preview --------
func AdminMarkdownPreview(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		md := r.FormValue("content")
		mdRenderer := goldmark.New(
			goldmark.WithExtensions(extension.GFM, extension.TaskList, extension.Footnote, extension.Typographer),
		)
		var buf strings.Builder
		if err := mdRenderer.Convert([]byte(md), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(buf.String()))
	})
}
