package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL, Collection string
	Dim int
	HTTP *http.Client
}

func New(baseURL, collection string, dim int) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Collection: collection, Dim: dim, HTTP: &http.Client{Timeout: 20 * time.Second}}
}

func (c *Client) Ready(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/readyz", nil)
	if err != nil { return err }
	res, err := c.HTTP.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	if res.StatusCode >= 300 { return fmt.Errorf("qdrant readyz %d", res.StatusCode) }
	return nil
}

func (c *Client) EnsureCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections/%s", c.BaseURL, c.Collection)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := c.HTTP.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	if res.StatusCode == 200 { return nil }
	body, _ := json.Marshal(map[string]any{"vectors": map[string]any{"size": c.Dim, "distance": "Cosine"}})
	createReq, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRes, err := c.HTTP.Do(createReq)
	if err != nil { return err }
	defer createRes.Body.Close()
	raw, _ := io.ReadAll(createRes.Body)
	if createRes.StatusCode >= 300 { return fmt.Errorf("create collection %d: %s", createRes.StatusCode, string(raw)) }
	return nil
}

type PointPayload struct {
	ID, Kind, Title, Snippet string
	Tags []string `json:"tags"`
}

func (c *Client) Upsert(ctx context.Context, id string, vector []float32, payload PointPayload) error {
	body, _ := json.Marshal(map[string]any{"points": []map[string]any{{"id": id, "vector": vector, "payload": payload}}})
	url := fmt.Sprintf("%s/collections/%s/points?wait=true", c.BaseURL, c.Collection)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 { return fmt.Errorf("upsert %d: %s", res.StatusCode, string(raw)) }
	return nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	body, _ := json.Marshal(map[string]any{"points": []string{id}})
	url := fmt.Sprintf("%s/collections/%s/points/delete?wait=true", c.BaseURL, c.Collection)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 { return fmt.Errorf("delete point %d: %s", res.StatusCode, string(raw)) }
	return nil
}

type SearchHit struct {
	ID string
	Score float64
	Payload map[string]any
}

func (c *Client) Search(ctx context.Context, vector []float32, limit int, kind string) ([]SearchHit, error) {
	if limit <= 0 { limit = 20 }
	payload := map[string]any{"vector": vector, "limit": limit, "with_payload": true}
	if kind != "" {
		payload["filter"] = map[string]any{"must": []map[string]any{{"key": "kind", "match": map[string]any{"value": kind}}}}
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/collections/%s/points/search", c.BaseURL, c.Collection)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil { return nil, err }
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 { return nil, fmt.Errorf("search %d: %s", res.StatusCode, string(raw)) }
	var parsed struct {
		Result []struct {
			ID any `json:"id"`
			Score float64 `json:"score"`
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil { return nil, err }
	out := make([]SearchHit, 0, len(parsed.Result))
	for _, r := range parsed.Result {
		out = append(out, SearchHit{ID: fmt.Sprint(r.ID), Score: r.Score, Payload: r.Payload})
	}
	return out, nil
}
