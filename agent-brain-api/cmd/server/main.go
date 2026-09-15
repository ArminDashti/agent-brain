package main

import (
	"context"
	"log"
	"time"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/auth"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/config"
	appdb "github.com/ArminDashti/agent-brain/agent-brain-api/internal/db"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/embed"
	httpapi "github.com/ArminDashti/agent-brain/agent-brain-api/internal/http"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/knowledge"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/qdrant"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadDotEnv(".env")
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	sqlDB, err := appdb.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer sqlDB.Close()
	if err := appdb.Migrate(sqlDB, cfg.MigrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hash, err := auth.HashPassword(appdb.DefaultPassword)
	if err != nil {
		log.Fatalf("hash default password: %v", err)
	}
	if err := appdb.SeedDefaultUser(ctx, sqlDB, hash); err != nil {
		log.Fatalf("seed default user: %v", err)
	}
	qd := qdrant.New(cfg.QdrantURL, cfg.QdrantCollection, cfg.EmbeddingDim)
	if err := qd.EnsureCollection(ctx); err != nil {
		log.Printf("warning: qdrant collection: %v", err)
	}
	emb := embed.New(cfg.EmbeddingDim, cfg.EmbeddingURL, cfg.EmbeddingModel, cfg.EmbeddingAPIKey)
	svc := &knowledge.Service{DB: sqlDB, Qdrant: qd, Embed: emb}
	h := httpapi.New(sqlDB, cfg, svc, qd)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(string) bool { return true },
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          12 * time.Hour,
	}))
	r.GET("/health", h.Health)
	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", h.Login)
		authed := api.Group("")
		authed.Use(auth.Middleware(cfg.JWTSecret))
		{
			authed.POST("/knowledge", h.CreateKnowledge)
			authed.GET("/knowledge", h.ListKnowledge)
			authed.POST("/knowledge/search", h.SearchKnowledge)
			authed.GET("/knowledge/:id", h.GetKnowledge)
			authed.DELETE("/knowledge/:id", h.DeleteKnowledge)

			authed.POST("/sessions/ingest", h.IngestSession)
			authed.GET("/sessions", h.ListSessions)
			authed.GET("/sessions/:uuid", h.GetSession)
			authed.GET("/sessions/:uuid/thinking", h.GetSessionThinking)
			authed.GET("/sessions/:uuid/turns", h.GetSessionTurns)
		}
	}
	log.Printf("listening on %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
