package main

import (
	"log"
	"net/http"
	"speedcraft/config"
	"speedcraft/database"
	"speedcraft/handlers"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("[FATAL] 数据库初始化失败: %v", err)
	}
	defer database.Close()

	if err := handlers.InitTemplates(); err != nil {
		log.Fatalf("[FATAL] 模板加载失败: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.Home(cfg))
	mux.HandleFunc("/services", handlers.Services(cfg))
	mux.HandleFunc("/portfolio", handlers.Portfolio(cfg))
	mux.HandleFunc("/opensource", handlers.OpenSource(cfg))
	mux.HandleFunc("/blog", handlers.Blog(cfg))
	mux.HandleFunc("/blog/", handlers.BlogPost(cfg))
	mux.HandleFunc("/about", handlers.About(cfg))
	mux.HandleFunc("/contact", handlers.About(cfg))
	mux.HandleFunc("/page/{slug}", handlers.PublicPage(cfg))
	mux.HandleFunc("/page/{slug}/raw", handlers.PublicPageRaw(cfg))

	mux.HandleFunc("/api/message", handlers.SubmitMessage(cfg))

	mux.HandleFunc("/admin/login", handlers.AdminLogin(cfg))
	mux.HandleFunc("/admin/logout", handlers.AdminLogout(cfg))

	mux.HandleFunc("/admin", handlers.AdminDashboardStats(cfg))
	mux.HandleFunc("/admin/services", handlers.AdminServices(cfg))
	mux.HandleFunc("/admin/services/edit", handlers.AdminServiceEdit(cfg))
	mux.HandleFunc("/admin/services/save", handlers.AdminServiceSave(cfg))
	mux.HandleFunc("/admin/services/delete", handlers.AdminServiceDelete(cfg))

	mux.HandleFunc("/admin/posts", handlers.AdminPosts(cfg))
	mux.HandleFunc("/admin/posts/edit", handlers.AdminPostEdit(cfg))
	mux.HandleFunc("/admin/posts/save", handlers.AdminPostSave(cfg))
	mux.HandleFunc("/admin/posts/delete", handlers.AdminPostDelete(cfg))

	mux.HandleFunc("/admin/projects", handlers.AdminProjects(cfg))
	mux.HandleFunc("/admin/projects/edit", handlers.AdminProjectEdit(cfg))
	mux.HandleFunc("/admin/projects/save", handlers.AdminProjectSave(cfg))
	mux.HandleFunc("/admin/projects/delete", handlers.AdminProjectDelete(cfg))

	mux.HandleFunc("/admin/opensource", handlers.AdminOpenSource(cfg))
	mux.HandleFunc("/admin/opensource/edit", handlers.AdminOpenSourceEdit(cfg))
	mux.HandleFunc("/admin/opensource/save", handlers.AdminOpenSourceSave(cfg))
	mux.HandleFunc("/admin/opensource/delete", handlers.AdminOpenSourceDelete(cfg))

	mux.HandleFunc("/admin/navigation", handlers.AdminNavigation(cfg))
	mux.HandleFunc("/admin/navigation/save", handlers.AdminNavigationSave(cfg))
	mux.HandleFunc("/admin/navigation/delete", handlers.AdminNavigationDelete(cfg))
	mux.HandleFunc("/admin/navigation/reorder", handlers.AdminNavigationReorder(cfg))

	mux.HandleFunc("/admin/settings", handlers.AdminSettings(cfg))
	mux.HandleFunc("/admin/settings/save", handlers.AdminSettingsSave(cfg))

	mux.HandleFunc("/admin/skills", handlers.AdminSkills(cfg))
	mux.HandleFunc("/admin/skills/save", handlers.AdminSkillSave(cfg))
	mux.HandleFunc("/admin/skills/delete", handlers.AdminSkillDelete(cfg))

	mux.HandleFunc("/admin/messages", handlers.AdminMessages(cfg))
	mux.HandleFunc("/admin/messages/update", handlers.AdminUpdateMessage(cfg))
	mux.HandleFunc("/admin/messages/export", handlers.AdminMessageExport(cfg))

	mux.HandleFunc("/admin/upload", handlers.AdminUpload(cfg))
	mux.HandleFunc("/admin/upload/delete", handlers.AdminUploadDelete(cfg))
	mux.HandleFunc("/admin/preview", handlers.AdminMarkdownPreview(cfg))
	mux.HandleFunc("/admin/toggle-publish", handlers.AdminTogglePublish(cfg))
	mux.HandleFunc("/admin/test-notification", handlers.AdminTestNotification(cfg))
	mux.HandleFunc("/admin/batch", handlers.AdminBatchAction(cfg))
	mux.HandleFunc("/admin/tags", handlers.AdminTags(cfg))
	mux.HandleFunc("/admin/tags/delete", handlers.AdminTagDelete(cfg))
	mux.HandleFunc("/admin/tags/json", handlers.AdminTagsJSON(cfg))

	mux.HandleFunc("/admin/pages", handlers.AdminPages(cfg))
	mux.HandleFunc("/admin/pages/save", handlers.AdminPageSave(cfg))
	mux.HandleFunc("/admin/pages/delete", handlers.AdminPageDelete(cfg))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	handler := withMiddleware(mux)

	addr := ":" + cfg.Port
	log.Printf("[INFO] 🚀 %s 服务启动于 http://localhost%s", cfg.SiteName, addr)
	log.Printf("[INFO] 📝 管理后台: http://localhost%s/admin", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("[FATAL] 服务启动失败: %v", err)
	}
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Allow same-origin iframe (used by standalone page rendering); still blocks external embedding
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}
