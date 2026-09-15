package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL  string
	Username string
	Password string
	HTTP     *http.Client
	token    string
}

type IngestResult struct {
	OK           bool     `json:"ok"`
	UUID         string   `json:"uuid"`
	TurnCount    int      `json:"turn_count"`
	ContextPct   *float64 `json:"context_pct"`
	InputTokens  *int64   `json:"input_tokens"`
	OutputTokens *int64   `json:"output_tokens"`
	Error        string   `json:"error,omitempty"`
}

func New(baseURL, username, password string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8220"
	}
	if username == "" {
		username = "armin"
	}
	if password == "" {
		password = "dopadopa123"
	}
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		HTTP:     &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Login() error {
	body, _ := json.Marshal(map[string]string{
		"username": c.Username,
		"password": c.Password,
	})
	res, err := c.HTTP.Post(c.BaseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("login %d: %s", res.StatusCode, truncate(raw))
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("login: empty token")
	}
	c.token = out.Token
	return nil
}

func (c *Client) IngestSession(extractJSON []byte) (*IngestResult, error) {
	if c.token == "" {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}
	res, err := c.doIngest(extractJSON)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized {
		if err := c.Login(); err != nil {
			return nil, err
		}
		res.Body.Close()
		res, err = c.doIngest(extractJSON)
		if err != nil {
			return nil, err
		}
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("ingest %d: %s", res.StatusCode, truncate(raw))
	}
	var out IngestResult
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) doIngest(extractJSON []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/v1/sessions/ingest", bytes.NewReader(extractJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.HTTP.Do(req)
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 500 {
		return s[:500]
	}
	return s
}
