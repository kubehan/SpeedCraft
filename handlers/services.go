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
		render(w, "services.html", PageData{
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
		render(w, "opensource.html", PageData{
			Title:   "开源 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    list,
			Current: "opensource",
		})
	}
}
