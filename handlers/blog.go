package handlers

import (
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
		render(w, "blog.html", PageData{
			Title:   "博客 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    posts,
			Current: "blog",
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

		render(w, "blog_post.html", PageData{
			Title:   post.Title + " · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    map[string]interface{}{
				"post":    post,
				"content": content,
			},
			Current: "blog",
		})
	}
}
