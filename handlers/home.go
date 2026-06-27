package handlers

import (
	"net/http"
	"speedcraft/config"
	"speedcraft/models"
)

type HomePageData struct {
	Settings map[string]string
	Skills   []models.Skill
	Services []models.Service
}

func Home(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		settings, _ := models.GetAllSettings()
		skills, _ := models.GetPublishedSkills()
		services, _ := models.GetPublishedServices()
		if skills == nil {
			skills = []models.Skill{}
		}
		if services == nil {
			services = []models.Service{}
		}
		render(w, "index.html", PageData{
			Title:   models.GetSetting("site_name") + " · 把创意快速变成产品",
			Site:    cfg,
			Data:    HomePageData{Settings: settings, Skills: skills, Services: services},
			Current: "home",
		})
	}
}
