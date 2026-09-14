package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/auth"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/config"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/knowledge"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/qdrant"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	db  *sql.DB
	cfg config.Config
	svc *knowledge.Service
	qd  *qdrant.Client
}

func New(db *sql.DB, cfg config.Config, svc *knowledge.Service, qd *qdrant.Client) *Handler {
	return &Handler{db: db, cfg: cfg, svc: svc, qd: qd}
}

func (h *Handler) Health(c *gin.Context) {
	pgOK := h.db.PingContext(c.Request.Context()) == nil
	qdOK := h.qd.Ready(c.Request.Context()) == nil
	status := http.StatusOK
	if !pgOK || !qdOK {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"ok": pgOK && qdOK, "service": "agent-brain-api", "postgres": pgOK, "qdrant": qdOK})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	var id, username, hash string
	err := h.db.QueryRowContext(c.Request.Context(), `SELECT id::text, username, password_hash FROM users WHERE username = $1`, req.Username).Scan(&id, &username, &hash)
	if err != nil || !auth.CheckPassword(hash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := auth.IssueToken(h.cfg.JWTSecret, id, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token issue failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": id, "username": username}})
}

type storeReq struct {
	Kind        string         `json:"kind"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Tags        []string       `json:"tags"`
	Metadata    map[string]any `json:"metadata"`
	StoreTarget string         `json:"store_target"`
}

func (h *Handler) CreateKnowledge(c *gin.Context) {
	var req storeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req.Kind, req.Title, req.Body, req.Tags, req.Metadata, knowledge.StoreTarget(req.StoreTarget))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListKnowledge(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := h.svc.List(c.Request.Context(), c.Query("kind"), c.Query("q"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetKnowledge(c *gin.Context) {
	item, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

type searchReq struct {
	Query string `json:"query"`
	Kind  string `json:"kind"`
	Mode  string `json:"mode"`
	Limit int    `json:"limit"`
}

func (h *Handler) SearchKnowledge(c *gin.Context) {
	var req searchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	items, err := h.svc.Search(c.Request.Context(), req.Query, req.Kind, knowledge.SearchMode(req.Mode), req.Limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) DeleteKnowledge(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
