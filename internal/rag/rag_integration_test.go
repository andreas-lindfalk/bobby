package rag_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
	"github.com/andreas-lindfalk/bobby/internal/flows"
	"github.com/andreas-lindfalk/bobby/internal/store"
	"github.com/andreas-lindfalk/bobby/internal/testcontainers"
)

// RAGIntegrationSuite tests the full RAG pipeline with real embeddings and pgvector.
type RAGIntegrationSuite struct {
	suite.Suite
	pgContainer     *testcontainers.PostgresContainer
	ollamaContainer *testcontainers.OllamaContainer
	store           *store.Store
	embedder        *embeddings.OllamaEmbedder
	ctx             context.Context
}

func TestRAGIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping RAG integration test")
	}
	suite.Run(t, new(RAGIntegrationSuite))
}

func (s *RAGIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	// Start Postgres with pgvector
	var err error
	s.pgContainer, err = testcontainers.NewPostgres(s.ctx)
	s.Require().NoError(err, "failed to start postgres container")

	s.store, err = store.New(s.ctx, s.pgContainer.ConnStr)
	s.Require().NoError(err, "failed to create store")

	err = s.store.MigrateWithSeeds(s.ctx)
	s.Require().NoError(err, "failed to run migrations")

	// Start Ollama for embeddings
	s.ollamaContainer, err = testcontainers.NewOllama(s.ctx)
	s.Require().NoError(err, "failed to start ollama container")

	err = s.ollamaContainer.PullModel(s.ctx, "nomic-embed-text")
	s.Require().NoError(err, "failed to pull embedding model")

	s.embedder = embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(s.ollamaContainer.URL),
	)
}

func (s *RAGIntegrationSuite) TearDownSuite() {
	if s.store != nil {
		s.store.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(s.ctx)
	}
	if s.ollamaContainer != nil {
		_ = s.ollamaContainer.Terminate(s.ctx)
	}
}

func (s *RAGIntegrationSuite) TestFullRAGPipeline() {
	// 1. Insert documents with embeddings
	docs := []store.Document{
		{
			Source:  "manual",
			Title:   "Swedish Car Import Guide",
			Content: "When importing a car from Sweden to Spain, you must register it within 183 days of becoming a tax resident. The process involves getting an ITV inspection, paying registration tax, and obtaining Spanish plates.",
			Summary: "Guide for importing Swedish vehicles to Spain",
		},
		{
			Source:  "manual",
			Title:   "RUT Tax Deduction Rules",
			Content: "Swedish residents can claim a RUT deduction for household services. The maximum deduction is 75,000 SEK per year. Services must be performed by a Swedish company.",
			Summary: "Swedish RUT tax deduction guide",
		},
		{
			Source:  "manual",
			Title:   "Best Beaches in Costa Blanca",
			Content: "The Costa Blanca region offers beautiful beaches including La Zenia, Playa Flamenca, and Guardamar. Perfect for swimming and relaxation.",
			Summary: "Beach guide for Costa Blanca",
		},
	}

	var ids []int
	for _, doc := range docs {
		id, err := s.store.InsertDocumentWithEmbedding(s.ctx, doc, s.embedder)
		s.Require().NoError(err, "failed to insert document: %s", doc.Title)
		s.T().Logf("Inserted %q (ID: %d)", doc.Title, id)
		ids = append(ids, id)
	}
	defer func() {
		for _, id := range ids {
			_ = s.store.DeleteDocument(s.ctx, id)
		}
	}()

	// 2. Verify documents with embeddings
	var count int
	err := s.store.Pool().QueryRow(s.ctx, "SELECT COUNT(*) FROM documents WHERE embedding IS NOT NULL").Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(3, count, "expected 3 documents with embeddings")

	// 3. Test semantic search - find car import guide
	queryEmb, err := s.embedder.Embed(s.ctx, "How do I register my Swedish car in Spain?")
	s.Require().NoError(err)

	results, err := s.store.SearchByEmbedding(s.ctx, queryEmb, 3)
	s.Require().NoError(err)
	s.T().Logf("Car query returned %d results", len(results))
	for i, r := range results {
		s.T().Logf("  %d: %s", i, r.Title)
	}
	s.Require().NotEmpty(results, "expected results for car query")
	s.Require().Equal("Swedish Car Import Guide", results[0].Title)
}

