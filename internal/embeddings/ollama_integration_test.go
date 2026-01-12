package embeddings_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
	"github.com/andreas-lindfalk/bobby/internal/testcontainers"
)

type OllamaIntegrationSuite struct {
	suite.Suite
	container *testcontainers.OllamaContainer
	embedder  *embeddings.OllamaEmbedder
	ctx       context.Context
}

func TestOllamaIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Ollama integration test")
	}
	suite.Run(t, new(OllamaIntegrationSuite))
}

func (s *OllamaIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	var err error
	s.container, err = testcontainers.NewOllama(s.ctx)
	s.Require().NoError(err, "failed to start Ollama container")

	err = s.container.PullModel(s.ctx, "nomic-embed-text")
	s.Require().NoError(err, "failed to pull model")

	s.embedder = embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(s.container.URL),
	)
}

func (s *OllamaIntegrationSuite) TearDownSuite() {
	if s.container != nil {
		if err := s.container.Terminate(s.ctx); err != nil {
			s.T().Logf("failed to terminate Ollama container: %v", err)
		}
	}
}

func (s *OllamaIntegrationSuite) TestEmbed_ReturnsValidVector() {
	emb, err := s.embedder.Embed(s.ctx, "Hello, this is a test sentence.")
	s.Require().NoError(err)
	s.Require().Len(emb, 768)

	var sum float32
	for _, v := range emb {
		sum += v * v
	}
	s.Require().Greater(sum, float32(0.1), "embedding should not be all zeros")
}

func (s *OllamaIntegrationSuite) TestEmbed_SimilarTextsHighSimilarity() {
	emb1, err := s.embedder.Embed(s.ctx, "Importing a car from Sweden to Spain")
	s.Require().NoError(err)

	emb2, err := s.embedder.Embed(s.ctx, "Vehicle import from Sweden to Spanish plates")
	s.Require().NoError(err)

	similarity := cosineSimilarity(emb1, emb2)
	s.T().Logf("Similar texts similarity: %.4f", similarity)
	s.Require().Greater(similarity, float32(0.5), "similar texts should have higher similarity than unrelated")
}

func (s *OllamaIntegrationSuite) TestEmbed_DifferentTextsLowSimilarity() {
	emb1, err := s.embedder.Embed(s.ctx, "Car import from Sweden to Spain")
	s.Require().NoError(err)

	emb2, err := s.embedder.Embed(s.ctx, "Recipe for chocolate cake")
	s.Require().NoError(err)

	similarity := cosineSimilarity(emb1, emb2)
	s.T().Logf("Different texts similarity: %.4f", similarity)
	s.Require().Less(similarity, float32(0.5))
}

func (s *OllamaIntegrationSuite) TestEmbedBatch() {
	texts := []string{"Doc about cars", "Doc about taxes", "Doc about Spain"}
	embs, err := s.embedder.EmbedBatch(s.ctx, texts)
	s.Require().NoError(err)
	s.Require().Len(embs, 3)
	for i, emb := range embs {
		s.Require().Len(emb, 768, "embedding %d", i)
	}
}

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt32(normA) * sqrt32(normB))
}

func sqrt32(x float32) float32 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
