package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

type Config struct {
	Addr             string
	SQLitePath       string
	JWTSecret        string
	MigrationsDir    string
	QdrantURL        string
	QdrantCollection string
	EmbeddingDim     int
	EmbeddingURL     string
	EmbeddingModel   string
	EmbeddingAPIKey  string
}

func Load() Config {
	return Config{
		Addr:             envOr("ADDR", ":8220"),
		SQLitePath:       envOr("SQLITE_PATH", "./data/agent_brain.db"),
		JWTSecret:        envOr("JWT_SECRET", "dev-agent-brain-jwt-change-me"),
		MigrationsDir:    envOr("MIGRATIONS_DIR", "migrations"),
		QdrantURL:        envOr("QDRANT_URL", "http://127.0.0.1:6335"),
		QdrantCollection: envOr("QDRANT_COLLECTION", "agent_brain_knowledge"),
		EmbeddingDim:     envInt("EMBEDDING_DIM", 64),
		EmbeddingURL:     os.Getenv("EMBEDDING_URL"),
		EmbeddingModel:   envOr("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingAPIKey:  os.Getenv("EMBEDDING_API_KEY"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
