package handlers

import (
	"net/http"
	"strings"
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

		// Standalone: keep site nav/footer, render raw HTML content (extract from full doc if needed)
		if page.RenderMode == "standalone" {
			body, css := extractHTMLBody(page.Content)
			render(w, r, "page_standalone.html", PageData{
				Title:   page.Title + " · " + models.GetSetting("site_name"),
				Site:    cfg,
				Data:    map[string]interface{}{"page": page, "body": body, "css": css},
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

// extractHTMLBody extracts <body> content and <style>/<link>/<script> blocks from a full HTML doc.
// If input is just a fragment, returns it unchanged.
func extractHTMLBody(html string) (body, css string) {
	lower := strings.ToLower(html)
	// If not a full document, return as-is
	if !strings.Contains(lower, "<html") && !strings.Contains(lower, "<body") {
		return html, ""
	}

	// Extract <style>...</style> blocks (so they survive inside main body)
	var styles strings.Builder
	for {
		start := strings.Index(lower, "<style")
		if start < 0 {
			break
		}
		// find end of opening tag
		openEnd := strings.Index(lower[start:], ">")
		if openEnd < 0 {
			break
		}
		end := strings.Index(lower[start:], "</style>")
		if end < 0 {
			break
		}
		end += start + len("</style>")
		styles.WriteString(html[start:end])
		styles.WriteString("\n")
		// Remove from working copies
		html = html[:start] + html[end:]
		lower = strings.ToLower(html)
	}

	// Extract <body>...</body>
	bStart := strings.Index(lower, "<body")
	if bStart >= 0 {
		bOpenEnd := strings.Index(lower[bStart:], ">")
		if bOpenEnd >= 0 {
			afterOpen := bStart + bOpenEnd + 1
			bEnd := strings.Index(lower[afterOpen:], "</body>")
			if bEnd >= 0 {
				body = html[afterOpen : afterOpen+bEnd]
			} else {
				body = html[afterOpen:]
			}
		}
	}
	if body == "" {
		body = html
	}
	return body, styles.String()
}
