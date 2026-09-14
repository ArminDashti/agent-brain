package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type StubEmbedder struct{ Dim int }

func (s StubEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	dim := s.Dim
	if dim <= 0 {
		dim = 64
	}
	vec := make([]float32, dim)
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(text))))
	seed := h.Sum64()
	for i := 0; i < dim; i++ {
		seed = seed*6364136223846793005 + 1
		v := float32(int64(seed>>33)%2001-1000) / 1000.0
		vec[i] = v
	}
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	norm := math.Sqrt(sum)
	if norm > 0 {
		for i := range vec {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}
	return vec, nil
}

type HTTPEmbedder struct {
	URL, Model, APIKey string
	Dim                int
	Client             *http.Client
	Fallback           Embedder
}

func (h HTTPEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if h.URL == "" {
		if h.Fallback != nil {
			return h.Fallback.Embed(ctx, text)
		}
		return StubEmbedder{Dim: h.Dim}.Embed(ctx, text)
	}
	body, _ := json.Marshal(map[string]any{"model": h.Model, "input": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(h.URL, "/")+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.APIKey)
	}
	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		if h.Fallback != nil {
			return h.Fallback.Embed(ctx, text)
		}
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		if h.Fallback != nil {
			return h.Fallback.Embed(ctx, text)
		}
		return nil, fmt.Errorf("embedding HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return parsed.Data[0].Embedding, nil
}

func New(dim int, url, model, apiKey string) Embedder {
	stub := StubEmbedder{Dim: dim}
	if url == "" {
		return stub
	}
	return HTTPEmbedder{URL: url, Model: model, APIKey: apiKey, Dim: dim, Fallback: stub}
}
