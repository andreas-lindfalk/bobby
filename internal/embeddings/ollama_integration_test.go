package embeddings_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
)

type OllamaIntegrationSuite struct {
	suite.Suite
	container testcontainers.Container
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

	req := testcontainers.ContainerRequest{
		Image:        "ollama/ollama:latest",
		ExposedPorts: []string{"11434/tcp"},
		WaitingFor: wait.ForHTTP("/").
			WithPort("11434").
			WithStartupTimeout(60 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericVolumeMountSource{
					Name: "bobby-ollama-models",
				},
				Target: "/root/.ollama",
			},
		},
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err, "failed to start Ollama container")
	s.container = container

	host, err := container.Host(s.ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(s.ctx, "11434")
	s.Require().NoError(err)

	ollamaURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	s.pullModel(ollamaURL, "nomic-embed-text")

	s.embedder = embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(ollamaURL),
	)
}

func (s *OllamaIntegrationSuite) TearDownSuite() {
	if s.container != nil {
		if err := s.container.Terminate(s.ctx); err != nil {
			s.T().Logf("failed to terminate Ollama container: %v", err)
		}
	}
}

func (s *OllamaIntegrationSuite) pullModel(ollamaURL, model string) {
	s.T().Logf("Checking model %s...", model)

	listResp, err := http.Get(ollamaURL + "/api/tags")
	s.Require().NoError(err)
	defer listResp.Body.Close()

	body, _ := io.ReadAll(listResp.Body)
	if strings.Contains(string(body), model) {
		s.T().Logf("Model %s already cached", model)
		return
	}

	s.T().Logf("Pulling model %s (first run only)...", model)
	pullReq := strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, model))
	resp, err := http.Post(ollamaURL+"/api/pull", "application/json", pullReq)
	s.Require().NoError(err)
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 && strings.Contains(string(buf[:n]), "success") {
			s.T().Logf("Model %s pulled successfully", model)
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Require().NoError(err)
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
