package handlers

import (
	"database/sql"
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

		// Page analytics
		topPages, _ := models.GetTopPageStats(8)
		slowestPages, _ := models.GetSlowestPages(5)
		totalViews := models.GetTotalPageViews()
		todayViews := models.GetTodayPageViews()
		dailyTrend, _ := models.GetRecentDailyViews(7)
		pendingFriendLinks := models.GetPendingFriendLinksCount()

		render(w, r, "admin/dashboard.html", PageData{
			Title:   "管理后台 · " + cfg.SiteName,
			Site:    cfg,
			Current: "dashboard",
			Data: map[string]interface{}{
				"totalMessages":      total,
				"pendingMessages":    pending,
				"pubServices":        pubServices,
				"privServices":       privServices,
				"pubPosts":           pubPosts,
				"privPosts":          privPosts,
				"pubProjects":        pubProjects,
				"privProjects":       privProjects,
				"pubOS":              pubOS,
				"privOS":             privOS,
				"totalViews":         totalViews,
				"todayViews":         todayViews,
				"topPages":           topPages,
				"slowestPages":       slowestPages,
				"dailyTrend":         dailyTrend,
				"pendingFriendLinks": pendingFriendLinks,
			},
		})
	})
}

// -------- Services CRUD --------
func AdminServices(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize := 50
		list, total, err := models.GetServicesPage(page, pageSize)
		if err != nil {
			list = []models.Service{}
		}
		totalPages := (total + pageSize - 1) / pageSize
		render(w, r, "admin/services.html", PageData{
			Title:   "服务管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "services",
			Data:    map[string]interface{}{"items": list, "Page": page, "TotalPages": totalPages, "Total": total},
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
			render(w, r, "admin/service_form.html", PageData{
				Title:   "新增服务 · " + cfg.SiteName,
				Site:    cfg,
				Current: "services",
				Data:    svc,
			})
		} else {
			render(w, r, "admin/service_form.html", PageData{
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
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize := 50
		list, total, err := models.GetAllPostsPage(page, pageSize)
		if err != nil {
			list = []models.BlogPost{}
		}
		totalPages := (total + pageSize - 1) / pageSize
		render(w, r, "admin/posts.html", PageData{
			Title:   "文章管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "posts",
			Data:    map[string]interface{}{"items": list, "Page": page, "TotalPages": totalPages, "Total": total},
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
		render(w, r, "admin/post_form.html", PageData{
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
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize := 50
		list, total, err := models.GetAllProjectsPage(page, pageSize)
		if err != nil {
			list = []models.Project{}
		}
		totalPages := (total + pageSize - 1) / pageSize
		render(w, r, "admin/projects.html", PageData{
			Title:   "项目管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "projects",
			Data:    map[string]interface{}{"items": list, "Page": page, "TotalPages": totalPages, "Total": total},
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
		render(w, r, "admin/project_form.html", PageData{
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
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize := 50
		list, total, err := models.GetOpenSourceProjectsPage(page, pageSize)
		if err != nil {
			list = []models.OpenSourceProject{}
		}
		totalPages := (total + pageSize - 1) / pageSize
		render(w, r, "admin/opensource.html", PageData{
			Title:   "开源项目管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "opensource",
			Data:    map[string]interface{}{"items": list, "Page": page, "TotalPages": totalPages, "Total": total},
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
		render(w, r, "admin/opensource_form.html", PageData{
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
		pages, _ := models.GetPublishedPages()
		if pages == nil {
			pages = []models.Page{}
		}

		editItem := &models.NavItem{}
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				for _, n := range list {
					if n.ID == id {
						editItem = &n
						break
					}
				}
			}
		}

		render(w, r, "admin/navigation.html", PageData{
			Title:   "导航管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "navigation",
			Data:    map[string]interface{}{"items": list, "edit": editItem, "pages": pages},
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

func AdminNavigationReorder(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
			return
		}
		r.ParseForm()
		idsStr := r.Form["ids"]
		var ids []int64
		for _, s := range idsStr {
			id, err := strconv.ParseInt(s, 10, 64)
			if err == nil {
				ids = append(ids, id)
			}
		}
		if err := models.ReorderNavigation(ids); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

// -------- Settings --------
func AdminSettings(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		settings, _ := models.GetAllSettings()
		if settings == nil {
			settings = make(map[string]string)
		}
		render(w, r, "admin/settings.html", PageData{
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
			if key == "id" || key == "csrf_token" {
				continue
			}
			settings[key] = r.FormValue(key)
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
		var editItem models.Skill
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				for _, s := range list {
					if s.ID == id {
						editItem = s
						break
					}
				}
			}
		}
		render(w, r, "admin/skills.html", PageData{
			Title:   "技能管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "skills",
			Data:    map[string]interface{}{"items": list, "edit": editItem},
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
				Name string `json:"name"`
				URL  string `json:"url"`
				Size int64  `json:"size"`
				Ext  string `json:"ext"`
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

			if r.URL.Query().Get("list") == "json" {
				respondJSON(w, http.StatusOK, fileList)
				return
			}

			render(w, r, "admin/upload.html", PageData{
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

// -------- File Delete --------
func AdminUploadDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "无效文件名"})
			return
		}
		path := filepath.Join("static/uploads", name)
		if err := os.Remove(path); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
			return
		}
		http.Redirect(w, r, "/admin/upload", http.StatusSeeOther)
	})
}

// -------- Markdown Preview --------
func AdminMarkdownPreview(cfg *config.Config) http.HandlerFunc {
	return SessionOnlyMiddleware(func(w http.ResponseWriter, r *http.Request) {
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

func AdminTogglePublish(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
			return
		}
		table := r.FormValue("table")
		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
			return
		}
		validTables := map[string]bool{"services": true, "blog_posts": true, "projects": true, "open_source_projects": true, "skills": true, "navigation_items": true}
		if !validTables[table] {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid table"})
			return
		}
		newVal, err := models.TogglePublish(table, id)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"published": newVal})
	})
}

func AdminTags(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			name := strings.TrimSpace(r.FormValue("name"))
			if name == "" {
				http.Redirect(w, r, "/admin/tags?error=名称不能为空", http.StatusSeeOther)
				return
			}
			slug := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", "-"), "/", "-"))
			_, err := models.CreateTag(name, slug)
			if err != nil {
				http.Redirect(w, r, "/admin/tags?error=标签已存在", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/admin/tags", http.StatusSeeOther)
			return
		}
		list, _ := models.GetAllTags()
		if list == nil {
			list = []models.Tag{}
		}
		render(w, r, "admin/tags.html", PageData{
			Title:   "标签管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "tags",
			Data:    list,
		})
	})
}

func AdminTagDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/admin/tags?error=无效ID", http.StatusSeeOther)
			return
		}
		models.DeleteTag(id)
		http.Redirect(w, r, "/admin/tags", http.StatusSeeOther)
	})
}

func AdminTagsJSON(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, err := models.GetAllTags()
		if err != nil {
			list = []models.Tag{}
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"tags": list})
	})
}

func AdminBatchAction(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
			return
		}
		table := r.FormValue("table")
		action := r.FormValue("action")
		idsStr := r.FormValue("ids")
		if idsStr == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "No IDs provided"})
			return
		}
		parts := strings.Split(idsStr, ",")
		var ids []int64
		for _, p := range parts {
			id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
			if err == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid IDs"})
			return
		}
		if err := models.BatchAction(table, action, ids); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
	})
}

