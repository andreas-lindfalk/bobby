package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "nomic-embed-text"
	defaultDimensions  = 768
)

// OllamaEmbedder implements Embedder using Ollama's local API.
type OllamaEmbedder struct {
	url        string
	model      string
	dimensions int
	client     *http.Client
}

// OllamaOption configures the OllamaEmbedder.
type OllamaOption func(*OllamaEmbedder)

// WithOllamaURL sets the Ollama API URL.
func WithOllamaURL(url string) OllamaOption {
	return func(o *OllamaEmbedder) {
		o.url = url
	}
}

// WithOllamaModel sets the embedding model.
func WithOllamaModel(model string) OllamaOption {
	return func(o *OllamaEmbedder) {
		o.model = model
	}
}

// WithDimensions sets the expected embedding dimensions.
func WithDimensions(dims int) OllamaOption {
	return func(o *OllamaEmbedder) {
		o.dimensions = dims
	}
}

// NewOllamaEmbedder creates a new Ollama-based embedder.
func NewOllamaEmbedder(opts ...OllamaOption) *OllamaEmbedder {
	o := &OllamaEmbedder{
		url:        defaultOllamaURL,
		model:      defaultOllamaModel,
		dimensions: defaultDimensions,
		client:     &http.Client{},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Embed returns the embedding vector for the given text.
func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]string{
		"model":  o.model,
		"prompt": text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.url+"/api/embeddings", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert float64 to float32 for pgvector compatibility
	embedding := make([]float32, len(result.Embedding))
	for i, v := range result.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// EmbedBatch returns embeddings for multiple texts.
func (o *OllamaEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := o.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// Dimensions returns the dimensionality of the embeddings.
func (o *OllamaEmbedder) Dimensions() int {
	return o.dimensions
}

// Ensure OllamaEmbedder implements Embedder.
var _ Embedder = (*OllamaEmbedder)(nil)
