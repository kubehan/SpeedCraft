package handlers

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"speedcraft/config"
	"speedcraft/models"
)

type sessionData struct {
	Valid     bool
	CSRFToken string
}

var (
	sessions   = make(map[string]*sessionData)
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
	Keyword     string
	ActiveTab   string
}

func AdminLogin(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")
			if username == "admin" && password == cfg.AdminPwd {
				token := generateToken()
				csrfToken := generateToken()
				sessionsMu.Lock()
				sessions[token] = &sessionData{Valid: true, CSRFToken: csrfToken}
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
			render(w, r, "admin/login.html", PageData{
				Title: "登录 · " + cfg.SiteName,
				Site:  cfg,
				Data:  map[string]interface{}{"error": "用户名或密码错误"},
			})
			return
		}
		render(w, r, "admin/login.html", PageData{
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
		session := sessions[cookie.Value]
		sessionsMu.RUnlock()
		if session == nil || !session.Valid {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}

		if r.Method == http.MethodPost {
			formToken := r.FormValue("csrf_token")
			if formToken == "" || formToken != session.CSRFToken {
				http.Error(w, "无效的 CSRF Token", http.StatusForbidden)
				return
			}
		}

		next(w, r)
	}
}

// SessionOnlyMiddleware checks admin session but skips CSRF — for read-only endpoints like preview
func SessionOnlyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_token")
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		sessionsMu.RLock()
		session := sessions[cookie.Value]
		sessionsMu.RUnlock()
		if session == nil || !session.Valid {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func AdminDashboard(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "admin/dashboard.html", PageData{
			Title: "管理后台 · " + cfg.SiteName,
			Site:  cfg,
		})
	})
}

func AdminMessages(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		keyword := r.URL.Query().Get("keyword")
		pageStr := r.URL.Query().Get("page")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize := 20

		messages, total, err := models.GetMessages(status, keyword, page, pageSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		totalPages := (total + pageSize - 1) / pageSize
		if messages == nil {
			messages = []models.Message{}
		}

		render(w, r, "admin/messages.html", PageData{
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
				Keyword:    keyword,
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

func AdminMessageExport(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		messages, err := models.GetAllMessages()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=messages.csv")
		// BOM for Excel UTF-8
		w.Write([]byte{0xEF, 0xBB, 0xBF})

		writer := csv.NewWriter(w)
		writer.Write([]string{"ID", "姓名", "邮箱", "电话", "公司", "服务类型", "预算", "留言", "状态", "时间"})
		for _, m := range messages {
			writer.Write([]string{
				fmt.Sprintf("%d", m.ID),
				m.Name, m.Email, m.Phone, m.Company,
				m.ServiceType, m.Budget, m.Message, m.Status,
				m.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		writer.Flush()
	})
}