// -------- Page CRUD --------
func AdminPages(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetAllPages()
		if list == nil {
			list = []models.Page{}
		}
		var editItem models.Page
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				for _, p := range list {
					if p.ID == id {
						editItem = p
						break
					}
				}
			}
		}
		render(w, r, "admin/pages.html", PageData{
			Title:   "页面管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "pages",
			Data:    map[string]interface{}{"items": list, "edit": editItem},
		})
	})
}

func AdminPageSave(cfg *config.Config) http.HandlerFunc {
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
		slug := strings.TrimSpace(r.FormValue("slug"))
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(r.FormValue("title")), " ", "-"))
		}
		renderMode := r.FormValue("render_mode")
		if renderMode != "standalone" {
			renderMode = "embed"
		}
		_, err := models.SavePage(&models.Page{
			ID:          id,
			Title:       r.FormValue("title"),
			Slug:        slug,
			Content:     r.FormValue("content"),
			ContentType: r.FormValue("content_type"),
			RenderMode:  renderMode,
			IsPublished: published,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/pages", http.StatusSeeOther)
	})
}

func AdminPageDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeletePage(id)
		http.Redirect(w, r, "/admin/pages", http.StatusSeeOther)
	})
}

// -------- Ad CRUD --------
func AdminAds(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetAllAds()
		if list == nil {
			list = []models.Ad{}
		}
		var editItem models.Ad
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				if a, err := models.GetAdByID(id); err == nil && a != nil {
					editItem = *a
				}
			}
		}
		render(w, r, "admin/ads.html", PageData{
			Title:   "广告管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "ads",
			Data:    map[string]interface{}{"items": list, "edit": editItem},
		})
	})
}

