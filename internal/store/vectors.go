package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

// Document represents an indexed document for RAG.
type Document struct {
	ID       int
	Source   string
	Title    string
	Content  string
	Summary  string
	URL      string
	Metadata map[string]any
}

// SearchByEmbedding searches documents by vector similarity.
func (s *Store) SearchByEmbedding(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
	// Convert embedding to pgvector format
	vec := pgvector.NewVector(embedding)
	rows, err := s.pool.Query(ctx, `
		SELECT id, source, title, content, COALESCE(summary, ''), COALESCE(url, ''), COALESCE(metadata, '{}')
		FROM documents
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`, vec, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		var metadataJSON []byte
		if err := rows.Scan(&d.ID, &d.Source, &d.Title, &d.Content, &d.Summary, &d.URL, &metadataJSON); err != nil {
			return nil, err
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &d.Metadata)
		}
		docs = append(docs, d)
	}

	return docs, rows.Err()
}

// SearchByText searches documents by text match (for fallback when no embeddings).
func (s *Store) SearchByText(ctx context.Context, query string, limit int) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, source, title, content, COALESCE(summary, ''), COALESCE(url, ''), COALESCE(metadata, '{}')
		FROM documents
		WHERE title ILIKE '%' || $1 || '%' OR content ILIKE '%' || $1 || '%'
		LIMIT $2
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		var metadataJSON []byte
		if err := rows.Scan(&d.ID, &d.Source, &d.Title, &d.Content, &d.Summary, &d.URL, &metadataJSON); err != nil {
			return nil, err
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &d.Metadata)
		}
		docs = append(docs, d)
	}

	return docs, rows.Err()
}

// InsertDocument inserts a new document with optional embedding.
func (s *Store) InsertDocument(ctx context.Context, doc Document, embedding []float32) (int, error) {
	metadataJSON, _ := json.Marshal(doc.Metadata)

	var emb any = nil
	var dims any = nil
	if len(embedding) > 0 {
		emb = pgvector.NewVector(embedding)
		dims = len(embedding)
	}

	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO documents (source, title, content, summary, url, embedding, embedding_dims, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, doc.Source, doc.Title, doc.Content, doc.Summary, doc.URL, emb, dims, metadataJSON).Scan(&id)

	return id, err
}

// Embedder interface for generating embeddings (to avoid circular imports).
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// InsertDocumentWithEmbedding inserts a document and generates its embedding.
func (s *Store) InsertDocumentWithEmbedding(ctx context.Context, doc Document, embedder Embedder) (int, error) {
	// Generate embedding from content (prefer summary if available)
	textToEmbed := doc.Content
	if doc.Summary != "" {
		textToEmbed = doc.Summary + " " + doc.Content
	}

	embedding, err := embedder.Embed(ctx, textToEmbed)
	if err != nil {
		return 0, fmt.Errorf("failed to generate embedding: %w", err)
	}

	return s.InsertDocument(ctx, doc, embedding)
}

// DeleteDocument removes a document by ID.
func (s *Store) DeleteDocument(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM documents WHERE id = $1", id)
	return err
}
