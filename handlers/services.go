package handlers

import (
	"net/http"
	"speedcraft/config"
	"speedcraft/models"
)

func Services(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetPublishedServices()
		if list == nil {
			list = []models.Service{}
		}
		render(w, r, "services.html", PageData{
			Title:   "服务 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    list,
			Current: "services",
		})
	}
}

func OpenSource(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, _ := models.GetPublishedOpenSource()
		if list == nil {
			list = []models.OpenSourceProject{}
		}
		render(w, r, "opensource.html", PageData{
			Title:   "开源 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    list,
			Current: "opensource",
		})
	}
}

func PublicPage(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			http.NotFound(w, r)
			return
		}
		page, err := models.GetPageBySlug(slug)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Standalone: keep site nav/footer, embed raw HTML in iframe for full style isolation
		if page.RenderMode == "standalone" {
			render(w, r, "page_standalone.html", PageData{
				Title:   page.Title + " · " + models.GetSetting("site_name"),
				Site:    cfg,
				Data:    map[string]interface{}{"page": page},
				Current: slug,
			})
			return
		}

		// Embed mode: wrap with site layout + hero, render markdown if applicable
		content := page.Content
		if page.ContentType == "markdown" {
			content = renderMarkdown(content)
		}
		render(w, r, "page.html", PageData{
			Title:   page.Title + " · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    map[string]interface{}{"page": page, "content": content},
			Current: slug,
		})
	}
}

// PublicPageRaw serves the raw HTML content for iframe srcdoc.
func PublicPageRaw(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			http.NotFound(w, r)
			return
		}
		page, err := models.GetPageBySlug(slug)
		if err != nil || page.RenderMode != "standalone" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page.Content))
	}
}
