package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"speedcraft/config"
	"speedcraft/models"
)

// Sitemap generates sitemap.xml dynamically
func Sitemap(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := scheme(r) + "://" + r.Host
		var sb strings.Builder
		sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
		sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

		// Static pages
		static := []struct {
			path     string
			priority string
		}{
			{"/", "1.0"},
			{"/services", "0.9"},
			{"/portfolio", "0.8"},
			{"/opensource", "0.8"},
			{"/blog", "0.9"},
			{"/about", "0.7"},
		}
		for _, p := range static {
			sb.WriteString(fmt.Sprintf(`  <url><loc>%s%s</loc><priority>%s</priority></url>`+"\n", base, p.path, p.priority))
		}

		// Blog posts
		if posts, err := models.GetPublishedPosts(); err == nil {
			for _, p := range posts {
				sb.WriteString(fmt.Sprintf(`  <url><loc>%s/blog/%s</loc><lastmod>%s</lastmod><priority>0.7</priority></url>`+"\n",
					base, p.Slug, p.CreatedAt.Format("2006-01-02")))
			}
		}

		// Custom pages
		if pages, err := models.GetPublishedPages(); err == nil {
			for _, p := range pages {
				sb.WriteString(fmt.Sprintf(`  <url><loc>%s/page/%s</loc><lastmod>%s</lastmod><priority>0.6</priority></url>`+"\n",
					base, p.Slug, p.UpdatedAt.Format("2006-01-02")))
			}
		}

		sb.WriteString(`</urlset>`)
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write([]byte(sb.String()))
	}
}

// RobotsTxt generates robots.txt
func RobotsTxt(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := scheme(r) + "://" + r.Host
		txt := fmt.Sprintf("User-agent: *\nDisallow: /admin/\nDisallow: /api/\nAllow: /\n\nSitemap: %s/sitemap.xml\n", base)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(txt))
	}
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}