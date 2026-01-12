package store_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/andreas-lindfalk/bobby/internal/store"
)

// StoreSuite is the integration test suite for the store package.
// It uses testcontainers for PostgreSQL with pgvector.
type StoreSuite struct {
	suite.Suite
	store     *store.Store
	container *postgres.PostgresContainer
	ctx       context.Context
}

// SetupSuite runs once before all tests - starts the container.
func (s *StoreSuite) SetupSuite() {
	s.ctx = context.Background()

	var err error
	s.container, err = postgres.Run(s.ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err, "failed to start postgres container")

	connString, err := s.container.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err, "failed to get connection string")

	s.store, err = store.New(s.ctx, connString)
	s.Require().NoError(err, "failed to create store")

	err = s.store.MigrateWithSeeds(s.ctx)
	s.Require().NoError(err, "failed to run migrations")
}

// TearDownSuite runs once after all tests - stops the container.
func (s *StoreSuite) TearDownSuite() {
	if s.store != nil {
		s.store.Close()
	}
	if s.container != nil {
		if err := s.container.Terminate(s.ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}
}

// deleteDocument removes a document by ID (for test cleanup).
func (s *StoreSuite) deleteDocument(id int) {
	_, err := s.store.Pool().Exec(s.ctx, "DELETE FROM documents WHERE id = $1", id)
	s.Require().NoError(err, "failed to delete document %d", id)
}

// deleteGestoria removes a gestoria by ID (for test cleanup).
func (s *StoreSuite) deleteGestoria(id int) {
	_, err := s.store.Pool().Exec(s.ctx, "DELETE FROM gestorias WHERE id = $1", id)
	s.Require().NoError(err, "failed to delete gestoria %d", id)
}

// insertTestDocument inserts a document and returns its ID for cleanup.
func (s *StoreSuite) insertTestDocument(doc store.Document) int {
	var id int
	err := s.store.Pool().QueryRow(s.ctx, `
		INSERT INTO documents (source, title, content, summary, url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, doc.Source, doc.Title, doc.Content, doc.Summary, doc.URL).Scan(&id)
	s.Require().NoError(err, "failed to insert test document")
	return id
}

// insertTestGestoria inserts a gestoria and returns its ID for cleanup.
func (s *StoreSuite) insertTestGestoria(name, city string, services, languages []string, verified, partner bool, rating float64) int {
	var id int
	err := s.store.Pool().QueryRow(s.ctx, `
		INSERT INTO gestorias (name, city, services, languages, verified, partner, rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, name, city, services, languages, verified, partner, rating).Scan(&id)
	s.Require().NoError(err, "failed to insert test gestoria")
	return id
}

// TestStoreSuite runs the test suite.
func TestStoreSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test suite")
	}
	suite.Run(t, new(StoreSuite))
}

// --- Gestoria Tests ---

func (s *StoreSuite) TestGetGestoriasByCity_ReturnsSeededData() {
	gestorias, err := s.store.GetGestoriasByCity(s.ctx, "Orihuela Costa")
	s.Require().NoError(err)
	s.Require().Len(gestorias, 1, "expected exactly 1 gestoria in Orihuela Costa")

	g := gestorias[0]
	s.Require().Equal("Gestoría López & Partners", g.Name)
	s.Require().True(g.Partner, "expected gestoria to be a partner")
	s.Require().True(g.Verified, "expected gestoria to be verified")
	s.Require().Equal(4.8, g.Rating)
	s.Require().Contains(g.Services, "matriculation")
	s.Require().Contains(g.Languages, "sv")
}

func (s *StoreSuite) TestGetGestoriasByCity_EmptyForUnknownCity() {
	gestorias, err := s.store.GetGestoriasByCity(s.ctx, "NonExistentCity")
	s.Require().NoError(err)
	s.Require().Empty(gestorias, "expected no gestorias for unknown city")
}

func (s *StoreSuite) TestGetVerifiedGestoriaForService_FiltersCorrectly() {
	gestorias, err := s.store.GetVerifiedGestoriaForService(s.ctx, "matriculation", []string{"sv", "en"})
	s.Require().NoError(err)
	s.Require().NotEmpty(gestorias, "expected at least one Swedish-speaking matriculation gestoria")

	// First one should be partner (sorted by partner DESC, rating DESC)
	s.Require().True(gestorias[0].Partner, "expected partner gestoria first")

	// All returned should have matriculation service and speak sv or en
	for _, g := range gestorias {
		s.Require().True(g.Verified, "all returned gestorias should be verified")
		s.Require().Contains(g.Services, "matriculation")
	}
}

func (s *StoreSuite) TestGetVerifiedGestoriaForService_NoResultsForRareLanguage() {
	gestorias, err := s.store.GetVerifiedGestoriaForService(s.ctx, "matriculation", []string{"zh"}) // Chinese
	s.Require().NoError(err)
	s.Require().Empty(gestorias, "expected no Chinese-speaking matriculation gestorias")
}

