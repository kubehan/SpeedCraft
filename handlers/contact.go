package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"speedcraft/config"
	"speedcraft/models"
)

func About(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, _ := models.GetAllSettings()
		services, _ := models.GetPublishedServices()
		if services == nil {
			services = []models.Service{}
		}
		render(w, r, "about.html", PageData{
			Title:   "关于 · " + models.GetSetting("site_name"),
			Site:    cfg,
			Data: map[string]interface{}{
				"settings": settings,
				"services": services,
			},
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
		req.Name, req.Email, req.Phone, req.Company,
		serviceLabel(req.ServiceType), budgetLabel(req.Budget), req.Message,
	)

	body, _ := json.Marshal(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	})

	http.Post(webhook, "application/json", bytes.NewReader(body))
}

// serviceLabel converts legacy short codes to human-readable labels.
// New submissions already store full titles like "MVP 极速开发", so we only need to map
// the legacy codes that some old records / forms might still produce.
func serviceLabel(v string) string {
	if v == "" {
		return "-"
	}
	legacy := map[string]string{
		"mvp":             "MVP 极速开发",
		"cloud":           "云架构设计",
		"cicd":            "CI/CD 流水线",
		"k8s":             "Kubernetes 运维",
		"monitoring":      "监控体系建设",
		"iac":             "基础设施即代码",
		"code":            "代码开发",
		"design":          "系统设计",
		"solution":        "方案设计",
		"security":        "安全等保 & 基线核查",
		"xc":              "信创适配",
		"cloud-migration": "服务上云 & 托管维护",
		"consulting":      "DevOps 咨询与培训",
		"other":           "其他",
	}
	if label, ok := legacy[v]; ok {
		return label
	}
	return v
}

// budgetLabel converts budget short codes to human-readable labels.
func budgetLabel(v string) string {
	if v == "" {
		return "-"
	}
	legacy := map[string]string{
		"lt5k":      "¥5,000 以下",
		"5k-20k":    "¥5,000 - ¥20,000",
		"20k-50k":   "¥20,000 - ¥50,000",
		"50k-100k":  "¥50,000 - ¥100,000",
		"gt100k":    "¥100,000 以上",
		"unknown":   "暂不确定",
	}
	if label, ok := legacy[v]; ok {
		return label
	}
	return v
}

// SendGenericNotification sends a notification through configured channels (webhook + email)
// with a generic title and content. Used for non-message events like friend link applications.
func SendGenericNotification(title, content string) {
	webhook := models.GetSetting("wechat_webhook")
	notifyEmail := models.GetSetting("notify_email")

	if webhook != "" {
		body, _ := json.Marshal(map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": "=== " + title + " ===\n" + content,
			},
		})
		http.Post(webhook, "application/json", bytes.NewReader(body))
	}

	if notifyEmail != "" {
		host := models.GetSetting("smtp_host")
		port := models.GetSetting("smtp_port")
		user := models.GetSetting("smtp_user")
		pass := models.GetSetting("smtp_pass")
		if host == "" || user == "" {
			return
		}
		subject := "[速创社] " + title
		msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
			user, notifyEmail, subject, content)
		addr := fmt.Sprintf("%s:%s", host, port)
		auth := smtp.PlainAuth("", user, pass, host)
		if err := smtp.SendMail(addr, auth, user, []string{notifyEmail}, []byte(msg)); err != nil {
			fmt.Printf("[EMAIL] 发送失败: %v\n", err)
		}
	}
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
		req.Name, req.Email, req.Phone, req.Company,
		serviceLabel(req.ServiceType), budgetLabel(req.Budget), req.Message,
	)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", user, to, subject, body)

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	if err := smtp.SendMail(addr, auth, user, []string{to}, []byte(msg)); err != nil {
		fmt.Printf("[EMAIL] 发送失败: %v\n", err)
	}
}

func AdminTestNotification(cfg *config.Config) http.HandlerFunc {
	return AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
			return
		}

		testReq := models.MessageRequest{
			Name:    "测试用户",
			Email:   "test@speedcraft.dev",
			Message: "这是一条来自速创社管理后台的测试通知消息。如果收到此消息，说明通知配置正确。",
		}

		errs := []string{}
		webhook := models.GetSetting("wechat_webhook")
		notifyEmail := models.GetSetting("notify_email")

		if webhook != "" {
			sendWechatWebhook(webhook, testReq)
		}
		if notifyEmail != "" {
			sendEmailNotification(notifyEmail, testReq)
		}
		if webhook == "" && notifyEmail == "" {
			errs = append(errs, "请先配置 Webhook 或 SMTP")
		}

		if len(errs) > 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"error": strings.Join(errs, "; ")})
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "测试通知已发送，请检查"})
	})
}
