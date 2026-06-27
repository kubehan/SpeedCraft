package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"
	"speedcraft/config"
	"speedcraft/models"
)

var (
	sessions   = make(map[string]bool)
	sessionsMu sync.RWMutex
)

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type AdminData struct {
	Messages    []models.Message
	TotalCount  int
	Page        int
	PageSize    int
	TotalPages  int
	Status      string
	ActiveTab   string
}

func AdminLogin(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")
			if username == "admin" && password == cfg.AdminPwd {
				token := generateToken()
				sessionsMu.Lock()
				sessions[token] = true
				sessionsMu.Unlock()

				http.SetCookie(w, &http.Cookie{
					Name:     "admin_token",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
					MaxAge:   86400 * 7,
				})
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
			render(w, "admin/login.html", PageData{
				Title: "登录 · " + cfg.SiteName,
				Site:  cfg,
				Data:  map[string]interface{}{"error": "用户名或密码错误"},
			})
			return
		}
		render(w, "admin/login.html", PageData{
			Title: "登录 · " + cfg.SiteName,
			Site:  cfg,
			Data:  map[string]interface{}{"error": ""},
		})
	}
}

func AdminLogout(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("admin_token"); err == nil {
			sessionsMu.Lock()
			delete(sessions, cookie.Value)
			sessionsMu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "admin_token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	}
}

func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_token")
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		sessionsMu.RLock()
		valid := sessions[cookie.Value]
		sessionsMu.RUnlock()
		if !valid {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func AdminDashboard(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		render(w, "admin/dashboard.html", PageData{
			Title: "管理后台 · " + cfg.SiteName,
			Site:  cfg,
		})
	})
}

func AdminMessages(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		pageStr := r.URL.Query().Get("page")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize := 20

		messages, total, err := models.GetMessages(status, page, pageSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		totalPages := (total + pageSize - 1) / pageSize
		if messages == nil {
			messages = []models.Message{}
		}

		render(w, "admin/messages.html", PageData{
			Title:   "留言管理 · " + cfg.SiteName,
			Site:    cfg,
			Current: "messages",
			Data: AdminData{
				Messages:   messages,
				TotalCount: total,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: totalPages,
				Status:     status,
				ActiveTab:  "messages",
			},
		})
	})
}

func AdminUpdateMessage(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.FormValue("id")
		status := r.FormValue("status")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		if err := models.UpdateMessageStatus(id, status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	})
}
