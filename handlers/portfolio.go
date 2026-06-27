package handlers

import (
	"net/http"
	"speedcraft/config"
	"speedcraft/models"
)

func Portfolio(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, _ := models.GetPublishedProjects()
		if projects == nil {
			projects = []models.Project{}
		}
		render(w, r, "portfolio.html", PageData{
			Title:   "案例 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    projects,
			Current: "portfolio",
		})
	}
}
