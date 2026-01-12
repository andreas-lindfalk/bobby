// Command bobby starts the AI agent server for car import assistance.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/agent"
	"github.com/andreas-lindfalk/bobby/internal/embeddings"
	"github.com/andreas-lindfalk/bobby/internal/store"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
	// CLI flags
	dbURL := flag.String("db", envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/bobby?sslmode=disable"), "PostgreSQL connection URL")
	ollamaURL := flag.String("ollama", envOrDefault("OLLAMA_URL", "http://localhost:11434"), "Ollama API URL")
	embeddingModel := flag.String("embedding-model", envOrDefault("EMBEDDING_MODEL", "nomic-embed-text"), "Embedding model name")
	llmModel := flag.String("model", envOrDefault("LLM_MODEL", "googleai/gemini-3-flash"), "LLM model name")
	port := flag.String("port", envOrDefault("PORT", "3400"), "Server port")
	flag.Parse()

	ctx := context.Background()

	log.Println("Starting Bobby - Car Import Assistant")
	log.Printf("LLM Model: %s", *llmModel)
	log.Printf("Embedding Model: %s @ %s", *embeddingModel, *ollamaURL)

	// Initialize Genkit with Google AI plugin
	g := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.GoogleAI{}))

	// Initialize store (optional - agent works without it, just no RAG)
	var s *store.Store
	var embedder embeddings.Embedder

	s, err := store.New(ctx, *dbURL)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		log.Println("Agent will run without RAG capabilities")
	} else {
		defer s.Close()

		// Run migrations
		if err := s.Migrate(ctx); err != nil {
			log.Printf("Warning: Migration failed: %v", err)
		}

		// Initialize embedder
		embedder = embeddings.NewOllamaEmbedder(
			embeddings.WithOllamaURL(*ollamaURL),
			embeddings.WithOllamaModel(*embeddingModel),
		)

		// Test embedder connectivity
		testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if _, err := embedder.Embed(testCtx, "test"); err != nil {
			log.Printf("Warning: Embedder not available: %v", err)
			embedder = nil
		}
		cancel()
	}

	// Create agent
	carAgent := agent.New(g, agent.Config{
		Store:       s,
		Embedder:    embedder,
		ModelName:   *llmModel,
		RAGTopK:     5,
		Temperature: 0.7,
	})

	// Define flows
	carAgent.DefineCarImportChatFlow()
	carAgent.DefineSimpleQueryFlow()

	log.Printf("Registered flows: carImportChat, carImportQuery")

	// Set up HTTP server for flows
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Add Genkit flow handlers
	for _, flow := range genkit.ListFlows(g) {
		mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
		log.Printf("Registered endpoint: POST /%s", flow.Name())
	}

	// Start server
	addr := ":" + *port
	log.Printf("Server listening on %s", addr)
	log.Printf("Try: curl -X POST http://localhost%s/carImportQuery -H 'Content-Type: application/json' -d '{\"data\":\"I arrived in Spain 3 months ago with my Volvo. What do I need to do?\"}'", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