func AdminAdSave(cfg *config.Config) http.HandlerFunc {
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
		ad := &models.Ad{
			ID:          id,
			Name:        r.FormValue("name"),
			Slot:        r.FormValue("slot"),
			AdType:      r.FormValue("ad_type"),
			ImageURL:    r.FormValue("image_url"),
			LinkURL:     r.FormValue("link_url"),
			HTMLContent: r.FormValue("html_content"),
			AltText:     r.FormValue("alt_text"),
			SortOrder:   sortOrder,
			IsPublished: published,
		}
		if s := strings.TrimSpace(r.FormValue("start_at")); s != "" {
			if t, err := time.Parse("2006-01-02T15:04", s); err == nil {
				ad.StartAt = sql.NullTime{Time: t, Valid: true}
			}
		}
		if s := strings.TrimSpace(r.FormValue("end_at")); s != "" {
			if t, err := time.Parse("2006-01-02T15:04", s); err == nil {
				ad.EndAt = sql.NullTime{Time: t, Valid: true}
			}
		}
		if _, err := models.SaveAd(ad); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/ads", http.StatusSeeOther)
	})
}

func AdminAdDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteAd(id)
		http.Redirect(w, r, "/admin/ads", http.StatusSeeOther)
	})
}

// AdClick records a click and redirects to the ad's link
func AdClick(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ad, err := models.GetAdByID(id)
		if err != nil || ad == nil || ad.LinkURL == "" {
			http.NotFound(w, r)
			return
		}
		models.IncrementAdClick(id)
		http.Redirect(w, r, ad.LinkURL, http.StatusFound)
	}
}

// -------- FriendLink Admin CRUD --------
func AdminFriendLinks(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetAllFriendLinks()
		if list == nil {
			list = []models.FriendLink{}
		}
		var editItem models.FriendLink
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				if f, err := models.GetFriendLinkByID(id); err == nil && f != nil {
					editItem = *f
				}
			}
		}
		render(w, r, "admin/friendlinks.html", PageData{
			Title:   "友链管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "friendlinks",
			Data:    map[string]interface{}{"items": list, "edit": editItem},
		})
	})
}

func AdminFriendLinkSave(cfg *config.Config) http.HandlerFunc {
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
		status := r.FormValue("status")
		if status == "" {
			status = "approved"
		}
		f := &models.FriendLink{
			ID:             id,
			Name:           r.FormValue("name"),
			URL:            r.FormValue("url"),
			LogoURL:        r.FormValue("logo_url"),
			Description:    r.FormValue("description"),
			Category:       r.FormValue("category"),
			Status:         status,
			SubmitterEmail: r.FormValue("submitter_email"),
			SubmitterNote:  r.FormValue("submitter_note"),
			SortOrder:      sortOrder,
			IsPublished:    published,
		}
		if _, err := models.SaveFriendLink(f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/friendlinks", http.StatusSeeOther)
	})
}

