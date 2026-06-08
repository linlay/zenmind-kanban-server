package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr           string
	DatabasePath   string
	AllowedOrigins []string
	Token          string
}

func Load() Config {
	return Config{
		Addr:           env("ZENMIND_KANBAN_ADDR", ":8080"),
		DatabasePath:   env("ZENMIND_KANBAN_DB", "./data/kanban.db"),
		AllowedOrigins: csvEnv("ZENMIND_KANBAN_ALLOWED_ORIGINS", "*"),
		Token:          strings.TrimSpace(os.Getenv("ZENMIND_KANBAN_TOKEN")),
	}
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func csvEnv(key string, fallback string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		value = fallback
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	if len(items) == 0 {
		return []string{"*"}
	}
	return items
}
