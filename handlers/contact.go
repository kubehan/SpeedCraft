package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"speedcraft/config"
	"speedcraft/models"
)

func About(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, _ := models.GetAllSettings()
		render(w, "about.html", PageData{
			Title:   "关于 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data:    settings,
			Current: "about",
		})
	}
}

func SubmitMessage(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
			return
		}

		req := models.MessageRequest{
			Name:        r.FormValue("name"),
			Email:       r.FormValue("email"),
			Phone:       r.FormValue("phone"),
			Company:     r.FormValue("company"),
			ServiceType: r.FormValue("service_type"),
			Budget:      r.FormValue("budget"),
			Message:     r.FormValue("message"),
		}

		if req.Name == "" || req.Email == "" || req.Message == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "请填写必填字段"})
			return
		}

		id, err := req.Save()
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "提交失败，请稍后重试"})
			return
		}

		go sendNotifications(req)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"id":      id,
			"message": "提交成功！我会在24小时内与您联系。",
		})
	}
}

func sendNotifications(req models.MessageRequest) {
	webhook := models.GetSetting("wechat_webhook")
	notifyEmail := models.GetSetting("notify_email")

	if webhook != "" {
		sendWechatWebhook(webhook, req)
	}
	if notifyEmail != "" {
		sendEmailNotification(notifyEmail, req)
	}
}

func sendWechatWebhook(webhook string, req models.MessageRequest) {
	content := fmt.Sprintf(
		"=== 速创社 新咨询 ===\n📌 姓名: %s\n📧 邮箱: %s\n📞 电话: %s\n🏢 公司: %s\n🔧 服务: %s\n💰 预算: %s\n📝 需求: %s",
		req.Name, req.Email, req.Phone, req.Company, req.ServiceType, req.Budget, req.Message,
	)

	body, _ := json.Marshal(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	})

	http.Post(webhook, "application/json", bytes.NewReader(body))
}

func sendEmailNotification(to string, req models.MessageRequest) {
	host := models.GetSetting("smtp_host")
	port := models.GetSetting("smtp_port")
	user := models.GetSetting("smtp_user")
	pass := models.GetSetting("smtp_pass")

	if host == "" || user == "" {
		return
	}

	subject := fmt.Sprintf("[速创社] 新咨询来自 %s", req.Name)
	body := fmt.Sprintf(
		"姓名: %s\n邮箱: %s\n电话: %s\n公司: %s\n服务类型: %s\n预算: %s\n\n需求描述:\n%s",
		req.Name, req.Email, req.Phone, req.Company, req.ServiceType, req.Budget, req.Message,
	)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", user, to, subject, body)

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	if err := smtp.SendMail(addr, auth, user, []string{to}, []byte(msg)); err != nil {
		fmt.Printf("[EMAIL] 发送失败: %v\n", err)
	}
}