// AdminFriendLinkAction handles approve/reject/delete via GET (for simplicity inside admin)
func AdminFriendLinkAction(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		action := r.URL.Query().Get("action")
		f, err := models.GetFriendLinkByID(id)
		if err != nil || f == nil {
			http.NotFound(w, r)
			return
		}
		switch action {
		case "approve":
			f.Status = "approved"
			f.IsPublished = 1
		case "reject":
			f.Status = "rejected"
			f.IsPublished = 0
		case "delete":
			models.DeleteFriendLink(id)
			http.Redirect(w, r, "/admin/friendlinks", http.StatusSeeOther)
			return
		}
		models.SaveFriendLink(f)
		http.Redirect(w, r, "/admin/friendlinks", http.StatusSeeOther)
	})
}

// -------- Public FriendLinks handlers --------
func PublicFriendLinks(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetPublishedFriendLinks()
		if list == nil {
			list = []models.FriendLink{}
		}
		render(w, r, "links.html", PageData{
			Title:        "友链 · " + models.GetSetting("site_name"),
			Site:         cfg,
			Data:         list,
			Current:      "links",
			CanonicalURL: scheme(r) + "://" + r.Host + "/links",
		})
	}
}

// ApplyFriendLink handles the public application form submission.
func ApplyFriendLink(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		url := strings.TrimSpace(r.FormValue("url"))
		if name == "" || url == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "error": "名称和链接必填"})
			return
		}
		// Basic URL validation
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "error": "链接必须以 http:// 或 https:// 开头"})
			return
		}
		f := &models.FriendLink{
			Name:           name,
			URL:            url,
			LogoURL:        strings.TrimSpace(r.FormValue("logo_url")),
			Description:    strings.TrimSpace(r.FormValue("description")),
			Category:       strings.TrimSpace(r.FormValue("category")),
			Status:         "pending",
			SubmitterEmail: strings.TrimSpace(r.FormValue("submitter_email")),
			SubmitterNote:  strings.TrimSpace(r.FormValue("submitter_note")),
			IsPublished:    0,
		}
		if _, err := models.SaveFriendLink(f); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "提交失败"})
			return
		}

		// Notify admin asynchronously
		go SendGenericNotification("🔗 新友链申请", fmt.Sprintf(
			"📌 站点: %s\n🌐 链接: %s\n📂 分类: %s\n📝 简介: %s\n📧 提交者: %s\n💬 备注: %s\n\n请到后台 /admin/friendlinks 审核",
			f.Name, f.URL, f.Category, f.Description, f.SubmitterEmail, f.SubmitterNote,
		))

		respondJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "申请已提交，等待审核"})
	}
}

// -------- SocialAccount Admin --------
func AdminSocialAccounts(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetAllSocialAccounts()
		if list == nil {
			list = []models.SocialAccount{}
		}
		var editItem models.SocialAccount
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				if s, err := models.GetSocialAccountByID(id); err == nil && s != nil {
					editItem = *s
				}
			}
		}
		render(w, r, "admin/social.html", PageData{
			Title:   "社交账号 · " + cfg.SiteName,
			Site:    cfg,
			Current: "social",
			Data:    map[string]interface{}{"items": list, "edit": editItem},
		})
	})
}

func AdminSocialSave(cfg *config.Config) http.HandlerFunc {
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
		s := &models.SocialAccount{
			ID:          id,
			Platform:    r.FormValue("platform"),
			Name:        r.FormValue("name"),
			Identifier:  r.FormValue("identifier"),
			URL:         r.FormValue("url"),
			QRURL:       r.FormValue("qr_url"),
			Description: r.FormValue("description"),
			SortOrder:   sortOrder,
			IsPublished: published,
		}
		if _, err := models.SaveSocialAccount(s); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/social", http.StatusSeeOther)
	})
}

func AdminSocialDelete(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		models.DeleteSocialAccount(id)
		http.Redirect(w, r, "/admin/social", http.StatusSeeOther)
	})
}
