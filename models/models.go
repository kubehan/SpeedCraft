package models

import (
	"database/sql"
	"fmt"
	"speedcraft/database"
	"strings"
	"time"
)

// -------- Message --------
type Message struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Company     string    `json:"company"`
	ServiceType string    `json:"service_type"`
	Budget      string    `json:"budget"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	Notified    int       `json:"notified"`
	CreatedAt   time.Time `json:"created_at"`
}

type MessageRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Company     string `json:"company"`
	ServiceType string `json:"service_type"`
	Budget      string `json:"budget"`
	Message     string `json:"message"`
}

func (r *MessageRequest) Save() (int64, error) {
	result, err := database.DB.Exec(
		`INSERT INTO messages (name, email, phone, company, service_type, budget, message) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.Name, r.Email, r.Phone, r.Company, r.ServiceType, r.Budget, r.Message,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetMessages(status, keyword string, page, pageSize int) ([]Message, int, error) {
	var count int
	query := "SELECT COUNT(*) FROM messages"
	args := []interface{}{}
	conditions := []string{}

	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if keyword != "" {
		conditions = append(conditions, "(name LIKE ? OR email LIKE ? OR company LIKE ? OR service_type LIKE ? OR message LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw, kw)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if err := database.DB.QueryRow(query, args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	q := "SELECT id, name, email, phone, company, service_type, budget, message, status, notified, created_at FROM messages"
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	rows, err := database.DB.Query(q, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Company,
			&m.ServiceType, &m.Budget, &m.Message, &m.Status, &m.Notified, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	return msgs, count, nil
}

func UpdateMessageStatus(id int64, status string) error {
	_, err := database.DB.Exec(
		"UPDATE messages SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, id)
	return err
}

func GetRecentMessagesCount(hours int) int {
	var count int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE created_at > datetime('now', ?)",
		"-"+strings.TrimPrefix(strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(string(rune(hours)), ""), ""), ""), "")+" hours",
	).Scan(&count)
	return count
}

func GetDashboardStats() (totalMsg, pendingMsg int64) {
	database.DB.QueryRow("SELECT COUNT(*) FROM messages").Scan(&totalMsg)
	database.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE status='pending'").Scan(&pendingMsg)
	return
}

func GetAllMessages() ([]Message, error) {
	rows, err := database.DB.Query("SELECT id, name, email, phone, company, service_type, budget, message, status, notified, created_at FROM messages ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var m Message
		rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Company,
			&m.ServiceType, &m.Budget, &m.Message, &m.Status, &m.Notified, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// -------- Service --------
type Service struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	Features    string    `json:"features"`
	Pricing     string    `json:"pricing"`
	SortOrder   int       `json:"sort_order"`
	IsPublished int       `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetServices() ([]Service, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, icon, description, features, pricing, sort_order, is_published, created_at FROM services ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Service
	for rows.Next() {
		var s Service
		rows.Scan(&s.ID, &s.Title, &s.Icon, &s.Description, &s.Features, &s.Pricing, &s.SortOrder, &s.IsPublished, &s.CreatedAt)
		list = append(list, s)
	}
	return list, nil
}

func GetServicesPage(page, pageSize int) ([]Service, int, error) {
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM services").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.DB.Query(
		"SELECT id, title, icon, description, features, pricing, sort_order, is_published, created_at FROM services ORDER BY sort_order ASC, id ASC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Service
	for rows.Next() {
		var s Service
		rows.Scan(&s.ID, &s.Title, &s.Icon, &s.Description, &s.Features, &s.Pricing, &s.SortOrder, &s.IsPublished, &s.CreatedAt)
		list = append(list, s)
	}
	return list, total, nil
}

func GetPublishedServices() ([]Service, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, icon, description, features, pricing, sort_order, is_published, created_at FROM services WHERE is_published=1 ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Service
	for rows.Next() {
		var s Service
		rows.Scan(&s.ID, &s.Title, &s.Icon, &s.Description, &s.Features, &s.Pricing, &s.SortOrder, &s.IsPublished, &s.CreatedAt)
		list = append(list, s)
	}
	return list, nil
}

func GetService(id int64) (*Service, error) {
	var s Service
	err := database.DB.QueryRow(
		"SELECT id, title, icon, description, features, pricing, sort_order, is_published, created_at FROM services WHERE id=?", id,
	).Scan(&s.ID, &s.Title, &s.Icon, &s.Description, &s.Features, &s.Pricing, &s.SortOrder, &s.IsPublished, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveService(s *Service) (int64, error) {
	if s.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE services SET title=?, icon=?, description=?, features=?, pricing=?, sort_order=?, is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			s.Title, s.Icon, s.Description, s.Features, s.Pricing, s.SortOrder, s.IsPublished, s.ID)
		return s.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO services (title, icon, description, features, pricing, sort_order, is_published) VALUES (?, ?, ?, ?, ?, ?, ?)",
		s.Title, s.Icon, s.Description, s.Features, s.Pricing, s.SortOrder, s.IsPublished)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteService(id int64) error {
	_, err := database.DB.Exec("DELETE FROM services WHERE id=?", id)
	return err
}

// -------- Open Source Project --------
type OpenSourceProject struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	GithubURL   string    `json:"github_url"`
	Stars       int       `json:"stars"`
	Language    string    `json:"language"`
	LicenseType string    `json:"license_type"`
	IsFeatured  int       `json:"is_featured"`
	SortOrder   int       `json:"sort_order"`
	IsPublished int       `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetOpenSourceProjects() ([]OpenSourceProject, error) {
	rows, err := database.DB.Query(
		"SELECT id, name, description, url, github_url, stars, language, license_type, is_featured, sort_order, is_published, created_at FROM open_source_projects ORDER BY sort_order ASC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []OpenSourceProject
	for rows.Next() {
		var p OpenSourceProject
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.URL, &p.GithubURL, &p.Stars, &p.Language, &p.LicenseType, &p.IsFeatured, &p.SortOrder, &p.IsPublished, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetOpenSourceProjectsPage(page, pageSize int) ([]OpenSourceProject, int, error) {
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM open_source_projects").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.DB.Query(
		"SELECT id, name, description, url, github_url, stars, language, license_type, is_featured, sort_order, is_published, created_at FROM open_source_projects ORDER BY sort_order ASC, id DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []OpenSourceProject
	for rows.Next() {
		var p OpenSourceProject
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.URL, &p.GithubURL, &p.Stars, &p.Language, &p.LicenseType, &p.IsFeatured, &p.SortOrder, &p.IsPublished, &p.CreatedAt)
		list = append(list, p)
	}
	return list, total, nil
}

func GetPublishedOpenSource() ([]OpenSourceProject, error) {
	rows, err := database.DB.Query(
		"SELECT id, name, description, url, github_url, stars, language, license_type, is_featured, sort_order, is_published, created_at FROM open_source_projects WHERE is_published=1 ORDER BY is_featured DESC, sort_order ASC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []OpenSourceProject
	for rows.Next() {
		var p OpenSourceProject
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.URL, &p.GithubURL, &p.Stars, &p.Language, &p.LicenseType, &p.IsFeatured, &p.SortOrder, &p.IsPublished, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetOpenSourceProject(id int64) (*OpenSourceProject, error) {
	var p OpenSourceProject
	err := database.DB.QueryRow(
		"SELECT id, name, description, url, github_url, stars, language, license_type, is_featured, sort_order, is_published, created_at FROM open_source_projects WHERE id=?", id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.URL, &p.GithubURL, &p.Stars, &p.Language, &p.LicenseType, &p.IsFeatured, &p.SortOrder, &p.IsPublished, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func SaveOpenSourceProject(p *OpenSourceProject) (int64, error) {
	if p.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE open_source_projects SET name=?, description=?, url=?, github_url=?, stars=?, language=?, license_type=?, is_featured=?, sort_order=?, is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			p.Name, p.Description, p.URL, p.GithubURL, p.Stars, p.Language, p.LicenseType, p.IsFeatured, p.SortOrder, p.IsPublished, p.ID)
		return p.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO open_source_projects (name, description, url, github_url, stars, language, license_type, is_featured, sort_order, is_published) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Description, p.URL, p.GithubURL, p.Stars, p.Language, p.LicenseType, p.IsFeatured, p.SortOrder, p.IsPublished)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteOpenSourceProject(id int64) error {
	_, err := database.DB.Exec("DELETE FROM open_source_projects WHERE id=?", id)
	return err
}

// -------- Navigation --------
type NavItem struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	URL       string `json:"url"`
	Icon      string `json:"icon"`
	ParentID  int64  `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
	Published int    `json:"published"`
}

func GetNavigation() ([]NavItem, error) {
	rows, err := database.DB.Query(
		"SELECT id, label, url, icon, parent_id, sort_order, is_published FROM navigation_items ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []NavItem
	for rows.Next() {
		var n NavItem
		rows.Scan(&n.ID, &n.Label, &n.URL, &n.Icon, &n.ParentID, &n.SortOrder, &n.Published)
		list = append(list, n)
	}
	return list, nil
}

func GetPublishedNavigation() ([]NavItem, error) {
	rows, err := database.DB.Query(
		"SELECT id, label, url, icon, parent_id, sort_order, is_published FROM navigation_items WHERE is_published=1 ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []NavItem
	for rows.Next() {
		var n NavItem
		rows.Scan(&n.ID, &n.Label, &n.URL, &n.Icon, &n.ParentID, &n.SortOrder, &n.Published)
		list = append(list, n)
	}
	return list, nil
}

func SaveNavItem(n *NavItem) (int64, error) {
	if n.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE navigation_items SET label=?, url=?, icon=?, parent_id=?, sort_order=?, is_published=? WHERE id=?",
			n.Label, n.URL, n.Icon, n.ParentID, n.SortOrder, n.Published, n.ID)
		return n.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO navigation_items (label, url, icon, parent_id, sort_order, is_published) VALUES (?, ?, ?, ?, ?, ?)",
		n.Label, n.URL, n.Icon, n.ParentID, n.SortOrder, n.Published)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteNavItem(id int64) error {
	_, err := database.DB.Exec("DELETE FROM navigation_items WHERE id=?", id)
	return err
}

func ReorderNavigation(ids []int64) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		_, err = tx.Exec("UPDATE navigation_items SET sort_order=? WHERE id=?", i, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func TogglePublish(table string, id int64) (int, error) {
	var current int
	err := database.DB.QueryRow(fmt.Sprintf("SELECT is_published FROM %s WHERE id=?", table), id).Scan(&current)
	if err != nil {
		return 0, err
	}
	newVal := 0
	if current == 0 {
		newVal = 1
	}
	// Some tables don't have updated_at column (navigation_items, skills)
	hasUpdatedAt := map[string]bool{
		"services": true, "blog_posts": true, "projects": true,
		"open_source_projects": true,
	}
	query := fmt.Sprintf("UPDATE %s SET is_published=? WHERE id=?", table)
	if hasUpdatedAt[table] {
		query = fmt.Sprintf("UPDATE %s SET is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", table)
	}
	_, err = database.DB.Exec(query, newVal, id)
	return newVal, err
}

// -------- Site Settings --------
func GetSetting(key string) string {
	var val string
	database.DB.QueryRow("SELECT setting_value FROM site_settings WHERE setting_key=?", key).Scan(&val)
	return val
}

func GetAllSettings() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT setting_key, setting_value FROM site_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	return settings, nil
}

func SaveSetting(key, value string) error {
	_, err := database.DB.Exec(
		"INSERT INTO site_settings (setting_key, setting_value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(setting_key) DO UPDATE SET setting_value=?, updated_at=CURRENT_TIMESTAMP",
		key, value, value)
	return err
}

func SaveSettings(settings map[string]string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for k, v := range settings {
		_, err := tx.Exec(
			"INSERT INTO site_settings (setting_key, setting_value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(setting_key) DO UPDATE SET setting_value=?, updated_at=CURRENT_TIMESTAMP",
			k, v, v)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// -------- Skills --------
type Skill struct {
	ID        int64     `json:"id"`
	Icon      string    `json:"icon"`
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	SortOrder int       `json:"sort_order"`
	Published int       `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

func GetSkills() ([]Skill, error) {
	rows, err := database.DB.Query(
		"SELECT id, icon, name, level, category, sort_order, is_published, created_at FROM skills ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Skill
	for rows.Next() {
		var s Skill
		rows.Scan(&s.ID, &s.Icon, &s.Name, &s.Level, &s.Category, &s.SortOrder, &s.Published, &s.CreatedAt)
		list = append(list, s)
	}
	return list, nil
}

func GetPublishedSkills() ([]Skill, error) {
	rows, err := database.DB.Query(
		"SELECT id, icon, name, level, category, sort_order, is_published, created_at FROM skills WHERE is_published=1 ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Skill
	for rows.Next() {
		var s Skill
		rows.Scan(&s.ID, &s.Icon, &s.Name, &s.Level, &s.Category, &s.SortOrder, &s.Published, &s.CreatedAt)
		list = append(list, s)
	}
	return list, nil
}

func SaveSkill(s *Skill) (int64, error) {
	if s.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE skills SET icon=?, name=?, level=?, category=?, sort_order=?, is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			s.Icon, s.Name, s.Level, s.Category, s.SortOrder, s.Published, s.ID)
		return s.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO skills (icon, name, level, category, sort_order, is_published) VALUES (?, ?, ?, ?, ?, ?)",
		s.Icon, s.Name, s.Level, s.Category, s.SortOrder, s.Published)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteSkill(id int64) error {
	_, err := database.DB.Exec("DELETE FROM skills WHERE id=?", id)
	return err
}

// -------- Project --------
type Project struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	TechStack   string    `json:"tech_stack"`
	ImageURL    string    `json:"image_url"`
	ClientName  string    `json:"client_name"`
	ClientURL   string    `json:"client_url"`
	IsPublished int       `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
}

// -------- Project (extended) --------
func GetProject(id int64) (*Project, error) {
	var p Project
	err := database.DB.QueryRow(
		"SELECT id, title, slug, summary, content, category, tech_stack, image_url, client_name, client_url, is_published, created_at FROM projects WHERE id=?", id,
	).Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Content, &p.Category, &p.TechStack, &p.ImageURL, &p.ClientName, &p.ClientURL, &p.IsPublished, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetAllProjects() ([]Project, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, slug, summary, category, tech_stack, image_url, client_name, client_url, is_published, created_at FROM projects ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Category, &p.TechStack, &p.ImageURL, &p.ClientName, &p.ClientURL, &p.IsPublished, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetAllProjectsPage(page, pageSize int) ([]Project, int, error) {
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM projects").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.DB.Query(
		"SELECT id, title, slug, summary, category, tech_stack, image_url, client_name, client_url, is_published, created_at FROM projects ORDER BY created_at DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Category, &p.TechStack, &p.ImageURL, &p.ClientName, &p.ClientURL, &p.IsPublished, &p.CreatedAt)
		list = append(list, p)
	}
	return list, total, nil
}

func GetPublishedProjects() ([]Project, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, slug, summary, category, tech_stack, image_url, client_name, client_url, is_published, created_at FROM projects WHERE is_published=1 ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Category, &p.TechStack, &p.ImageURL, &p.ClientName, &p.ClientURL, &p.IsPublished, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func SaveProject(p *Project) (int64, error) {
	if p.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE projects SET title=?, slug=?, summary=?, content=?, category=?, tech_stack=?, image_url=?, client_name=?, client_url=?, is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			p.Title, p.Slug, p.Summary, p.Content, p.Category, p.TechStack, p.ImageURL, p.ClientName, p.ClientURL, p.IsPublished, p.ID)
		return p.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO projects (title, slug, summary, content, category, tech_stack, image_url, client_name, client_url, is_published) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Title, p.Slug, p.Summary, p.Content, p.Category, p.TechStack, p.ImageURL, p.ClientName, p.ClientURL, p.IsPublished)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteProject(id int64) error {
	_, err := database.DB.Exec("DELETE FROM projects WHERE id=?", id)
	return err
}

// -------- Blog (extended) --------
type BlogPost struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	Tags        string    `json:"tags"`
	IsPublished int       `json:"is_published"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetAllPosts() ([]BlogPost, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, slug, summary, content_type, tags, is_published, views, created_at FROM blog_posts ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BlogPost
	for rows.Next() {
		var p BlogPost
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.ContentType, &p.Tags, &p.IsPublished, &p.Views, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetAllPostsPage(page, pageSize int) ([]BlogPost, int, error) {
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM blog_posts").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.DB.Query(
		"SELECT id, title, slug, summary, content_type, tags, is_published, views, created_at FROM blog_posts ORDER BY created_at DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []BlogPost
	for rows.Next() {
		var p BlogPost
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.ContentType, &p.Tags, &p.IsPublished, &p.Views, &p.CreatedAt)
		list = append(list, p)
	}
	return list, total, nil
}

func GetPublishedPosts() ([]BlogPost, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, slug, summary, content_type, tags, is_published, views, created_at FROM blog_posts WHERE is_published=1 ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BlogPost
	for rows.Next() {
		var p BlogPost
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.ContentType, &p.Tags, &p.IsPublished, &p.Views, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetPost(id int64) (*BlogPost, error) {
	var p BlogPost
	err := database.DB.QueryRow(
		"SELECT id, title, slug, summary, content, content_type, tags, is_published, views, created_at FROM blog_posts WHERE id=?", id,
	).Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Content, &p.ContentType, &p.Tags, &p.IsPublished, &p.Views, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetPostBySlug(slug string) (*BlogPost, error) {
	var p BlogPost
	err := database.DB.QueryRow(
		"SELECT id, title, slug, summary, content, content_type, tags, is_published, views, created_at FROM blog_posts WHERE slug=? AND is_published=1", slug,
	).Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Content, &p.ContentType, &p.Tags, &p.IsPublished, &p.Views, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func SavePost(p *BlogPost) (int64, error) {
	if p.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE blog_posts SET title=?, slug=?, summary=?, content=?, content_type=?, tags=?, is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			p.Title, p.Slug, p.Summary, p.Content, p.ContentType, p.Tags, p.IsPublished, p.ID)
		return p.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO blog_posts (title, slug, summary, content, content_type, tags, is_published) VALUES (?, ?, ?, ?, ?, ?, ?)",
		p.Title, p.Slug, p.Summary, p.Content, p.ContentType, p.Tags, p.IsPublished)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeletePost(id int64) error {
	_, err := database.DB.Exec("DELETE FROM blog_posts WHERE id=?", id)
	return err
}

func IncrementPostViews(id int64) {
	database.DB.Exec("UPDATE blog_posts SET views = views + 1 WHERE id=?", id)
}

// -------- Tag --------
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

func GetAllTags() ([]Tag, error) {
	rows, err := database.DB.Query("SELECT id, name, slug, created_at FROM tags ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Tag
	for rows.Next() {
		var t Tag
		rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
		list = append(list, t)
	}
	return list, nil
}

func CreateTag(name, slug string) (int64, error) {
	result, err := database.DB.Exec("INSERT INTO tags (name, slug) VALUES (?, ?)", name, slug)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteTag(id int64) error {
	_, err := database.DB.Exec("DELETE FROM tags WHERE id=?", id)
	return err
}

// -------- Page --------
type Page struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	IsPublished int       `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func GetAllPages() ([]Page, error) {
	rows, err := database.DB.Query("SELECT id, title, slug, content, content_type, is_published, created_at, updated_at FROM pages ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Page
	for rows.Next() {
		var p Page
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.ContentType, &p.IsPublished, &p.CreatedAt, &p.UpdatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetPublishedPages() ([]Page, error) {
	rows, err := database.DB.Query("SELECT id, title, slug, content, content_type, is_published, created_at, updated_at FROM pages WHERE is_published=1 ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Page
	for rows.Next() {
		var p Page
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.ContentType, &p.IsPublished, &p.CreatedAt, &p.UpdatedAt)
		list = append(list, p)
	}
	return list, nil
}

func GetPageBySlug(slug string) (*Page, error) {
	var p Page
	err := database.DB.QueryRow("SELECT id, title, slug, content, content_type, is_published, created_at, updated_at FROM pages WHERE slug=? AND is_published=1", slug).
		Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.ContentType, &p.IsPublished, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func SavePage(p *Page) (int64, error) {
	if p.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE pages SET title=?, slug=?, content=?, content_type=?, is_published=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			p.Title, p.Slug, p.Content, p.ContentType, p.IsPublished, p.ID)
		return p.ID, err
	}
	result, err := database.DB.Exec(
		"INSERT INTO pages (title, slug, content, content_type, is_published) VALUES (?, ?, ?, ?, ?)",
		p.Title, p.Slug, p.Content, p.ContentType, p.IsPublished)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeletePage(id int64) error {
	_, err := database.DB.Exec("DELETE FROM pages WHERE id=?", id)
	return err
}

func BatchAction(table, action string, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	validTables := map[string]bool{
		"services": true, "blog_posts": true, "projects": true,
		"open_source_projects": true, "skills": true, "navigation_items": true,
	}
	if !validTables[table] {
		return fmt.Errorf("invalid table: %s", table)
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := ""
	switch action {
	case "publish":
		query = fmt.Sprintf("UPDATE %s SET is_published=1 WHERE id IN (%s)", table, strings.Join(placeholders, ","))
	case "unpublish":
		query = fmt.Sprintf("UPDATE %s SET is_published=0 WHERE id IN (%s)", table, strings.Join(placeholders, ","))
	case "delete":
		query = fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)", table, strings.Join(placeholders, ","))
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
	_, err := database.DB.Exec(query, args...)
	return err
}
