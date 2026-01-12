package store

import (
	"context"
	"encoding/json"
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
	rows, err := s.pool.Query(ctx, `
		SELECT id, source, title, content, COALESCE(summary, ''), COALESCE(url, ''), COALESCE(metadata, '{}')
		FROM documents
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`, embedding, limit)
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
func (s *Store) InsertDocument(ctx context.Context, doc Document, embedding []float32) error {
	metadataJSON, _ := json.Marshal(doc.Metadata)

	var emb any = nil
	if len(embedding) > 0 {
		emb = embedding
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO documents (source, title, content, summary, url, embedding, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, doc.Source, doc.Title, doc.Content, doc.Summary, doc.URL, emb, metadataJSON)

	return err
}
