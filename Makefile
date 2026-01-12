.PHONY: dev up down seed run build test clean pull-model

# Start dependencies (Postgres + Ollama)
up:
	docker compose up -d

# Stop dependencies
down:
	docker compose down

# Pull embedding model (run once after 'up')
pull-model:
	docker exec bobby-ollama ollama pull nomic-embed-text

# Seed the database with RAG documents
seed: 
	go run ./cmd/seed -migrate

# Build the server
build:
	go build -o bobby ./cmd/bobby

# Run the server
run: build
	./bobby

# Full dev setup: start deps, pull model, seed, run
dev: up
	@echo "Waiting for services to be ready..."
	@sleep 5
	@docker exec bobby-ollama ollama pull nomic-embed-text || true
	@go run ./cmd/seed -migrate || true
	@./bobby

# Run tests
test:
	go test ./... -count=1 -p=1 -short

# Run tests with integration tests
test-all:
	go test ./... -count=1 -p=1

# Clean up
clean:
	rm -f bobby
	docker compose down -v

# Quick chat test
chat:
	@curl -s -X POST http://localhost:3400/chat \
		-H 'Content-Type: application/json' \
		-d '{"data":{"session_id":"dev","message":"Hej! Jag kom till Spanien för 3 månader sedan med min Volvo V60."}}' | jq -r '.result.response'
