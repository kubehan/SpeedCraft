package config

import (
	"os"
)

type Config struct {
	Port     string
	DBPath   string
	SiteName string
	SiteDesc string
	SiteURL  string
	AdminPwd string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		DBPath:   getEnv("DB_PATH", "data/speedcraft.db"),
		SiteName: "速创社",
		SiteDesc: "专业 DevOps & 云原生架构咨询 · 让技术驱动业务增长",
		SiteURL:  getEnv("SITE_URL", "http://localhost:8080"),
		AdminPwd: getEnv("ADMIN_PWD", "admin888"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
