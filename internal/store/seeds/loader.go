// Package seeds provides document loading for the RAG knowledge base.
package seeds

import (
	"context"
	"fmt"
	"log"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
	"github.com/andreas-lindfalk/bobby/internal/store"
)

// Loader handles seeding documents into the store with embeddings.
type Loader struct {
	store    *store.Store
	embedder embeddings.Embedder
}

// NewLoader creates a new seed loader.
func NewLoader(s *store.Store, e embeddings.Embedder) *Loader {
	return &Loader{
		store:    s,
		embedder: e,
	}
}

// LoadCarImportDocs loads all car import documents into the store.
// It skips documents that already exist (based on title).
func (l *Loader) LoadCarImportDocs(ctx context.Context) error {
	docs := CarImportDocs()
	return l.loadDocs(ctx, docs, "car_import")
}

// LoadAllDocs loads all seed documents into the store.
func (l *Loader) LoadAllDocs(ctx context.Context) error {
	if err := l.LoadCarImportDocs(ctx); err != nil {
		return fmt.Errorf("loading car import docs: %w", err)
	}
	// Add more document types here as they're created
	return nil
}

// loadDocs loads a slice of documents, generating embeddings and skipping duplicates.
func (l *Loader) loadDocs(ctx context.Context, docs []store.Document, category string) error {
	log.Printf("Loading %d %s documents...", len(docs), category)

	inserted := 0
	skipped := 0

	for _, doc := range docs {
		// Check if document already exists
		exists, err := l.docExists(ctx, doc.Title)
		if err != nil {
			return fmt.Errorf("checking doc existence: %w", err)
		}
		if exists {
			log.Printf("  Skipping (exists): %s", doc.Title)
			skipped++
			continue
		}

		// Generate embedding from content
		embedding, err := l.embedder.Embed(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("generating embedding for %q: %w", doc.Title, err)
		}

		// Insert document with embedding
		if _, err := l.store.InsertDocument(ctx, doc, embedding); err != nil {
			return fmt.Errorf("inserting doc %q: %w", doc.Title, err)
		}

		log.Printf("  Inserted: %s", doc.Title)
		inserted++
	}

	log.Printf("Completed: %d inserted, %d skipped", inserted, skipped)
	return nil
}

// docExists checks if a document with the given title already exists.
func (l *Loader) docExists(ctx context.Context, title string) (bool, error) {
	// Use semantic search with the title to find potential matches
	// This is a simple approach - could be improved with exact title lookup
	embedding, err := l.embedder.Embed(ctx, title)
	if err != nil {
		return false, err
	}

	results, err := l.store.SearchByEmbedding(ctx, embedding, 5)
	if err != nil {
		return false, err
	}

	for _, r := range results {
		if r.Title == title {
			return true, nil
		}
	}
	return false, nil
}
