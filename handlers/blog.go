package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"speedcraft/config"
	"speedcraft/models"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.TaskList, extension.Footnote, extension.Typographer),
)

func renderMarkdown(content string) string {
	var buf strings.Builder
	mdRenderer.Convert([]byte(content), &buf)
	return buf.String()
}

func Blog(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, _ := models.GetPublishedPosts()
		if posts == nil {
			posts = []models.BlogPost{}
		}
		siteName := models.GetSetting("site_name")
		render(w, r, "blog.html", PageData{
			Title:           "博客 · " + siteName,
			Site:            cfg,
			Data:            posts,
			Current:         "blog",
			MetaDescription: "技术博客 - " + siteName + " 分享 DevOps、云原生、MVP 开发等实战经验",
			CanonicalURL:    scheme(r) + "://" + r.Host + "/blog",
		})
	}
}

func BlogPost(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/blog/")
		post, err := models.GetPostBySlug(slug)
		if err != nil || post == nil {
			http.NotFound(w, r)
			return
		}

		content := post.Content
		if post.ContentType == "markdown" {
			content = renderMarkdown(content)
		}

		models.IncrementPostViews(post.ID)
		post.Views++

		siteName := models.GetSetting("site_name")
		canonical := scheme(r) + "://" + r.Host + "/blog/" + post.Slug

		// JSON-LD Article schema
		ld := map[string]interface{}{
			"@context":      "https://schema.org",
			"@type":         "BlogPosting",
			"headline":      post.Title,
			"description":   post.Summary,
			"datePublished": post.CreatedAt.Format("2006-01-02"),
			"author":        map[string]interface{}{"@type": "Person", "name": siteName},
			"publisher":     map[string]interface{}{"@type": "Organization", "name": siteName},
			"mainEntityOfPage": map[string]interface{}{
				"@type": "WebPage",
				"@id":   canonical,
			},
			"keywords": post.Tags,
		}
		ldJSON, _ := json.Marshal(ld)

		desc := post.Summary
		if desc == "" {
			// fallback: first 150 chars of content stripped
			plain := stripHTML(content)
			if len(plain) > 150 {
				desc = plain[:150] + "..."
			} else {
				desc = plain
			}
		}

		render(w, r, "blog_post.html", PageData{
			Title: post.Title + " · " + siteName,
			Site:  cfg,
			Data: map[string]interface{}{
				"post":    post,
				"content": content,
			},
			Current:         "blog",
			MetaDescription: desc,
			MetaKeywords:    post.Tags,
			OGType:          "article",
			CanonicalURL:    canonical,
			JSONLD:          string(ldJSON),
		})
	}
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			b.WriteRune(' ')
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(fmt.Sprint(b.String())), " ")
}
