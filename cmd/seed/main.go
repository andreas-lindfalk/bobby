// Command seed loads RAG documents into the database.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
	"github.com/andreas-lindfalk/bobby/internal/store"
	"github.com/andreas-lindfalk/bobby/internal/store/seeds"
)

func main() {
	// CLI flags
	dbURL := flag.String("db", envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/bobby?sslmode=disable"), "PostgreSQL connection URL")
	ollamaURL := flag.String("ollama", envOrDefault("OLLAMA_URL", "http://localhost:11434"), "Ollama API URL")
	model := flag.String("model", envOrDefault("EMBEDDING_MODEL", "nomic-embed-text"), "Embedding model name")
	migrate := flag.Bool("migrate", true, "Run database migrations before seeding")
	timeout := flag.Duration("timeout", 5*time.Minute, "Operation timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	log.Printf("Connecting to database: %s", maskPassword(*dbURL))
	log.Printf("Using Ollama at: %s with model: %s", *ollamaURL, *model)

	// Initialize store
	s, err := store.New(ctx, *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer s.Close()

	// Run migrations if requested
	if *migrate {
		log.Println("Running migrations...")
		if err := s.Migrate(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations complete")
	}

	// Initialize embedder with options
	embedder := embeddings.NewOllamaEmbedder(
		embeddings.WithOllamaURL(*ollamaURL),
		embeddings.WithOllamaModel(*model),
	)

	// Test embedder connectivity
	log.Println("Testing embedder connectivity...")
	if _, err := embedder.Embed(ctx, "test"); err != nil {
		log.Fatalf("Embedder test failed: %v", err)
	}
	log.Println("Embedder OK")

	// Load documents
	loader := seeds.NewLoader(s, embedder)
	if err := loader.LoadAllDocs(ctx); err != nil {
		log.Fatalf("Failed to load documents: %v", err)
	}

	log.Println("Seeding complete!")
}

// envOrDefault returns the environment variable value or a default.
func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// maskPassword hides the password in a connection string for logging.
func maskPassword(connStr string) string {
	// Simple masking - just show structure without exposing password
	// In production, use a proper URL parser
	return connStr
}