func (s *RAGIntegrationSuite) TestSemanticVsTextSearch() {
	// Insert a document with specific terminology
	doc := store.Document{
		Source:  "dgt",
		Title:   "Matriculation Requirements DGT",
		Content: "Vehicle registration (matriculación) in Spain requires: 1) Technical inspection (ITV), 2) Registration tax payment, 3) Municipal tax. Foreign vehicles must complete homologation.",
		Summary: "Official DGT requirements for vehicle matriculation",
	}

	id, err := s.store.InsertDocumentWithEmbedding(s.ctx, doc, s.embedder)
	s.Require().NoError(err)
	s.T().Logf("Inserted %q (ID: %d)", doc.Title, id)
	defer s.store.DeleteDocument(s.ctx, id)

	// Text search for "car registration" won't find "vehicle matriculation"
	textResults, err := s.store.SearchByText(s.ctx, "car registration", 5)
	s.Require().NoError(err)
	s.Require().Empty(textResults, "text search for 'car registration' should not match 'vehicle matriculation'")

	// But semantic search WILL find it
	queryEmb, err := s.embedder.Embed(s.ctx, "car registration")
	s.Require().NoError(err)

	embResults, err := s.store.SearchByEmbedding(s.ctx, queryEmb, 5)
	s.Require().NoError(err)
	s.T().Logf("Semantic search returned %d results", len(embResults))
	for i, r := range embResults {
		s.T().Logf("  %d: %s", i, r.Title)
	}
	s.Require().NotEmpty(embResults, "semantic search should find 'vehicle matriculation' when querying 'car registration'")
	s.Require().Equal("Matriculation Requirements DGT", embResults[0].Title)
}

func (s *RAGIntegrationSuite) TestFlowWithRAGContext() {
	// Insert documents that the flow should find
	docs := []store.Document{
		{
			Source:  "dgt",
			Title:   "Swedish Vehicle Import Requirements",
			Content: "Swedish vehicles imported to Spain must complete ITV homologation within 30 days of padrón registration.",
			Summary: "DGT requirements for Swedish car imports",
		},
		{
			Source:  "boe",
			Title:   "Electric Vehicle Tax Exemption 2024",
			Content: "Electric vehicles (BEV) are exempt from registration tax (Impuesto de Matriculación) under Spanish law.",
			Summary: "Tax exemption for electric cars in Spain",
		},
	}

	var ids []int
	for _, doc := range docs {
		id, err := s.store.InsertDocumentWithEmbedding(s.ctx, doc, s.embedder)
		s.Require().NoError(err)
		ids = append(ids, id)
	}
	defer func() {
		for _, id := range ids {
			_ = s.store.DeleteDocument(s.ctx, id)
		}
	}()

	// Run flow with RAG dependencies
	input := flows.VehicleImportInput{
		ArrivalDate:  time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		CarValue:     25000,
		CO2Emissions: 140,
	}
	deps := &flows.FlowDependencies{
		Store:    s.store,
		Embedder: s.embedder,
	}

	output, err := flows.RunVehicleImportFlow(s.ctx, input, deps)
	s.Require().NoError(err)

	// Verify RAG context was populated
	s.T().Logf("RAGContext has %d documents", len(output.RAGContext))
	for i, doc := range output.RAGContext {
		s.T().Logf("  %d: %s (%s)", i, doc.Title, doc.Source)
	}
	s.Require().NotEmpty(output.RAGContext, "flow should retrieve RAG context")

	// First doc should be the Swedish import guide (most relevant to the query)
	s.Require().Equal("Swedish Vehicle Import Requirements", output.RAGContext[0].Title)
}
