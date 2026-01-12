// Package embeddings provides an interface for text embedding services.
package embeddings

import "context"

// Embedder generates vector embeddings for text.
type Embedder interface {
	// Embed returns the embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch returns embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions returns the dimensionality of the embeddings.
	Dimensions() int
}
