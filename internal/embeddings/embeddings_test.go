package embeddings_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
)

type EmbeddingsSuite struct {
	suite.Suite
}

func TestEmbeddingsSuite(t *testing.T) {
	suite.Run(t, new(EmbeddingsSuite))
}

func (s *EmbeddingsSuite) TestOllamaEmbedder_Embed() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Require().Equal("/api/embeddings", r.URL.Path)
		s.Require().Equal("POST", r.Method)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		s.Require().NoError(err)
		s.Require().Equal("nomic-embed-text", req["model"])
		s.Require().NotEmpty(req["prompt"])

		mockEmbedding := make([]float64, 768)
		for i := range mockEmbedding {
			mockEmbedding[i] = float64(i) / 768.0
		}

		resp := map[string]interface{}{
			"embedding": mockEmbedding,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	embedder := embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(server.URL),
	)

	ctx := context.Background()
	emb, err := embedder.Embed(ctx, "test text")
	s.Require().NoError(err)
	s.Require().Len(emb, 768)
	s.Require().Equal(768, embedder.Dimensions())
}

func (s *EmbeddingsSuite) TestOllamaEmbedder_EmbedBatch() {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		mockEmbedding := make([]float64, 768)
		resp := map[string]interface{}{"embedding": mockEmbedding}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	embedder := embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(server.URL),
	)

	ctx := context.Background()
	texts := []string{"text1", "text2", "text3"}
	embs, err := embedder.EmbedBatch(ctx, texts)
	s.Require().NoError(err)
	s.Require().Len(embs, 3)
	s.Require().Equal(3, callCount, "should call API once per text")
}

func (s *EmbeddingsSuite) TestOllamaEmbedder_CustomModel() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		s.Require().Equal("mxbai-embed-large", req["model"])

		mockEmbedding := make([]float64, 1024)
		resp := map[string]interface{}{"embedding": mockEmbedding}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	embedder := embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(server.URL),
		embeddings.WithOllamaModel("mxbai-embed-large"),
		embeddings.WithDimensions(1024),
	)

	s.Require().Equal(1024, embedder.Dimensions())

	ctx := context.Background()
	emb, err := embedder.Embed(ctx, "test")
	s.Require().NoError(err)
	s.Require().Len(emb, 1024)
}

func (s *EmbeddingsSuite) TestEmbedderInterface() {
	// Verify OllamaEmbedder implements Embedder interface
	var _ embeddings.Embedder = embeddings.NewOllamaEmbedder()
}