func (s *StoreSuite) TestGestoriaWithDynamicData_InsertAndCleanup() {
	// Insert test data
	id := s.insertTestGestoria("Test Gestoria", "Test City", []string{"matriculation"}, []string{"sv"}, true, false, 4.0)
	defer s.deleteGestoria(id)

	// Verify it's findable
	gestorias, err := s.store.GetGestoriasByCity(s.ctx, "Test City")
	s.Require().NoError(err)
	s.Require().Len(gestorias, 1)
	s.Require().Equal("Test Gestoria", gestorias[0].Name)
	s.Require().Equal(id, gestorias[0].ID)
}

func (s *StoreSuite) TestGestoriaCleanup_NoLeftoverData() {
	// This test verifies that previous test's defer cleanup worked
	gestorias, err := s.store.GetGestoriasByCity(s.ctx, "Test City")
	s.Require().NoError(err)
	s.Require().Empty(gestorias, "expected no gestorias in Test City - cleanup should have removed them")
}

// --- ITV Station Tests ---

func (s *StoreSuite) TestGetBestITVForImports_ReturnsShortest() {
	itv, err := s.store.GetBestITVForImports(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(itv)

	// Torrevieja has shortest wait (7 days) in seeds
	s.Require().Equal("Torrevieja", itv.City)
	s.Require().Equal(7, itv.AvgWaitDays)
	s.Require().True(itv.HandlesImports)
	s.Require().NotEmpty(itv.AppointmentURL)
}

func (s *StoreSuite) TestGetITVStationsByCity_ReturnsAllOrderedByWait() {
	stations, err := s.store.GetITVStationsByCity(s.ctx, "Torrevieja")
	s.Require().NoError(err)

	// We have 4 ITV stations in seeds, all handle imports
	s.Require().Len(stations, 4)

	// Verify ordering by wait time ascending
	for i := 1; i < len(stations); i++ {
		s.Require().LessOrEqual(stations[i-1].AvgWaitDays, stations[i].AvgWaitDays,
			"expected stations ordered by wait time ascending")
	}

	// All should handle imports
	for _, st := range stations {
		s.Require().True(st.HandlesImports)
	}
}

// --- Document/Vector Tests ---

func (s *StoreSuite) TestDocumentInsertAndSearch_TextMatch() {
	doc := store.Document{
		Source:  "manual",
		Title:   "Car Import Guide for Sweden",
		Content: "When importing a car from Sweden to Spain, you must complete matriculation within 183 days.",
		Summary: "Swedish car import guide",
		URL:     "https://example.com/guide",
	}

	id := s.insertTestDocument(doc)
	defer s.deleteDocument(id)

	// Search by text
	results, err := s.store.SearchByText(s.ctx, "Sweden", 10)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("Car Import Guide for Sweden", results[0].Title)
	s.Require().Equal("manual", results[0].Source)
	s.Require().Contains(results[0].Content, "183 days")
}

func (s *StoreSuite) TestDocumentSearch_MultipleDocuments() {
	docs := []store.Document{
		{Source: "test", Title: "Norwegian Tax Guide", Content: "Tax info for Norway", Summary: "Norway taxes"},
		{Source: "test", Title: "Danish Residency", Content: "Residency rules in Denmark", Summary: "Denmark living"},
		{Source: "test", Title: "Finnish Car Import", Content: "Importing cars from Finland", Summary: "Finland cars"},
	}

	var ids []int
	for _, doc := range docs {
		ids = append(ids, s.insertTestDocument(doc))
	}
	defer func() {
		for _, id := range ids {
			s.deleteDocument(id)
		}
	}()

	// Search for Norway - should find exactly 1
	results, err := s.store.SearchByText(s.ctx, "Norway", 10)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("Norwegian Tax Guide", results[0].Title)

	// Search for Denmark - should find exactly 1
	results, err = s.store.SearchByText(s.ctx, "Denmark", 10)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("Danish Residency", results[0].Title)

	// Search for Finland - should find exactly 1
	results, err = s.store.SearchByText(s.ctx, "Finland", 10)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("Finnish Car Import", results[0].Title)
}

func (s *StoreSuite) TestDocumentSearch_NoResults() {
	results, err := s.store.SearchByText(s.ctx, "NonExistentSearchTerm12345", 10)
	s.Require().NoError(err)
	s.Require().Empty(results)
}

func (s *StoreSuite) TestDocumentCleanup_NoLeftoverData() {
	// Verify previous test's documents were cleaned up
	results, err := s.store.SearchByText(s.ctx, "Norway", 10)
	s.Require().NoError(err)
	s.Require().Empty(results, "expected no Norway documents - cleanup should have removed them")

	results, err = s.store.SearchByText(s.ctx, "Denmark", 10)
	s.Require().NoError(err)
	s.Require().Empty(results, "expected no Denmark documents - cleanup should have removed them")
}
